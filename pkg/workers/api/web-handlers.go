package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"github.com/allenta/varnishmon/assets"
	"github.com/allenta/varnishmon/pkg/config"
	"github.com/allenta/varnishmon/pkg/workers/storage"
	"github.com/valyala/fasthttp"
)

const (
	developmentAssetsRoot = "/mnt/host/assets"
)

var (
	errMissingQueryArgsParam = errors.New("missing query string parameter")
	errInvalidQueryArgsParam = errors.New("invalid query string parameter")
)

func (h *Handler) filesystemHandler() *fasthttp.FS {
	fs := &fasthttp.FS{
		FS:                 assets.StaticFS,
		IndexNames:         []string{},
		Compress:           true,
		GenerateIndexPages: false,
		PathRewrite: func(rctx *fasthttp.RequestCtx) []byte {
			return append([]byte("/static"), rctx.Path()...)
		},
		PathNotFound: func(rctx *fasthttp.RequestCtx) {
			rctx.SetStatusCode(fasthttp.StatusNotFound)
		},
	}

	// In development mode, bypass the embedded filesystem and serve assets
	// directly from the host filesystem. This avoids the need to rebuild the
	// binary for every change. The absolute path to the assets directory is
	// hardcoded here, assuming the official development environment is
	// reasonable enough for anyone contributing to the project.
	if config.IsDevelopment() {
		fs.FS = nil
		fs.Compress = false
		fs.Root = developmentAssetsRoot
		fs.CacheDuration = 0
		fs.SkipCache = true
	}

	return fs
}

func (h *Handler) handleHomeRequest(rctx *fasthttp.RequestCtx) {
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
				rctx.SetStatusCode(fasthttp.StatusInternalServerError)
				return
			}
		}
		tmpl = h.homeTemplate
	} else {
		// The absolute path to the assets directory is hardcoded here, assuming
		// the official development environment is reasonable enough for anyone
		// contributing to the project.
		tmpl, err = template.ParseFiles(developmentAssetsRoot + "/templates/index.html.tmpl")
		if err != nil {
			h.app.Cfg().Log().Error().
				Err(err).
				Msg("Failed to parse 'templates/index.html.tmpl' template!")
			rctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
	}

	// Prepare template data & render it.
	cfg, err := json.Marshal(map[string]interface{}{
		"version":  config.Version(),
		"revision": config.Revision(),
		"config": map[string]interface{}{
			"scraper": map[string]interface{}{
				"enabled": h.app.Cfg().ScraperEnabled(),
				"period":  h.app.Cfg().ScraperPeriod().Seconds(),
			},
		},
		"storage": map[string]interface{}{
			"hostname": h.storage.Hostname(),
			"earliest": h.storage.Earliest().Unix(),
			"latest":   h.storage.Latest().Unix(),
		},
	})
	if err != nil {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to config for 'templates/index.html.tmpl' template!")
		rctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	tmplData := map[string]interface{}{
		"Version":  config.Version(),
		"Revision": config.Revision(),
		"Hostname": h.storage.Hostname(),
		"Config":   string(cfg),
	}
	var renderedTmpl bytes.Buffer
	if err := tmpl.Execute(&renderedTmpl, tmplData); err != nil {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to render 'static/index.html.tmpl' template!")
		rctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	// Set response headers & body.
	rctx.SetContentType("text/html; charset=utf-8")
	rctx.SetStatusCode(fasthttp.StatusOK)
	rctx.SetBody(renderedTmpl.Bytes())
}

func (h *Handler) handleStorageMetricsRequest(rctx *fasthttp.RequestCtx) {
	idRaw := rctx.UserValue("id")
	var result map[string]interface{}
	var err error

	// Extract 'from' query string parameter.
	from, err := h.getQueryArgsTimeParam(rctx, "from")
	if err != nil {
		rctx.SetStatusCode(fasthttp.StatusBadRequest)
		rctx.SetBodyString("Invalid 'from' parameter")
		return
	}

	// Extract 'to' query string parameter.
	to, err := h.getQueryArgsTimeParam(rctx, "to")
	if err != nil {
		rctx.SetStatusCode(fasthttp.StatusBadRequest)
		rctx.SetBodyString("Invalid 'to' parameter")
		return
	}

	// Extract 'step' query string parameter.
	step, err := rctx.QueryArgs().GetUint("step")
	if err != nil {
		rctx.SetStatusCode(fasthttp.StatusBadRequest)
		rctx.SetBodyString("Invalid 'step' parameter")
		return
	}

	// If no metric ID is provided, return info about all metrics, filtering
	// out the irrelevant (i.e., without samples) ones.
	if idRaw == nil {
		result, err = h.storage.GetMetrics(from, to, step)
	} else {
		// Validate metric ID.
		var id int
		id, err = strconv.Atoi(idRaw.(string))
		if err != nil {
			rctx.SetStatusCode(fasthttp.StatusBadRequest)
			rctx.SetBodyString(fmt.Sprintf("Invalid metric ID: %s", idRaw))
			return
		}

		// Extract 'aggregator' query string parameter.
		if !rctx.QueryArgs().Has("aggregator") {
			rctx.SetStatusCode(fasthttp.StatusBadRequest)
			rctx.SetBodyString("Missing 'aggregator' parameter")
			return
		}
		aggregator := string(rctx.QueryArgs().Peek("aggregator"))

		// Get metric data.
		result, err = h.storage.GetMetric(id, from, to, step, aggregator)
	}

	// Check for errors.
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUnknownMetricID):
			rctx.SetStatusCode(fasthttp.StatusNotFound)
			rctx.SetBodyString("Unknown metric ID")
		case errors.Is(err, storage.ErrInvalidFromTo):
			rctx.SetStatusCode(fasthttp.StatusBadRequest)
			rctx.SetBodyString("Invalid 'from' and 'to' parameters")
		case errors.Is(err, storage.ErrInvalidAggregator):
			rctx.SetStatusCode(fasthttp.StatusBadRequest)
			rctx.SetBodyString("Invalid 'aggregator' parameter")
		default:
			h.app.Cfg().Log().Error().
				Err(err).
				Msg("Failed to get metric(s) from storage!")
			rctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
		return
	}

	// Encode response.
	if err := json.NewEncoder(rctx).Encode(result); err == nil {
		rctx.SetContentType("application/json; charset=utf-8")
		rctx.SetStatusCode(fasthttp.StatusOK)
	} else {
		h.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to encode response!")
		rctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}

func (h *Handler) getQueryArgsTimeParam(rctx *fasthttp.RequestCtx, name string) (time.Time, error) {
	value := rctx.QueryArgs().Peek(name)
	if value == nil {
		return time.Time{}, fmt.Errorf("%w: %s", errMissingQueryArgsParam, name)
	}

	seconds, err := strconv.Atoi(string(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s", errInvalidQueryArgsParam, name)
	}

	return time.Unix(int64(seconds), 0), nil
}
