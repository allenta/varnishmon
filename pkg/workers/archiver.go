package workers

import (
	"context"
	"sync"
	"time"

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
				Help: "Successful pushes of batches of samples by the archiver worker",
			}),
		pushFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "archiver_push_failed_total",
				Help: "Failed pushes of batches of samples by the archiver worker",
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
			batch := make([]*storage.MetricSample, 0, len(metrics.Items))

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

					// Skip if this is the first time seeing the metric:
					// counters are stored as rates calculated based on the
					// previously seen value.
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

					// Transform the 'uint64' value of the metric by dropping
					// the highest bit. See: https://github.com/golang/go/issues/6113.
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

				// Append the metric sample to the batch.
				batch = append(batch, &storage.MetricSample{
					Name:        name,
					Flag:        details.Flag,
					Format:      details.Format,
					Description: details.Description,
					Value:       value,
				})
			}

			if err := aw.storage.PushMetricSamples(metrics.Timestamp, batch); err != nil {
				aw.pushFailed.Inc()
				aw.app.Cfg().Log().Error().
					Err(err).
					Time("timestamp", metrics.Timestamp).
					Int("count", len(batch)).
					Msg("Failed to store batch of samples!")
			} else {
				aw.pushCompleted.Inc()
			}
		}
	}
}

func (aw *ArchiverWorker) stop() {
}
