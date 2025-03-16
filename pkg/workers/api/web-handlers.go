package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"text/template"
	"time"

	"github.com/allenta/varnishmon/assets"
	"github.com/allenta/varnishmon/pkg/config"
	"github.com/allenta/varnishmon/pkg/workers/storage"
	"github.com/gorilla/mux"
)

const (
	developmentAssetsRoot = "/mnt/host/assets"
)

var (
	errMissingQueryArgsParam = errors.New("missing query string parameter")
	errInvalidQueryArgsParam = errors.New("invalid query string parameter")
)

type webFilesystem struct {
	fs   http.FileSystem
	root string
}

func (wf webFilesystem) Open(name string) (http.File, error) {
	return wf.fs.Open(path.Join(wf.root, name)) //nolint:wrapcheck
}

type webFileServer struct {
	fs http.FileSystem
}

func (wfs webFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if the file exists.
	file, err := wfs.fs.Open(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Disable directory listing.
	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	// Serve the file.
	http.ServeContent(w, r, r.URL.Path, stat.ModTime(), file)
}

func (h *Handler) filesystemHandler() http.Handler {
	// In development mode, bypass the embedded filesystem and serve assets
	// directly from the host filesystem. This avoids the need to rebuild the
	// binary for every change. The absolute path to the assets directory is
	// hardcoded here, assuming the official development environment is
	// reasonable enough for anyone contributing to the project.
	var fs http.FileSystem
	if config.IsDevelopment() {
		fs = http.Dir(path.Join(developmentAssetsRoot, "static"))
	} else {
		fs = webFilesystem{fs: http.FS(assets.StaticFS), root: "static"}
	}
	return webFileServer{fs: fs}
}

func (h *Handler) handleHomeRequest(w http.ResponseWriter, _ *http.Request) {
	// Fetch the template. In development mode, the template is loaded from the
	// host filesystem. In production mode, the template is loaded from the
	// embedded filesystem, parsed once, and reused for every request.
	var tmpl *template.Template
	var err error
	if !config.IsDevelopment() {
		if h.homeTemplate == nil {
			h.homeTemplate, err = template.ParseFS(assets.TemplatesFS, "templates/index.html.tmpl")
			if err != nil {
				h.app.Cfg().Log().Error().
					Err(err).
					Msg("Failed to parse 'templates/index.html.tmpl' template!")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
		tmpl = h.homeTemplate
	} else {
		// The absolute path to the assets directory is hardcoded here, assuming
		// the official development environment is reasonable enough for anyone
		// contributing to the project.
		tmpl, err = template.ParseFiles(path.Join(developmentAssetsRoot, "templates", "index.html.tmpl"))
		if err != nil {
			h.app.Cfg().Log().Error().
				Err(err).
				Msg("Failed to parse 'templates/index.html.tmpl' template!")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Prepare template data & render it.
	cfg, err := h.getConfigObject()
	if err != nil {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to build config for 'templates/index.html.tmpl' template!")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmplData := map[string]any{
		"Version":  config.Version(),
		"Revision": config.Revision(),
		"Config":   string(cfg),
	}
	var renderedTmpl bytes.Buffer
	if err := tmpl.Execute(&renderedTmpl, tmplData); err != nil {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to render 'static/index.html.tmpl' template!")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set response headers & body.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(renderedTmpl.Bytes()) //nolint:errcheck
}

func (h *Handler) handleConfigRequest(w http.ResponseWriter, _ *http.Request) {
	cfg, err := h.getConfigObject()
	if err != nil {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to build config object!")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(cfg) //nolint:errcheck
}

func (h *Handler) handleStorageMetricsRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idRaw := vars["id"]
	var result map[string]any
	var err error

	// Extract 'from' query string parameter.
	from, err := h.getQueryArgsTimeParam(r, "from")
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}

	// Extract 'to' query string parameter.
	to, err := h.getQueryArgsTimeParam(r, "to")
	if err != nil {
		http.Error(w, "Invalid 'to' parameter", http.StatusBadRequest)
		return
	}

	// Extract 'step' query string parameter.
	step, err := strconv.Atoi(r.URL.Query().Get("step"))
	if err != nil {
		http.Error(w, "Invalid 'step' parameter", http.StatusBadRequest)
		return
	}

	// If no metric ID is provided, return info about all metrics, filtering
	// out the irrelevant (i.e., without samples) ones.
	if idRaw == "" {
		result, err = h.storage.GetMetrics(from, to, step)
	} else {
		// Validate metric ID.
		var id int
		id, err = strconv.Atoi(idRaw)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid metric ID: %s", idRaw), http.StatusBadRequest) //nolint:perfsprint
			return
		}

		// Extract 'aggregator' query string parameter.
		aggregator := r.URL.Query().Get("aggregator")
		if aggregator == "" {
			http.Error(w, "Missing 'aggregator' parameter", http.StatusBadRequest)
			return
		}

		// Get metric data.
		result, err = h.storage.GetMetric(id, from, to, step, aggregator)
	}

	// Check for errors.
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUnknownMetricID):
			http.Error(w, "Unknown metric ID", http.StatusNotFound)
		case errors.Is(err, storage.ErrInvalidFromTo):
			http.Error(w, "Invalid 'from' and 'to' parameters", http.StatusBadRequest)
		case errors.Is(err, storage.ErrInvalidAggregator):
			http.Error(w, "Invalid 'aggregator' parameter", http.StatusBadRequest)
		default:
			h.app.Cfg().Log().Error().
				Err(err).
				Msg("Failed to get metric(s) from storage!")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Encode response.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	} else {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to encode response!")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *Handler) getConfigObject() ([]byte, error) {
	scraperPeriod := 0
	if h.app.Cfg().ScraperEnabled() {
		scraperPeriod = int(h.app.Cfg().ScraperPeriod().Seconds())
	}
	cfg, err := json.Marshal(map[string]any{
		"version":  config.Version(),
		"revision": config.Revision(),
		"config": map[string]any{
			"scraper": map[string]any{
				"enabled": h.app.Cfg().ScraperEnabled(),
				"period":  scraperPeriod,
			},
		},
		"storage": map[string]any{
			"hostname": h.storage.Hostname(),
			"earliest": h.storage.Earliest().Unix(),
			"latest":   h.storage.Latest().Unix(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config object: %w", err)
	}

	return cfg, nil
}

func (h *Handler) getQueryArgsTimeParam(r *http.Request, name string) (time.Time, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return time.Time{}, fmt.Errorf("%w: %s", errMissingQueryArgsParam, name)
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s", errInvalidQueryArgsParam, name)
	}

	return time.Unix(int64(seconds), 0), nil
}
