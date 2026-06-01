- ?
    + Migrate do [github.com/duckdb/duckdb-go](https://github.com/duckdb/duckdb-go#migration-from-marcboekergo-duckdb).
        * DB schema version bumped to 2.
        * Dropped FOREIGN KEY constraint on `metric_values.metric_id`.
    + Drop RHEL 8 support.
    + Updated dependencies:
        * Dev & CI environments:
            - DuckDB: 1.5.2 → 1.5.3.
            - Node.js: 24.15.0 → 24.16.0.
            - uv: 0.11.11 → 0.11.17.
        * Go:
            - modernize: v0.21.1 → v0.45.0.
            - mcp-go: v0.52.0 → v0.54.1.
            - x/sys: v0.44.0 → v0.45.0.
        * Node.js:
            - @types/react: 19.2.14 → 19.2.15.
            - @vitejs/plugin-react-swc: 4.3.0 → 4.3.1.
            - eslint-plugin-prettier: 5.5.5 → 5.5.6.
            - eslint-plugin-react-hooks: 7.0.1 → 7.1.1.
            - sass-embedded: 1.99.0 → 1.100.0.
            - vite: 8.0.12 → 8.0.15.
            - vite-plugin-minify: 2.1.0 → 3.0.0.

- 0.8.12-1 (2026-05-11):
    + Test release to try out changes in GitHub actions.

- 0.8.11-1 (2026-05-11):
    + Added support for Ubuntu Resolute.
    + Updated dependencies.
        * DuckDB 1.4.4 ➙ 1.5.2
        * Go 1.26.0 ➙ 1.26.3
            - gzip 1.2.5 ➙ 1.2.6
            - pprof 1.5.3 ➙ 1.5.4
            - mcp-go 0.44.1 ➙ 0.52.0
            - zerolog 1.34.0 ➙ 1.35.1
            - sys 0.41.0 ➙ v0.44.0
        * Node.js 24.14.0 ➙ 24.15.0
            - @vitejs/plugin-react-swc 4.2.3 ➙ 4.3.0
            - globals 17.4.0 ➙ 17.6.0
            - plotly.js-basic-dist 3.4.0 ➙ 3.5.1
            - prettier 3.8.1 ➙ 3.8.3
            - react 19.2.4 ➙ 19.2.6
            - react-dom 19.2.4 ➙ 19.2.6
            - sass-embedded 1.98.0 ➙ 1.99.0
            - vite 7.3.1 ➙ 8.0.12
            - vite-plugin-static-copy 3.2.0 ➙ 4.1.0
        * uv 0.10.7 ➙ 0.11.11

- 0.8.10-1 (2026-03-04):
    + Updated dependencies.
        * DuckDB 1.4.3 ➙ 1.4.4
        * Go 1.25.5 ➙ 1.26.0
            - gin 1.11.0 ➙ 1.12.0
            - mcp-go 0.43.2 ➙ 0.44.1
            - sys 0.40.0 ➙ 0.41.0
        * Node.js 24.12.0 ➙ 24.14.0
            - @fortawesome/fontawesome-free 7.1.0 ➙ 7.2.0
            - @types/react 19.2.7 ➙ 19.2.14
            - @vitejs/plugin-react-swc 4.2.2 ➙ 4.2.3
            - eslint-plugin-prettier 5.5.4 ➙ 5.5.5
            - eslint-plugin-react-refresh 0.4.26 ➙ 0.5.2
            - globals 17.0.0 ➙ 17.4.0
            - plotly.js-basic-dist 3.3.1 ➙ 3.4.0
            - prettier 3.7.4 ➙ 3.8.1
            - react 19.2.3 ➙ 19.2.4
            - react-dom 19.2.3 ➙ 19.2.4
            - sass-embedded 1.97.2 ➙ 1.97.3
            - vite-plugin-static-copy 3.1.4 ➙ 3.2.0
        * uv 0.9.22 ➙ 0.10.7

- 0.8.9-1 (2026-01-09):
    + Updated dependencies.
        * DuckDB 1.4.1 ➙ 1.4.3
        * Go 1.25.3 ➙ 1.25.5
            - cobra 1.10.1 ➙ 1.10.2
            - golangci-lint 2.6.0 ➙ 2.8.0
            - mcp-go 0.43.0 ➙ 0.43.2
            - modernize 0.20.0 ➙ 0.21.0
            - sys 0.37.0 ➙ 0.40.0
        * Node.js 24.11.0 ➙ 24.12.0
            - @eslint/js 9.39.1 ➙ 9.39.2
            - @types/react 19.2.2 ➙ 19.2.7
            - @types/react-dom 19.2.2 ➙ 19.2.3
            - @vitejs/plugin-react-swc 4.2.0 ➙ 4.2.2
            - eslint 9.39.1 ➙ 9.39.2
            - eslint-plugin-react-refresh 0.4.24 ➙ 0.4.26
            - globals 16.5.0 ➙ 17.0.0
            - plotly.js-basic-dist 3.2.0 ➙ 3.3.1
            - prettier 3.6.2 ➙ 3.7.4
            - react 19.2.0 ➙ 19.2.3
            - react-dom 19.2.0 ➙ 19.2.3
            - sass-embedded 1.93.3 ➙ 1.97.2
            - vite 7.1.12 ➙ 7.3.1
        * uv 0.9.7 ➙ 0.9.22

- 0.8.8-1 (2025-11-04):
    + Updated dependencies.
        * Node.js
            - @eslint/js 9.39.0 ➙ 9.39.1
            - eslint 9.39.0 ➙ 9.39.1

- 0.8.7-1 (2025-11-04):
    + Updated dependencies.
        * DuckDB 1.3.2 ➙ 1.4.1
        * FPM 1.16.0 ➙ 1.17.0
        * Go 1.25.0 ➙ 1.25.3
            - client_golang 1.23.0 ➙ 1.23.2
            - cobra 1.9.1 ➙ 1.10.1
            - gin 1.10.1 ➙ 1.11.0
            - golangci-lint 2.4.0 ➙ 2.5.0
            - gzip 1.2.3 ➙ 1.2.5
            - mcp-go 0.38.0 ➙ 0.43.0
            - sys 0.35.0 ➙ 0.37.0
            - testify 1.10.0 ➙ 1.11.1
            - viper 1.20.1 ➙ 1.21.0
        * Node.js 24.6.0 ➙ 24.11.0
            - @eslint/js 9.33.0 ➙ 9.39.0
            - @fortawesome/fontawesome-free 7.0.0 ➙ 7.1.0
            - @types/react 19.1.10 ➙ 19.2.2
            - @types/react-dom 19.1.7 ➙ 19.2.2
            - @vitejs/plugin-react-swc 4.0.1 ➙ 4.2.0
            - bootstrap 5.3.7 ➙ 5.3.8
            - eslint 9.33.0 ➙ 9.39.0
            - eslint-plugin-react-hooks 5.2.0 ➙ 7.0.1
            - eslint-plugin-react-refresh 0.4.20 ➙ 0.4.24
            - globals 16.3.0 ➙ 16.5.0
            - plotly.js-basic-dist 3.1.0 ➙ 3.2.0
            - react 19.1.1 ➙ 19.2.0
            - react-dom 19.1.1 ➙ 19.2.0
            - sass-embedded 1.90.0 ➙ 1.93.3
            - vite 7.1.3 ➙ 7.1.12
            - vite-plugin-static-copy 3.1.2 ➙ 3.1.4
        * uv 0.8.12 ➙ 0.9.7
    + Fixed segmentation fault when files referenced in the configuration are in an unreadable directory because of permissions.

- 0.8.6-1 (2025-08-21):
    + Added support for Debian Trixie and RHEL (Rocky Linux) 10.
    + Updated dependencies.
        * Go 1.24.4 ➙ 1.25.0
            - golangci-lint 2.3.0 ➙ 2.4.0
            - mcp-go 0.36.0 ➙ 0.38.0
            - sys 0.34.0 ➙ 0.35.0
        * Node.js 24.3.0 ➙ 24.6.0
            - @eslint/js 9.32.0 ➙ 9.3.0
            - @types/react 19.1.9 ➙ 19.1.0
            - @vitejs/plugin-react-swc 3.11.0 ➙ 4.0.1
            - eslint 9.32.0  ➙ 9.33.0
            - eslint-plugin-prettier 5.5.3 ➙ 5.5.4
            - plotly.js-basic-dist 3.0.3 ➙ 3.1.0
            - sass-embedded 1.89.2 ➙ 1.90.0
            - vite 7.0.6 ➙ 7.1.3
            - vite-plugin-static-copy 3.1.1 ➙ 3.1.2
        * uv 0.7.19 ➙ 0.8.12

- 0.8.5-1 (2025-08-18):
    + Updated dependencies.
        * DuckDB 1.2.2 ➙ 1.3.1
        * Go 1.24.2 ➙ 1.24.4
            - gin 1.10.0 ➙ 1.10.1
            - mcp-go 0.28.0 ➙ 0.33.0
        * Node.js 22.15.0 ➙ 24.3.0
            - @eslint/js 9.26.0 ➙ 9.30.1
            - @types/react 19.1.4 ➙ 19.1.8
            - @types/react-dom 19.1.3 ➙ 19.1.6
            - @vitejs/plugin-react-swc 3.9.0 ➙ 3.10.2
            - air-datepicker 3.5.3 ➙ 3.6.0
            - bootstrap 5.3.6 ➙ 5.3.7
            - eslint 9.26.0 ➙ 9.30.1
            - eslint-plugin-prettier 5.4.0 ➙ 5.5.1
            - globals 16.0.0 ➙ 16.3.0
            - prettier 3.5.3 ➙ 3.6.2
            - sass-embedded 1.89.0 ➙ 1.89.2
            - vite 6.3.5 ➙ 7.0.3
            - vite-plugin-static-copy 3.0.0 ➙ 3.1.0
        * uv 0.7.2 ➙ 0.7.19
    + Fixed 'fpm' installation in RHEL8.

- 0.8.4-1 (2025-05-19):
    + Updated dependencies.
        * DuckDB 1.2.1 ➙ 1.2.2
        * FPM 1.15.1 ➙ 1.16.0
        * Go 1.24.1 ➙ 1.24.2
            - golangci-lint 2.0.2 ➙ 2.1.6
            - mcp-go 0.22.0 ➙ 0.28.0
            - sys 0.32.0 ➙ 0.33.0
        * Node.js 22.14.0 ➙ 22.15.0
            - @eslint/js 9.25.1 ➙ 9.26.0
            - @types/react 19.0.12 ➙ 19.1.4
            - @types/react-dom 19.0.4 ➙ 19.1.3
            - @vitejs/plugin-react-swc 3.8.1 ➙ 3.9.0
            - bootstrap 5.3.5 ➙ 5.3.6
            - eslint 9.23.0 ➙ 9.26.0
            - eslint-config-prettier 10.1.2 ➙ 10.1.5
            - eslint-plugin-prettier 5.2.6 ➙ 5.4.0
            - eslint-plugin-react 7.37.4 ➙ 7.37.5
            - eslint-plugin-react-refresh 0.4.19 ➙ 0.4.20
            - sass-embedded 1.86.3 ➙ 1.89.0
            - vite 6.3.2 ➙ 6.3.5
            - vite-plugin-static-copy 2.3.1 ➙ 3.0.0
        * uv 0.6.0 ➙ 0.7.2
    + Fixed filter history UI.

- 0.8.3-1 (2025-04-23):
    + Enabled MCP keep-alive mechanism to enhance connection stability.

- 0.8.2-1 (2025-04-23):
    + Added pagination support to the MCP tool used to collect metrics.
    + Fixed response of the MCP config tool.

- 0.8.1-1 (2025-04-23):
    + Fixed double response compression in the '/metrics' endpoint.

- 0.8.0-1 (2025-04-22):
    + Added basic support for MCP (Model Context Protocol) in the API, enabling queries to varnishmon using LLMs.
    + Updated dependencies.
        * Go
          - gin-contrib/pprof 1.5.2 ➙ 1.5.3
          - prometheus/client_golang 1.21.1 ➙ 1.22.0
        * Node.js
          - @eslint/js 9.23.0 ➙ 9.25.1
          - bootstrap 5.3.3 ➙ 5.3.5
          - eslint-config-prettier 10.1.1 ➙ 10.1.2
          - vite 6.2.5 ➙ 6.3.2
          - vite-plugin-static-copy 2.3.0 ➙ 2.3.1

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
