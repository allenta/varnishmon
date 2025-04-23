package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/allenta/varnishmon/pkg/config"
	"github.com/allenta/varnishmon/pkg/workers/storage"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	app          Application
	serverHeader string
	storage      *storage.Storage
	router       *gin.Engine

	homeTemplate *template.Template
	sseServer    *server.SSEServer

	requestsTotal          *prometheus.CounterVec
	requestsInflightTotal  prometheus.Gauge
	requestDurationSeconds *prometheus.SummaryVec
}

func NewHandler(app Application, storage *storage.Storage) *Handler {
	gin.SetMode(gin.ReleaseMode)

	h := &Handler{
		app:          app,
		serverHeader: fmt.Sprintf("varnishmon/%s (%s)", config.Version(), config.Revision()),
		storage:      storage,
		router:       gin.New(),

		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_requests_total",
				Help: "API requests processed, partitioned by status code and HTTP method",
			},
			[]string{"method", "code"}),

		requestsInflightTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "api_requests_inflight",
				Help: "API requests inflight",
			}),

		requestDurationSeconds: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "api_request_duration_seconds",
				Help:       "API request duration, partitioned by status code and HTTP method",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
				MaxAge:     1 * time.Minute,
			},
			[]string{"method", "code"}),
	}

	h.sseServer = server.NewSSEServer(
		h.newMCPServer(),
		server.WithBasePath("/mcp"),
		server.WithUseFullURLForMessageEndpoint(true),
		server.WithBaseURL(""))

	h.router.RedirectTrailingSlash = true
	h.router.RedirectFixedPath = false
	h.router.HandleMethodNotAllowed = false
	h.router.ForwardedByClientIP = true
	h.router.UseRawPath = false
	h.router.UnescapePathValues = true
	h.router.SetTrustedProxies(nil) //nolint:errcheck

	h.router.Use(
		gin.CustomRecoveryWithWriter(
			h.app.Cfg().Log().ErrorWriter(),
			func(c *gin.Context, recovered any) {
				h.app.Cfg().Log().Error().
					Interface("recovered", recovered).
					Str("url", c.Request.URL.String()).
					Str("method", c.Request.Method).
					Msg("Recovered from gin panic!")
				c.AbortWithStatus(http.StatusInternalServerError)
			}),
		gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))
	if app.Cfg().APIBasicAuthUsername() != "" && app.Cfg().APIBasicAuthPassword() != "" {
		h.router.Use(
			gin.BasicAuth(gin.Accounts{
				app.Cfg().APIBasicAuthUsername(): app.Cfg().APIBasicAuthPassword(),
			}))
	}

	pprof.Register(h.router)
	h.router.GET("/metrics", h.handleMetricsRequest)
	h.router.GET("/storage/metrics", h.handleStorageMetricsRequest)
	h.router.GET("/storage/metrics/:id", h.handleStorageMetricsRequest)
	h.router.GET("/", h.handleHomeRequest)
	h.router.GET("/config", h.handleConfigRequest)
	h.router.Any("/mcp/*action", func(ctx *gin.Context) {
		h.sseServer.ServeHTTP(ctx.Writer, ctx.Request)
	})
	h.router.NoRoute(h.filesystemHandler())

	h.app.Cfg().Metrics().Registry.MustRegister(h.requestsTotal)
	h.app.Cfg().Metrics().Registry.MustRegister(h.requestsInflightTotal)
	h.app.Cfg().Metrics().Registry.MustRegister(h.requestDurationSeconds)

	return h
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Gin skips the execution of middlewares in some cases (e.g., automatic
	// redirects due to trailing slashes). This is not a major issue, but it is
	// the reason why the following logic is here and not part of a middleware.

	// Wrap the original 'ResponseWriter' in order to capture the status code,
	// needed for metrics.
	rw := &responseWriter{ResponseWriter: w}

	// Update metrics.
	h.requestsInflightTotal.Inc()
	defer func(method string, start time.Time) {
		elapsed := time.Since(start).Seconds()
		code := strconv.Itoa(rw.code)

		h.requestsTotal.WithLabelValues(method, code).Inc()
		h.requestsInflightTotal.Dec()
		h.requestDurationSeconds.WithLabelValues(method, code).Observe(elapsed)
	}(r.Method, time.Now())

	// Set default 'Server' header for all responses.
	rw.Header().Set("Server", h.serverHeader)

	// Set default no-cache headers for all responses.
	rw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	rw.Header().Set("Pragma", "no-cache")
	rw.Header().Set("Expires", "0")

	// Route the request.
	h.router.ServeHTTP(rw, r)
}

func (h *Handler) Shutdown() error {
	return h.sseServer.Shutdown(context.Background()) //nolint:wrapcheck
}

func (h *Handler) handleMetricsRequest(c *gin.Context) {
	// Compression is explicitly disabled here to avoid double compression by
	// the 'gzip' middleware.
	promhttp.HandlerFor(
		h.app.Cfg().Metrics().Registry,
		promhttp.HandlerOpts{DisableCompression: true}).ServeHTTP(c.Writer, c.Request)
}
