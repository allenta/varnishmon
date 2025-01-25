package workers

import (
	"context"
	"errors"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"allenta.com/varnishmon/pkg/helpers"
	"github.com/prometheus/client_golang/prometheus"
)

type ScraperWorker struct {
	*worker
	wg           sync.WaitGroup
	metricsQueue chan *helpers.VarnishMetrics

	executionCompleted prometheus.Counter
	executionFailed    prometheus.Counter
	queuingFailed      prometheus.Counter
}

func NewScraperWorker(
	ctx context.Context, wg *sync.WaitGroup, app Application,
	metricsQueue chan *helpers.VarnishMetrics) *ScraperWorker {
	sw := &ScraperWorker{
		metricsQueue: metricsQueue,

		executionCompleted: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "scrapper_execution_completed_total",
				Help: "Successful 'varnishstat' executions by the scraper worker",
			}),
		executionFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "scrapper_execution_failed_total",
				Help: "Failed 'varnishstat' executions by the scraper worker",
			}),
		queuingFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "scrapper_queuing_failed_total",
				Help: "Failed attempts to queue metrics by the scraper worker",
			}),
	}

	sw.worker = &worker{
		ctx:  ctx,
		wg:   wg,
		app:  app,
		id:   "Scraper",
		init: sw.init,
		run:  sw.run,
		stop: sw.stop,
	}

	sw.app.Cfg().Metrics().Registry.MustRegister(sw.executionCompleted)
	sw.app.Cfg().Metrics().Registry.MustRegister(sw.executionFailed)
	sw.app.Cfg().Metrics().Registry.MustRegister(sw.queuingFailed)

	return sw
}

func (sw *ScraperWorker) init() {
}

func (sw *ScraperWorker) run() {
	// Do an initial scrape.
	sw.scrape()

	// Start a ticker to go on scraping periodically.
	ticker := time.NewTicker(sw.worker.app.Cfg().ScraperPeriod())
	defer ticker.Stop()

	for {
		select {
		case <-sw.worker.ctx.Done():
			sw.wg.Wait() // Wait for all goroutines to finish.
			return
		case <-ticker.C:
			sw.scrape()
		}
	}
}

func (sw *ScraperWorker) scrape() {
	sw.wg.Add(1)
	go func() {
		defer sw.wg.Done()

		// Create a new context with a timeout to limit 'varnishstat'
		// time execution.
		contextWithTimeout, cancel := context.WithTimeout(
			sw.worker.ctx, sw.worker.app.Cfg().ScraperPeriod())
		defer cancel()

		//nolint:lll
		// Execute 'varnishstat' command. See:
		//   - https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773.
		//   - https://stackoverflow.com/questions/67750520/golang-context-withtimeout-doesnt-work-with-exec-commandcontext-su-c-command.
		//   - https://hackernoon.com/everything-you-need-to-know-about-managing-go-processes#h-enhanced-cancellation-with-wait-delay-and-cancel
		cmd := exec.CommandContext( //nolint:gosec
			contextWithTimeout,
			sw.worker.app.Cfg().ScraperCommand()[0],
			sw.worker.app.Cfg().ScraperCommand()[1:]...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Cancel = func() error {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}

		// Parse the output and send the metrics to the queue.
		if out, err := cmd.CombinedOutput(); err == nil {
			if metrics, err := helpers.ParseVarnishMetrics(out); err == nil {
				sw.executionCompleted.Inc()
				sw.worker.app.Cfg().Log().Debug().
					Interface("metrics", metrics).
					Msg("Successfully fetched 'varnishstat' output")

				// Avoid blocking indefinitely if the metrics queue is full.
				// This is unlikely, but if insertions into the storage are slow,
				// the queue may fill up, causing a backlog of goroutines
				// waiting to insert metrics.
				select {
				case sw.metricsQueue <- metrics:
				case <-sw.worker.ctx.Done():
				default:
					sw.queuingFailed.Inc()
					sw.worker.app.Cfg().Log().Error().
						Msg("Metrics queue is full, dropping metrics!")
				}
			} else {
				sw.executionFailed.Inc()
				sw.worker.app.Cfg().Log().Error().
					Err(err).
					Str("output", string(out)).
					Msg("Failed to parse 'varnishstat' output!")
			}
		} else {
			sw.executionFailed.Inc()

			// Check the error type to log the appropriate message.
			// Ideally, we'd check the returned err as in
			// '!errors.Is(err, context.Canceled)' but this doesn't
			// seem to work as expected, so we're checking the
			// context error. This works but may suffer from race
			// conditions.
			if errors.Is(contextWithTimeout.Err(), context.DeadlineExceeded) {
				sw.worker.app.Cfg().Log().Error().
					Dur("timeout", sw.worker.app.Cfg().ScraperPeriod()).
					Msg("'varnishstat' execution timed out!")
			} else if errors.Is(contextWithTimeout.Err(), context.Canceled) {
				sw.worker.app.Cfg().Log().Error().
					Err(err).
					Str("output", string(out)).
					Msg("Failed to execute 'varnishstat'!")
			}
		}
	}()
}

func (sw *ScraperWorker) stop() {
}
