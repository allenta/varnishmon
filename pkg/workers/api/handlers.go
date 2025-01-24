package api

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"
	"text/template"
	"time"

	"allenta.com/varnishmon/pkg/workers/storage"
	"github.com/fasthttp/router"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/valyala/fasthttp/pprofhandler"
)

type Handler struct {
	app     Application
	storage *storage.Storage
	router  *router.Router

	homeTemplate *template.Template

	requestsTotal          *prometheus.CounterVec
	requestsInflightTotal  prometheus.Gauge
	requestDurationSeconds *prometheus.SummaryVec
}

func NewHandler(app Application, storage *storage.Storage) *Handler {
	h := &Handler{
		app:     app,
		storage: storage,
		router:  router.New(),

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

	h.router.RedirectTrailingSlash = true
	h.router.GET("/debug/pprof/{profile:*}", h.handlePprofRequest)
	h.router.GET("/metrics", h.handleMetricsRequest)
	h.router.GET("/storage/metrics", h.handleStorageMetricsRequest)
	h.router.GET("/storage/metrics/{id:[0-9]+}", h.handleStorageMetricsRequest)
	h.router.GET("/", h.handleHomeRequest)
	h.router.ServeFilesCustom("/{filepath:*}", h.filesystemHandler())

	h.app.Cfg().Metrics().Registry.MustRegister(h.requestsTotal)
	h.app.Cfg().Metrics().Registry.MustRegister(h.requestsInflightTotal)
	h.app.Cfg().Metrics().Registry.MustRegister(h.requestDurationSeconds)

	return h
}

func (h *Handler) HandleRequest(_ context.Context, rctx *fasthttp.RequestCtx) {
	// Update metrics.
	h.requestsInflightTotal.Inc()
	defer func(method string, start time.Time) {
		elapsed := time.Since(start).Seconds()
		code := strconv.Itoa(rctx.Response.StatusCode())

		h.requestsTotal.WithLabelValues(method, code).Inc()
		h.requestsInflightTotal.Dec()
		h.requestDurationSeconds.WithLabelValues(method, code).Observe(elapsed)
	}(string(rctx.Method()), time.Now())

	// Set no-cache headers for all responses.
	rctx.Response.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	rctx.Response.Header.Set("Pragma", "no-cache")
	rctx.Response.Header.Set("Expires", "0")

	// Check authentication.
	if h.app.Cfg().APIBasicAuthUsername() != "" && h.app.Cfg().APIBasicAuthPassword() != "" {
		authorized := false

		const prefix = "Basic "
		authHeader := string(rctx.Request.Header.Peek("Authorization"))
		if strings.HasPrefix(authHeader, prefix) {
			if c, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):]); err == nil {
				cs := string(c)
				s := strings.IndexByte(cs, ':')
				authorized = s >= 0 &&
					h.app.Cfg().APIBasicAuthUsername() == cs[:s] &&
					h.app.Cfg().APIBasicAuthPassword() == cs[s+1:]
			}
		}

		if !authorized {
			rctx.SetStatusCode(fasthttp.StatusUnauthorized)
			rctx.Response.Header.Add("WWW-Authenticate", "Basic realm=Restricted")
			return
		}
	}

	// Route request.
	h.router.Handler(rctx)
}

func (h *Handler) handlePprofRequest(rctx *fasthttp.RequestCtx) {
	pprofhandler.PprofHandler(rctx)
}

func (h *Handler) handleMetricsRequest(rctx *fasthttp.RequestCtx) {
	handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.HandlerFor(
		h.app.Cfg().Metrics().Registry,
		promhttp.HandlerOpts{}))
	handler(rctx)
}
