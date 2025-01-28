package application

import (
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/allenta/varnishmon/pkg/config"
	"github.com/allenta/varnishmon/pkg/helpers"
	"github.com/allenta/varnishmon/pkg/workers"
)

type Application struct {
	cfg      *config.Config
	startTst int64
	manager  *workers.Manager
}

func NewApplication(cfg *config.Config) *Application {
	app := &Application{
		cfg: cfg,
	}

	app.manager = workers.NewManager(app)

	app.cfg.Metrics().Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "uptime_seconds",
			Help: "Service uptime (seconds)",
		},
		func() float64 {
			return float64(time.Now().Unix() - app.startTst)
		},
	))

	return app
}

func (app *Application) Start() {
	app.startTst = time.Now().Unix()

	app.updateLogging()

	app.manager.Start()
	app.cfg.Log().Info().
		Int("pid", os.Getpid()).
		Msg("Service is ready")

	app.waitForShutdown()
}

func (app *Application) updateLogging() {
	// Get ready to listen to SIGHUP events.
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGHUP)

	// Decide output: file vs. inherited stdout.
	var out io.Writer
	if app.cfg.Logfile() != "" {
		// Open & set output file.
		file, err := helpers.NewLogFileWriter(app.cfg.Logfile(), 0640, true)
		if err != nil {
			app.cfg.Log().Fatal().
				Err(err).
				Str("file", app.cfg.Logfile()).
				Msg("Failed to open log file!")
			return
		}
		out = file

		// Listen to SIGHUP events.
		go func() {
			for {
				sig := <-channel

				app.cfg.Log().Info().
					Stringer("signal", sig).
					Msg("Got system signal: reloading log file")

				if err := file.Reopen(); err == nil {
					app.cfg.Log().Info().
						Str("file", app.cfg.Logfile()).
						Msg("Log file has been reopened")
				} else {
					app.cfg.Log().Error().
						Err(err).
						Str("file", app.cfg.Logfile()).
						Msg("Failed to reopen log file!")
				}
			}
		}()
	} else {
		// Use inherited stdout as output.
		out = os.Stdout

		// Listen to SIGHUP events.
		go func() {
			for {
				sig := <-channel

				app.cfg.Log().Info().
					Stringer("signal", sig).
					Msg("Got system signal: using stdout, so ignoring it")
			}
		}()
	}

	// Configure logging.
	logContext := zerolog.New(out).With().Timestamp()
	if app.cfg.LogCaller() {
		logContext = logContext.Caller()
	}
	log := logContext.Logger().Level(app.cfg.Loglevel())
	if !app.cfg.LogJSON() {
		log = log.Output(zerolog.ConsoleWriter{
			Out:        out,
			NoColor:    true,
			TimeFormat: time.RFC3339,
		})
	}
	app.cfg.SetLog(config.NewLogger(&log))
}

func (app *Application) waitForShutdown() {
	signals := []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}
	channel := make(chan os.Signal, len(signals))
	signal.Notify(channel, signals...)
	sig := <-channel

	app.cfg.Log().Info().
		Stringer("signal", sig).
		Msg("Got system signal: shutting the service down")

	app.manager.Stop()
}

func (app *Application) Cfg() *config.Config {
	return app.cfg
}
