package workers

import (
	"context"
	"sync"

	"allenta.com/varnishmon/pkg/helpers"
	"allenta.com/varnishmon/pkg/workers/api"
	"allenta.com/varnishmon/pkg/workers/storage"
	"github.com/prometheus/client_golang/prometheus"
)

type Manager struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         *sync.WaitGroup
	app        Application

	metricsQueue chan *helpers.VarnishMetrics

	storage *storage.Storage
}

func NewManager(app Application) *Manager {
	m := &Manager{
		wg:           &sync.WaitGroup{},
		app:          app,
		metricsQueue: make(chan *helpers.VarnishMetrics, 1024),
	}

	m.ctx, m.cancelFunc = context.WithCancel(context.Background())

	app.Cfg().Metrics().Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "metrics_queue",
			Help: "Items in the metrics queue",
		},
		func() float64 {
			return float64(len(m.metricsQueue))
		},
	))

	return m
}

func (m *Manager) Start() {
	m.storage = storage.NewStorage(m.app)

	if m.app.Cfg().ScraperEnabled() {
		NewScraperWorker(m.ctx, m.wg, m.app, m.metricsQueue).Start()
		NewArchiverWorker(m.ctx, m.wg, m.app, m.metricsQueue, m.storage).Start()
	}

	if m.app.Cfg().APIEnabled() {
		apiHandler := api.NewHandler(m.app, m.storage)
		for i := range m.app.Cfg().APIWorkers() {
			NewAPIWorker(m.ctx, m.wg, m.app, i, apiHandler).Start()
		}
	}
}

func (m *Manager) Stop() {
	// Request workers to stop and wait for their termination.
	m.cancelFunc()
	m.wg.Wait()

	// Shutdown storage, assuming this is the last blocking operation just
	// before termination.
	if err := m.storage.Shutdown(); err != nil {
		m.app.Cfg().Log().Error().Err(err).
			Msg("Failed to shutdown storage!")
	}

	// We intentionally discard pending messages in 'm.metricsQueue'.
	pending := len(m.metricsQueue)
	if pending > 0 {
		m.app.Cfg().Log().Warn().
			Msgf("%d messages in metrics queue dropped during shutdown!", pending)
	}
}
