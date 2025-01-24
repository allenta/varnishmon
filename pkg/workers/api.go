package workers

import (
	"context"
	"fmt"
	"sync"

	"allenta.com/varnishmon/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/tcplisten"
)

type APIWorker struct {
	*worker
	handler APIHandler
	server  *fasthttp.Server
}

type APIHandler interface {
	HandleRequest(ctx context.Context, rctx *fasthttp.RequestCtx)
}

func NewAPIWorker(
	ctx context.Context, wg *sync.WaitGroup, app Application,
	id int, handler APIHandler) *APIWorker {
	aw := &APIWorker{
		handler: handler,
	}

	aw.worker = &worker{
		ctx:  ctx,
		wg:   wg,
		app:  app,
		id:   fmt.Sprintf("API #%d", id),
		init: aw.init,
		run:  aw.run,
		stop: aw.stop,
	}

	return aw
}

func (aw *APIWorker) init() {
	aw.server = &fasthttp.Server{
		Name:    fmt.Sprintf("varnishmon/%s (%s)", config.Version(), config.Revision()),
		Handler: fasthttp.CompressHandler(aw.handleRequest),
		Logger:  aw.app.Cfg().Log(),

		Concurrency: aw.app.Cfg().APIConcurrency(),

		ReadBufferSize:     aw.app.Cfg().APIReadBufferSize(),
		WriteBufferSize:    aw.app.Cfg().APIWriteBufferSize(),
		MaxRequestBodySize: aw.app.Cfg().APIMaxRequestBodySize(),

		ReadTimeout:  aw.app.Cfg().APIReadTimeout(),
		WriteTimeout: aw.app.Cfg().APIWriteTimeout(),
		IdleTimeout:  aw.app.Cfg().APIIdleTimeout(),

		TCPKeepalive:       aw.app.Cfg().APITCPKeepalive(),
		TCPKeepalivePeriod: aw.app.Cfg().APITCPKeepalivePeriod(),
	}

	aw.app.Cfg().Metrics().Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "api_worker_concurrency",
			Help:        "Served connections, partitioned by API worker",
			ConstLabels: prometheus.Labels{"id": aw.worker.id},
		},
		func() float64 {
			return float64(aw.server.GetCurrentConcurrency())
		},
	))

	aw.app.Cfg().Metrics().Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        "api_worker_open_connections",
			Help:        "Opened connections, partitioned by API worker",
			ConstLabels: prometheus.Labels{"id": aw.worker.id},
		},
		func() float64 {
			return float64(aw.server.GetOpenConnectionsCount())
		},
	))
}

func (aw *APIWorker) handleRequest(rctx *fasthttp.RequestCtx) {
	aw.handler.HandleRequest(aw.ctx, rctx)
}

func (aw *APIWorker) run() {
	address := fmt.Sprintf(
		"%s:%d",
		aw.app.Cfg().APIListenIP(),
		aw.app.Cfg().APIListenPort())

	var cfg = &tcplisten.Config{
		ReusePort:   true,
		DeferAccept: true,
		FastOpen:    true,
		Backlog:     aw.app.Cfg().APIBacklog(),
	}

	if listener, err := cfg.NewListener("tcp4", address); err == nil {
		if aw.app.Cfg().APITLSCertfile() != "" && aw.app.Cfg().APITLSKeyfile() != "" {
			if err := aw.server.ServeTLS(listener, aw.app.Cfg().APITLSCertfile(), aw.app.Cfg().APITLSKeyfile()); err != nil {
				aw.app.Cfg().Log().Fatal().
					Err(err).
					Str("address", address).
					Msgf("Failed to launch %v!", aw)
			}
		} else {
			if err := aw.server.Serve(listener); err != nil {
				aw.app.Cfg().Log().Fatal().
					Err(err).
					Str("address", address).
					Msgf("Failed to launch %v!", aw)
			}
		}
	} else {
		aw.app.Cfg().Log().Fatal().
			Err(err).
			Str("address", address).
			Msgf("Failed to create TCP listener for '%v' worker!", aw)
	}
}

func (aw *APIWorker) stop() {
	// 'aw.server.Shutdown()' blocks indefinitely while in-flight connections
	// are gracefully closed. Even worst, if 'APIReadTimeout()' or
	// 'APIIdleTimeout()' are set to 0 (i.e. no timeout) it will block forever.
	// Ideally we'd prefer a controlled but ungraceful shutdown in order to
	// avoid blocking the service shutdown too much time. Apparently that's not
	// possible for the fasthttp library. Best alternative is triggering
	// 'aw.server.Shutdown()' asynchronously: 'aw.server.Serve()' will exit
	// immediately and nobody will wait for goroutines handling in-flight
	// connections.
	go func() {
		if err := aw.server.Shutdown(); err != nil {
			aw.app.Cfg().Log().Error().
				Err(err).
				Msgf("Got error while shutting down '%v' worker!", aw)
		}
	}()
}
