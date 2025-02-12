- 0.5.2-1 (2025-02-12):
    + Improved logrotate configuration.
    + Fixed DuckDB query when normalizing input `from` and `to` timestamps.

- 0.5.1-1 (2025-02-12):
    + Updated dependencies.
        * Go 1.23.5 ➙ 1.24.0
            - go-duckdb 1.8.3 ➙ 1.8.4
        * Node.js
            - eslint 9.19.0 ➙ 9.20.0
            - sass 1.83.4 ➙ 1.84.0
    + Packaging adjusted to run the service as `varnishlog:varnish` instead of `varnish:varnish`.

- 0.5.0-1 (2025-02-04):
    + Added support for newest varnishstat output format.

- 0.4.8-1 (2025-02-04):
    + Fixed the X-axis layout in charts during refresh, adding additional styling.
    + Used a monospace font for cluster names.
    + Fixed an issue with packaging that caused the service to reload instead of restart after an upgrade.

- 0.4.7-1 (2025-02-03):
    + Updated dependencies.
        * Go 1.23.3 ➙ 1.23.5
            - fasthttp 1.57.0 ➙ 1.58.0
            - goimports 0.27.0 ➙ 0.29.0
            - golangci-lint 1.62.2 ➙ 1.63.4
            - mockery 2.49.1 ➙ 2.52.1
            - router 1.5.3 ➙ 1.5.4
        * Node.js 22.13.0 ➙ 22.13.1
            - eslint 9.18.0 ➙ 9.19.0
            - plotly.js-dist 2.35.3 ➙ 3.0.0

- 0.4.6-1 (2025-02-03):
    + Added an action to the chart widget to copy the plot to the clipboard.

- 0.4.5-1 (2025-02-01):
    + Added `nocreate` to logrotate configuration.

- 0.4.4-1 (2025-01-31):
    + Fixed major I/O performance issue when writing metrics to DuckDB.

- 0.4.3-1 (2025-01-30):
    + Fixed zoom reset in charts.

- 0.4.2-1 (2025-01-28):
    + Fixed handling of gaps in metrics.
    + Moved the Go project to the GitHub namespace.

- 0.4.1-1 (2025-01-28):
    + Enabled ARM64 builds.

- 0.4.0-1 (2025-01-28):
    + Reworked the internals of the chart widget.
    + Modified the behavior to update (but not apply) the time range picker during zoom in/out events.
    + Limited the zoom of charts to a minimum and maximum range.
    + Adjusted the width of lines in charts.
    + Added a filter history to easily reuse previous filter strings.
    + Added visual feedback to charts when the effective step is different from the selected step.
    + Improved refreshing of charts.

- 0.3.1-1 (2025-01-26):
    + Fixed filtering of metrics when no search terms are provided.
    + Changed line shape in charts back to linear.

- 0.3.0-1 (2025-01-26):
    + Added extra logging during bootstrap / rotation of the storage.
    + Adjusted shape and width of lines in charts.
    + Added support to zoom in and out in the charts.
    + Increased default DuckDB memory limit to 512 MiB.
    + Added support to filter metrics by multiple search terms.

- 0.2.4-1 (2025-01-25):
    + Added `--memory-limit` flag to control the DuckDB memory limit from the command line.
    + Added check to avoid piling up scraping jobs when the internal metrics queue is full.
    + Changed behavior of the archiver worker when hitting DuckDB errors.

- 0.2.3-1 (2025-01-25):
    + Added spinner for visual feedback when loading a metric.

- 0.2.2-1 (2025-01-24):
    + Added event handlers to apply the time range when 'Enter' is pressed in one of the time range inputs.
    + Fixed overflow when processing bitmap metrics.

- 0.2.1-1 (2025-01-24):
    + Improved rendering of timeseries.

- 0.2.0-1 (2025-01-24):
    + Fixed logrotate configuration in RPM packages.
    + Extended the rules used to tag debug metrics in the client side.
    + Modified handling of uptimes, now processing them as a gauge.

- 0.1.1-1 (2025-01-24):
    + Fixed wrong user in RPM post-install script.
    + Fixed configuration discovery.
    + Fixed timestamp of scraped metrics ignoring the incomplete timestamp in the `varnishstat` output.

- 0.1.0-1 (2025-01-24):
    + Initial release.
