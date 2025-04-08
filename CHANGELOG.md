- 0.7.1-1 (2025-04-08):
    + Updated dependencies.
        * DuckDB 1.2.0 ➙ 1.2.1
        * Go 1.24.0 ➙ 1.24.1
            - golangci-lint 1.64.5 ➙ 2.0.2
            - gzip 1.2.2 ➙ 1.2.3
            - sys 0.31.0 ➙ 0.32.0
            - viper 1.20.0 ➙ 1.20.1
            - zerolog 1.33.0 ➙ 1.34.0
        * Node.js
            - @eslint/js 9.22.0 ➙ 9.23.0
            - @types/react 19.0.11 ➙ 19.0.12
            - @vitejs/plugin-react-swc 3.8.0 ➙ 3.8.1
            - eslint 9.22.0 ➙ 9.23.0
            - eslint-plugin-prettier 5.2.3 ➙ 5.2.6
            - react 19.0.0 ➙ 19.1.0
            - react-dom 19.0.0 ➙ 19.1.0
            - sass-embedded 1.86.0 ➙ 1.86.3
            - vite 6.2.2 ➙ 6.2.5

- 0.7.0-1 (2025-03-18):
    + Dropped fasthttp in favor of the standard net/http package.
    + Integrated gin for routing.
    + Dropped 'api_worker_concurrency' & 'api_worker_open_connections' metrics.
    + Updated dependencies.
        * Go
            - sys 0.30.0 ➙ 0.31.0
            - viper 1.19.0 ➙ 1.20.0
        * Node.js
            - @types/react 19.0.10 ➙ 19.0.11
            - sass-embedded 1.85.1 ➙ 1.86.0
            - vite 6.2.0 ➙ 6.2.2

- 0.6.6-1 (2025-03-10):
    + Added `/config` endpoint to retrieve the configuration, as an alternative to hydrating the `index.html` template server-side.

- 0.6.5-1 (2025-03-10):
    + Assorted improvements to the new React-based UI.
    + Updated dependencies.
        * Node.js
            - @eslint/js 9.21.0 ➙ 9.22.0
            - eslint 9.21.0 ➙ 9.22.0
            - eslint-config-prettier 10.0.2 ➙ 10.1.1

- 0.6.4-1 (2025-03-07):
    + Assorted improvements to the new React-based UI.
    + Updated dependencies.
        * Go
            - client_golang 1.21.0 ➙ 1.21.1

- 0.6.3-1 (2025-03-06):
    + Replaced Webpack with Vite for the frontend build.
    + Used a smaller bundle of Plotly.js.

- 0.6.2-1 (2025-03-05):
    + Fixed rendering of Y-axis labels in charts.

- 0.6.1-1 (2025-03-05):
    + Applied assorted improvements to the new React-based UI.

- 0.6.0-1 (2025-03-03):
    + Replaced the vanilla JavaScript-based UI with React.
    + Replaced Flatpickr with Air Datepicker.
    + Fixed rendering of bitmap metrics when gaps are present.
    + Updated dependencies.
        * Go
            - go-duckdb 1.8.4 ➙ 1.8.5
        * Node.js
            - eslint 9.20.1 ➙ 9.21.0
            - eslint-webpack-plugin 4.2.0 ➙ 5.0.0
            - sass 1.85.0 ➙ 1.85.1
            - terser-webpack-plugin 5.3.11 ➙ 5.3.12

- 0.5.4-1 (2025-02-19):
    + Updated dependencies.
        * DuckDB 1.1.3 ➙ 1.2.0
        * uv 0.5.27 ➙ 0.6.0
        * Go
            - client_golang 1.20.5 ➙ 1.21.0
            - cobra 1.8.1 ➙ 1.9.1
            - fasthttp 1.58.0 ➙ 1.59.0
        * Node.js 22.13.1 ➙ 22.14.0
            - eslint 9.20.0 ➙ 9.20.1
            - plotly.js-dist 3.0.0 ➙ 3.0.1
            - postcss 8.5.1 ➙ 8.5.2
            - sass 1.84.0 ➙ 1.85.0
            - sass-loader 16.0.4 ➙ 16.0.5
            - webpack 5.97.1 ➙ 5.98.0
    + Added support for `rhel8` packaging in AMD64 architecture.

- 0.5.3-1 (2025-02-13):
    + Fixed step calculation when scraper is disabled (i.e., no scraping period available).

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
    + Added event handlers to apply the time range when `Enter` is pressed in one of the time range inputs.
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
