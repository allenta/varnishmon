package workers

import (
	"context"
	"sync"
)

type worker struct {
	ctx context.Context
	wg  *sync.WaitGroup
	app Application
	id  string
	// Blocking initialization callback executed before any of the worker
	// goroutines are launched. This logic is guaranteed to be fully executed
	// before the core logic (i.e. '.run()') or the request for finalization
	// logic (i.e. '.stop()') of the worker are launched.
	init func()
	// Blocking callback embedding the core logic of the worker. That logic is
	// expected to be executed indefinitely while context is not cancelled.
	// Finalization logic is expected to be executed here too, just before
	// leaving this callback.
	run func()
	// Blocking callback *exclusively* used to request finalization of the core
	// logic (i.e. '.run()') when execution of blocking operations do not allow
	// the core logic to use the context. Finalization logic is not expected to
	// be executed here in order to avoid concurrency issues.
	stop func()
}

func (wrk *worker) Start() {
	wrk.app.Cfg().Log().Info().Msgf("Starting '%v' worker", wrk)

	wrk.init()

	wrk.wg.Add(1)
	go func() {
		defer wrk.wg.Done()

		wrk.run()

		if wrk.ctx.Err() == nil {
			wrk.app.Cfg().Log().Warn().Msgf("Worker '%v' terminated unexpectedly", wrk)
		}
	}()

	wrk.wg.Add(1)
	go func() {
		defer wrk.wg.Done()

		<-wrk.ctx.Done()

		wrk.app.Cfg().Log().Info().Msgf("Stopping '%v' worker", wrk)

		wrk.stop()
	}()
}

func (wrk *worker) String() string {
	return wrk.id
}
