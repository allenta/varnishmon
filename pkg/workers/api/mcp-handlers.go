package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/allenta/varnishmon/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	errInvalidMCPArg = errors.New("invalid MCP argument")
)

func (h *Handler) newMCPServer() *server.MCPServer {
	mcpServer := server.NewMCPServer(
		"varnishmon",
		config.Version(),
	)

	// XXX: tools vs. resources. To be explored. This is just a PoC for now.

	mcpServer.AddTool(
		mcp.NewTool(
			"varnishmon-timestamp",
			mcp.WithDescription(
				`Given a UNIX timestamp in seconds, returns a human-friendly
representation in the timezone of the 'varnishmon' server`),
			mcp.WithNumber("value",
				mcp.Description(
					"The UNIX timestamp in seconds to convert to a human-friendly format"),
				mcp.Required(),
			)),
		h.handleHumanFriendlyTimestampTool)

	mcpServer.AddTool(
		mcp.NewTool(
			"varnishmon-config",
			mcp.WithDescription(
				`Retrieves the 'varnishmon' configuration, including the earliest
('storage.earliest') and latest ('storage.latest') UNIX timestamps in seconds
for metrics, as well as the metric collection frequency in seconds
('config.scrapper.period') when the scraper is enabled. The 'varnishmon-timestamp'
tool should be used to convert the timestamps to a human-friendly format`)),
		h.handleConfigTool)

	mcpServer.AddTool(
		mcp.NewTool("varnishmon-metrics",
			mcp.WithDescription(
				`Retrieves a list of 'varnishmon' metrics that have at least one
sample within the specified time range. Use 'varnishmon-config' to obtain recommended
default values for 'from' (i.e., 'storage.earliest'), 'to' (i.e., 'storage.latest'),
and 'step' (i.e., 'config.scrapper.period' if greater than 0, otherwise a
reasonable default like 60 seconds). For each metric, the response includes an
'id', 'name', 'description', 'flag' ('c': counter, 'g': gauge, 'b': bitmap,
'q': boolean), and 'format' ('i': integer, 'd': duration, 'B': bytes, 'b': bitmap,
'q': boolean). The 'varnishmon-timestamp' tool should be used to convert the
timestamps to a human-friendly format. Pagination is supported with 'page' and
'page-size' parameters`),
			mcp.WithNumber("from",
				mcp.Description(
					"The start of the time range as a UNIX timestamp in seconds"),
				mcp.Required(),
			),
			mcp.WithNumber("to",
				mcp.Description(
					"The end of the time range as a UNIX timestamp in seconds"),
				mcp.Required(),
			),
			mcp.WithNumber("step",
				mcp.Description(
					"The step size in seconds, defining the granularity of the time axis"),
				mcp.Required(),
			),
			mcp.WithNumber("page",
				mcp.Description(
					"The page number for pagination. Use 0 to disable pagination"),
				mcp.Required(),
				mcp.DefaultNumber(500),
				mcp.Min(0),
			),
			mcp.WithNumber("page-size",
				mcp.Description(
					"The number of items per page for pagination. Use 0 to disable pagination"),
				mcp.Required(),
				mcp.DefaultNumber(500),
				mcp.Min(0),
			)),
		h.handleMetricsTool)

	mcpServer.AddTool(
		mcp.NewTool("varnishmon-metric",
			mcp.WithDescription(
				`Retrieves samples of a specific 'varnishmon' metric. Use
'varnishmon-config' to obtain recommended default values for 'from' (i.e.,
'storage.earliest'), 'to' (i.e., 'storage.latest'), and 'step' (i.e.,
'config.scrapper.period' if greater than 0, otherwise a reasonable default like
60 seconds). The recommended default value for 'aggregator' depends on the
metric's 'flag': for binary metrics (i.e., 'flag' = 'b'), valid aggregators
include 'first', 'last', 'bit_and', 'bit_or', 'bit_xor', and 'count', with
'bit_and' being the recommended default. For other metrics, valid aggregators
include 'avg', 'min', 'max', 'first', 'last', and 'count', with 'avg' being the
recommended default. Aggregators are used to consolidate samples when multiple
samples exist within a single time step. Samples are returned as a list of
tuples containing the UNIX timestamp in seconds and the sample value. For
counter metrics (i.e., 'flag' = 'c'), the value is expressed as a rate per
second. The 'varnishmon-timestamp' tool should be used to convert the timestamps
to a human-friendly format. Pagination is not supported for this tool, but you
can control the number of samples returned by using the 'step' parameter`),
			mcp.WithNumber("id",
				mcp.Description(
					"The ID of the metric, as returned by 'varnishmon-metrics'"),
				mcp.Required(),
			),
			mcp.WithNumber("from",
				mcp.Description(
					"The start of the time range as a UNIX timestamp in seconds"),
				mcp.Required(),
			),
			mcp.WithNumber("to",
				mcp.Description(
					"The end of the time range as a UNIX timestamp in seconds"),
				mcp.Required(),
			),
			mcp.WithNumber("step",
				mcp.Description(
					"The step size in seconds, defining the granularity of the time axis"),
				mcp.Required(),
			),
			mcp.WithString("aggregator",
				mcp.Description(
					"The aggregator to use for consolidating samples when multiple "+
						"samples exist within a single time step"),
				mcp.Required(),
			)),
		h.handleMetricTool)

	return mcpServer
}

type humanFriendlyTimestampToolArgs struct {
	Value float64 `json:"value"`
}

func (h *Handler) handleHumanFriendlyTimestampTool(
	_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args humanFriendlyTimestampToolArgs
	if err := request.BindArguments(&args); err != nil {
		return nil, errInvalidMCPArg
	}

	timestamp := time.Unix(int64(args.Value), 0).In(time.Local) //nolint:gosmopolitan
	humanFriendlyTimestamp := timestamp.Format(time.RFC1123)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: humanFriendlyTimestamp,
			},
		},
	}, nil
}

func (h *Handler) handleConfigTool(
	_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfgMarshaled, err := h.getConfigObject()
	if err != nil {
		return nil, fmt.Errorf("failed to get config object: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(cfgMarshaled),
			},
		},
	}, nil
}

type MetricsToolArgs struct {
	From     float64 `json:"from"`
	To       float64 `json:"to"`
	Step     float64 `json:"step"`
	Page     float64 `json:"page"`
	PageSize float64 `json:"page-size"`
}

func (h *Handler) handleMetricsTool(
	_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args MetricsToolArgs
	if err := request.BindArguments(&args); err != nil {
		return nil, errInvalidMCPArg
	}

	metrics, err := h.storage.GetMetrics(
		time.Unix(int64(args.From), 0),
		time.Unix(int64(args.To), 0),
		int(args.Step),
		int(args.Page),
		int(args.PageSize))
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	metricsMarshaled, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(metricsMarshaled),
			},
		},
	}, nil
}

type MetricToolArgs struct {
	ID         float64 `json:"id"`
	From       float64 `json:"from"`
	To         float64 `json:"to"`
	Step       float64 `json:"step"`
	Aggregator string  `json:"aggregator"`
}

func (h *Handler) handleMetricTool(
	_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args MetricToolArgs
	if err := request.BindArguments(&args); err != nil {
		return nil, errInvalidMCPArg
	}

	metric, err := h.storage.GetMetric(
		int(args.ID),
		time.Unix(int64(args.From), 0),
		time.Unix(int64(args.To), 0),
		int(args.Step),
		args.Aggregator)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	metricMarshaled, err := json.Marshal(metric)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metric: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(metricMarshaled),
			},
		},
	}, nil
}
