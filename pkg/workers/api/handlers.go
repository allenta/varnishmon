package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "net/http/pprof" //nolint:gosec

	"github.com/allenta/varnishmon/pkg/config"
	"github.com/allenta/varnishmon/pkg/workers/storage"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	app          Application
	serverHeader string
	storage      *storage.Storage
	router       *mux.Router

	homeTemplate *template.Template

	requestsTotal          *prometheus.CounterVec
	requestsInflightTotal  prometheus.Gauge
	requestDurationSeconds *prometheus.SummaryVec
}

func NewHandler(app Application, storage *storage.Storage) *Handler {
	h := &Handler{
		app:          app,
		serverHeader: fmt.Sprintf("varnishmon/%s (%s)", config.Version(), config.Revision()),
		storage:      storage,
		router:       mux.NewRouter().StrictSlash(true),

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

	sr := h.router.Methods("GET", "HEAD").Subrouter()
	sr.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	sr.HandleFunc("/metrics", h.handleMetricsRequest)
	sr.HandleFunc("/storage/metrics", h.handleStorageMetricsRequest)
	sr.HandleFunc("/storage/metrics/{id:[0-9]+}", h.handleStorageMetricsRequest)
	sr.HandleFunc("/", h.handleHomeRequest)
	sr.HandleFunc("/config", h.handleConfigRequest)
	sr.PathPrefix("/").Handler(h.filesystemHandler())

	h.app.Cfg().Metrics().Registry.MustRegister(h.requestsTotal)
	h.app.Cfg().Metrics().Registry.MustRegister(h.requestsInflightTotal)
	h.app.Cfg().Metrics().Registry.MustRegister(h.requestDurationSeconds)

	return h
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Wrap the original 'ResponseWriter' in order to capture the status code,
	// needed for metrics.
	rw := &responseWriter{ResponseWriter: w}

	// Update metrics.
	h.requestsInflightTotal.Inc()
	defer func(method string, start time.Time) {
		elapsed := time.Since(start).Seconds()
		code := strconv.Itoa(rw.statusCode)

		h.requestsTotal.WithLabelValues(method, code).Inc()
		h.requestsInflightTotal.Dec()
		h.requestDurationSeconds.WithLabelValues(method, code).Observe(elapsed)
	}(r.Method, time.Now())

	// Set 'Server' header.
	rw.Header().Set("Server", h.serverHeader)

	// Set no-cache headers for all responses.
	rw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	rw.Header().Set("Pragma", "no-cache")
	rw.Header().Set("Expires", "0")

	// Check authorization.
	if h.app.Cfg().APIBasicAuthUsername() != "" && h.app.Cfg().APIBasicAuthPassword() != "" {
		authorized := false

		const prefix = "Basic "
		authHeader := r.Header.Get("Authorization")
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
			rw.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Route request.
	h.router.ServeHTTP(rw, r)
}

func (h *Handler) handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	promhttp.HandlerFor(
		h.app.Cfg().Metrics().Registry,
		promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
