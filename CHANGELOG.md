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
