package workers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type APIWorker struct {
	*worker

	handler APIHandler
	server  *http.Server
}

type APIHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Shutdown() error
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
	aw.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", aw.app.Cfg().APIListenIP(), aw.app.Cfg().APIListenPort()),
		Handler:        aw.handler,
		ReadTimeout:    aw.app.Cfg().APIReadTimeout(),
		WriteTimeout:   aw.app.Cfg().APIWriteTimeout(),
		IdleTimeout:    aw.app.Cfg().APIIdleTimeout(),
		MaxHeaderBytes: aw.app.Cfg().APIMaxHeaderBytes(),
		BaseContext: func(net.Listener) context.Context {
			return aw.ctx
		},
		ErrorLog: log.New(aw.app.Cfg().Log().ErrorWriter(), "", 0),
	}
}

func (aw *APIWorker) run() {
	listenConfig := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil { //nolint:gosec
					aw.app.Cfg().Log().Error().
						Err(err).
						Str("network", network).
						Str("address", address).
						Msgf("Failed to set 'SO_REUSEPORT' socket option for API worker '%v'!", aw)
				}
			})
		},
	}

	listener, err := listenConfig.Listen(aw.ctx, "tcp", aw.server.Addr)
	if err != nil {
		aw.app.Cfg().Log().Fatal().
			Err(err).
			Msgf("API worker '%v' failed to listen!", aw)
	}

	if aw.app.Cfg().APITLSCertfile() != "" && aw.app.Cfg().APITLSKeyfile() != "" {
		if err := aw.server.ServeTLS(listener, aw.app.Cfg().APITLSCertfile(), aw.app.Cfg().APITLSKeyfile()); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			aw.app.Cfg().Log().Fatal().
				Err(err).
				Str("address", aw.server.Addr).
				Msgf("API worker '%v' failed to serve!", aw)
		}
	} else {
		if err := aw.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			aw.app.Cfg().Log().Fatal().
				Err(err).
				Str("address", aw.server.Addr).
				Msgf("API worker '%v' failed to serve!", aw)
		}
	}
}

func (aw *APIWorker) stop() {
	// 'aw.server.Shutdown()' performs an orderly shutdown of the server, waiting
	// for in-flight connections to close before returning. We limit that time
	// to 1 second here. Also, note that incoming requests use the base context
	// 'aw.worker.ctx', which is canceled when the worker is stopped.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := aw.server.Shutdown(ctx); err != nil {
		aw.app.Cfg().Log().Error().
			Err(err).
			Msgf("Got error while shutting down '%v' worker!", aw)
	}
}
