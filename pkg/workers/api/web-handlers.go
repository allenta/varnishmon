package api //nolint:revive,nolintlint

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
	"github.com/gin-gonic/gin"
)

const (
	developmentAssetsRoot = "/mnt/host/assets"
)

var (
	errMissingQueryArgsParam = errors.New("missing query string parameter")
	errInvalidQueryArgsParam = errors.New("invalid query string parameter")
)

type embedFilesystem struct {
	fs   http.FileSystem
	root string
}

func (efs embedFilesystem) Open(name string) (http.File, error) {
	return efs.fs.Open(path.Join(efs.root, name)) //nolint:wrapcheck
}

type wrappedFilesystem struct {
	fs http.FileSystem
}

func (wfs wrappedFilesystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) filesystemHandler() gin.HandlerFunc {
	// In development mode, bypass the embedded filesystem and serve assets
	// directly from the host filesystem. This avoids the need to rebuild the
	// binary for every change. The absolute path to the assets directory is
	// hardcoded here, assuming the official development environment is
	// reasonable enough for anyone contributing to the project.
	var fs http.FileSystem
	if config.IsDevelopment() {
		fs = http.Dir(path.Join(developmentAssetsRoot, "static"))
	} else {
		fs = embedFilesystem{fs: http.FS(assets.StaticFS), root: "static"}
	}
	wfs := wrappedFilesystem{fs: fs}

	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			wfs.ServeHTTP(c.Writer, c.Request)
		} else {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
		}
	}
}

func (h *Handler) handleHomeRequest(c *gin.Context) {
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
				c.AbortWithStatus(http.StatusInternalServerError)
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
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	// Prepare template data & render it.
	cfg, err := h.getConfigObject()
	if err != nil {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to build config for 'templates/index.html.tmpl' template!")
		c.AbortWithStatus(http.StatusInternalServerError)
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
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Set response headers & body.
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, renderedTmpl.String())
}

func (h *Handler) handleConfigRequest(c *gin.Context) {
	cfg, err := h.getConfigObject()
	if err != nil {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to build config object!")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.String(http.StatusOK, string(cfg))
}

func (h *Handler) handleStorageMetricsRequest(c *gin.Context) {
	idRaw := c.Param("id")
	var result map[string]any
	var err error

	// Extract 'from' query string parameter.
	from, err := h.getQueryArgsTimeParam(c, "from")
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid 'from' parameter")
		return
	}

	// Extract 'to' query string parameter.
	to, err := h.getQueryArgsTimeParam(c, "to")
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid 'to' parameter")
		return
	}

	// Extract 'step' query string parameter.
	step, err := strconv.Atoi(c.Query("step"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid 'step' parameter")
		return
	}

	// If no metric ID is provided, return info about all metrics, filtering
	// out the irrelevant (i.e., without samples) ones.
	if idRaw == "" {
		result, err = h.storage.GetMetrics(from, to, step, 0, 0)
	} else {
		// Validate metric ID.
		var id int
		id, err = strconv.Atoi(idRaw)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Invalid metric ID: %s", idRaw)) //nolint:perfsprint
			return
		}

		// Extract 'aggregator' query string parameter.
		aggregator := c.Query("aggregator")
		if aggregator == "" {
			c.String(http.StatusBadRequest, "Missing 'aggregator' parameter")
			return
		}

		// Get metric data.
		result, err = h.storage.GetMetric(id, from, to, step, aggregator)
	}

	// Check for errors.
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUnknownMetricID):
			c.String(http.StatusNotFound, "Unknown metric ID")
		case errors.Is(err, storage.ErrInvalidFromTo):
			c.String(http.StatusBadRequest, "Invalid 'from' and 'to' parameters")
		case errors.Is(err, storage.ErrInvalidAggregator):
			c.String(http.StatusBadRequest, "Invalid 'aggregator' parameter")
		default:
			h.app.Cfg().Log().Error().
				Err(err).
				Msg("Failed to get metric(s) from storage!")
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	// Encode response.
	c.JSON(http.StatusOK, result)
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

func (h *Handler) getQueryArgsTimeParam(c *gin.Context, name string) (time.Time, error) {
	value := c.Query(name)
	if value == "" {
		return time.Time{}, fmt.Errorf("%w: %s", errMissingQueryArgsParam, name)
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s", errInvalidQueryArgsParam, name)
	}

	return time.Unix(int64(seconds), 0), nil
}
