package workers

import (
	"context"
	"errors"
	"sync"
	"time"

	duckdb "github.com/marcboeker/go-duckdb"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/stone.code/assert"

	"github.com/allenta/varnishmon/pkg/helpers"
	"github.com/allenta/varnishmon/pkg/workers/storage"
)

type ArchiverWorker struct {
	*worker
	metricsQueue chan *helpers.VarnishMetrics
	lastMetrics  map[string]*lastMetrics
	storage      *storage.Storage

	outOfOrderSamples prometheus.Counter
	resetCounters     prometheus.Counter
	truncatedSamples  prometheus.Counter
	pushCompleted     prometheus.Counter
	pushFailed        prometheus.Counter
}

type lastMetrics struct {
	timestamp time.Time
	value     uint64
}

func NewArchiverWorker(
	ctx context.Context, wg *sync.WaitGroup, app Application,
	metricsQueue chan *helpers.VarnishMetrics,
	storage *storage.Storage) *ArchiverWorker {
	aw := &ArchiverWorker{
		metricsQueue: metricsQueue,
		lastMetrics:  make(map[string]*lastMetrics),
		storage:      storage,

		outOfOrderSamples: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "archiver_out_of_order_samples_total",
				Help: "Out-of-order samples received by the archiver worker",
			}),
		resetCounters: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "archiver_reset_counters_total",
				Help: "Counter resets detected by the archiver worker",
			}),
		truncatedSamples: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "archiver_truncated_samples_total",
				Help: "Samples (bitmaps excluded) truncated by the archiver worker",
			}),
		pushCompleted: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "archiver_push_completed_total",
				Help: "Successful pushes of metrics by the archiver worker",
			}),
		pushFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "archiver_push_failed_total",
				Help: "Failed pushes of metrics by the archiver worker",
			}),
	}

	aw.worker = &worker{
		ctx:  ctx,
		wg:   wg,
		app:  app,
		id:   "Archiver",
		init: aw.init,
		run:  aw.run,
		stop: aw.stop,
	}

	aw.app.Cfg().Metrics().Registry.MustRegister(aw.outOfOrderSamples)
	aw.app.Cfg().Metrics().Registry.MustRegister(aw.resetCounters)
	aw.app.Cfg().Metrics().Registry.MustRegister(aw.truncatedSamples)
	aw.app.Cfg().Metrics().Registry.MustRegister(aw.pushCompleted)
	aw.app.Cfg().Metrics().Registry.MustRegister(aw.pushFailed)

	return aw
}

func (aw *ArchiverWorker) init() {
}

func (aw *ArchiverWorker) run() {
	for {
		select {
		case <-aw.ctx.Done():
			return
		case metrics := <-aw.metricsQueue:
			for name, details := range metrics.Items {
				// Check if this is the first time seeing the metric.
				previousMetric, ok := aw.lastMetrics[name]
				if !ok {
					aw.lastMetrics[name] = &lastMetrics{
						timestamp: metrics.Timestamp,
						value:     details.Value,
					}
				}

				var value any
				if details.IsCounter() && !details.HasDurationFormat() {
					// Counters are stored as rates, so we need to calculate the
					// rate based on the previously seen value. However, there
					// is a special case for uptimes (i.e., 'd' format in the
					// 'varnishstat' output) which are handled as gauges.
					// Otherwise, the rate per second of an uptime would be
					// pretty much useless.

					// Skip if this is the first time seeing the metric: counters
					// are stored as rates calculated based on the previously seen
					// value.
					if previousMetric == nil {
						continue
					}

					// Skip if this is an out-of-order sample.
					if !metrics.Timestamp.After(previousMetric.timestamp) {
						aw.outOfOrderSamples.Inc()
						continue
					}

					// Skip if this looks like a reset of the counter.
					if details.Value < previousMetric.value {
						previousMetric.timestamp = metrics.Timestamp
						previousMetric.value = details.Value
						aw.resetCounters.Inc()
						continue
					}

					// Transform the 'uint64' value of the counter into a
					// 'float64' rate per second.
					value = float64(details.Value-previousMetric.value) /
						metrics.Timestamp.Sub(previousMetric.timestamp).Seconds()
				} else {
					// Skip if this is an out-of-order sample. Not strictly
					// necessary here, but it is nice to keep things consistent.
					if previousMetric != nil &&
						!metrics.Timestamp.After(previousMetric.timestamp) {
						aw.outOfOrderSamples.Inc()
						continue
					}

					// Transform the 'uint64' value of the metric by dropping the
					// highest bit. See: https://github.com/golang/go/issues/6113.
					if !details.IsBitmap() && details.Value&0x8000000000000000 != 0 {
						aw.truncatedSamples.Inc()
					}
					value = details.Value & 0x7FFFFFFFFFFFFFFF
				}

				// At this point a value should have been set.
				assert.Assert(value != nil, "invalid value")

				// Update the last seen value of the metric.
				if previousMetric != nil {
					previousMetric.timestamp = metrics.Timestamp
					previousMetric.value = details.Value
				}

				// Store the metric.
				if err := aw.storage.PushMetricSample(
					name, metrics.Timestamp, details.Flag, details.Format,
					details.Description, value); err != nil {
					aw.pushFailed.Inc()

					aw.app.Cfg().Log().Error().
						Err(err).
						Str("name", name).
						Str("flag", details.Flag).
						Str("format", details.Format).
						Interface("value", value).
						Msg("Failed to store metric!")

					// On DuckDB errors, discard remaining metrics. Typically,
					// when DuckDB fails to store a metric, it indicates a
					// permanent issue (e.g., memory allocation failure). It is
					// better to stop early to avoid flooding the logs with one
					// error entry for each metric in the batch an to prevent
					// further damage like a CPU spike.
					var duckdbErr *duckdb.Error
					if errors.As(err, &duckdbErr) {
						aw.app.Cfg().Log().Error().
							Interface("type", duckdbErr.Type).
							Interface("msg", duckdbErr.Msg).
							Msg("Hitting a DuckDB error, stopping further processing of current batch of metrics!")
						break
					}
				} else {
					aw.pushCompleted.Inc()
				}
			}
		}
	}
}

func (aw *ArchiverWorker) stop() {
}
