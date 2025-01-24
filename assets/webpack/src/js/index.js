import '../scss/main.scss';
import Collapse from 'bootstrap/js/dist/collapse';

import * as config from './config';
import * as helpers from './helpers';
import * as storage from './storage';
import Chart from './chart';
import { TimeRangePicker } from './time-picker';

function syncConfigWithUI() {
  function populateSelect(selector, values, selectedValue) {
    values.forEach(value => {
      const option = document.createElement('option');
      if (Array.isArray(value) && value.length === 2) {
        option.value = value[0];
        option.text = value[1];
      } else {
        option.value = value;
        option.text = value;
      }
      selector.appendChild(option);
    });
    selector.value = selectedValue;
  }

  // Time range. Beware the time range is synchronized to the config only when
  // the user clicks the apply button and the selected range is valid.
  const range = document.getElementById('range').timeRangePicker;
  try {
    range.setDates(...config.getTimeRange());
  } catch {
    // If whatever comes from the local storage is invalid, skip it and use
    // default values.
    range.setDates(...config.getTimeRange(true));
  }

  // Refresh interval.
  const refreshInterval = document.getElementById('refresh-interval');
  populateSelect(refreshInterval, config.getRefreshIntervalValues(), config.getRefreshInterval());
  refreshInterval.addEventListener('change', (event) => {
    config.setRefreshInterval(parseInt(event.target.value, 10));
  });

  // Filter.
  const filterSelector = document.getElementById('filter');
  filterSelector.value = config.getFilter();
  filterSelector.addEventListener('change', (event) => {
    config.setFilter(event.target.value);
  });

  // Verbosity.
  const verbositySelector = document.getElementById('verbosity');
  populateSelect(verbositySelector, config.getVerbosityValues(), config.getVerbosity());
  verbositySelector.addEventListener('change', (event) => {
    config.setVerbosity(event.target.value);
  });

  // Columns.
  const columnsSelector = document.getElementById('columns');
  populateSelect(columnsSelector, config.getColumnsValues(), config.getColumns());
  columnsSelector.addEventListener('change', (event) => {
    config.setColumns(parseInt(event.target.value, 10));
  });

  // Aggregator.
  const aggregatorSelector = document.getElementById('aggregator');
  populateSelect(aggregatorSelector, config.getAggregatorValues(), config.getAggregator());
  aggregatorSelector.addEventListener('change', (event) => {
    config.setAggregator(event.target.value);
  });

  // Step.
  const stepSelector = document.getElementById('step');
  stepSelector.min = varnishmon.config.scraper.period;
  stepSelector.value = config.getStep();
  stepSelector.addEventListener('change', (event) => {
    const value = parseInt(event.target.value, 10);
    if (value >= varnishmon.config.scraper.period) {
      config.setStep(value);
    } else {
      event.stopPropagation();
      stepSelector.value = varnishmon.config.scraper.period;
      helpers.notify(
        'error',
        `Step must be at least ${varnishmon.config.scraper.period} seconds, ` +
        'which is the metrics scraping period');
    }
  });
}

function getRefreshInterval() {
  let value = parseInt(document.getElementById('refresh-interval').value, 10);
  if (value < 0) {
    value = varnishmon.config.scraper.period;
  }
  return value;
}

function getStep() {
  let value = parseInt(document.getElementById('step').value, 10);
  if (value < varnishmon.config.scraper.period) {
    value = varnishmon.config.scraper.period;
  }
  return value;
}

function setUpEventListeners() {
  // On click in the apply time range button, the search results must be
  // rebuilt from scratch because a different time range might lead to a
  // different set of metrics.
  document.getElementById('apply-time-range').addEventListener('click', (event) => {
    // Validate the selected time range.
    const range = document.getElementById('range').timeRangePicker;
    if (!range.hasValidDates()) {
      event.stopPropagation();
      helpers.notify(
        'error',
        'The selected time range is invalid. ISO 8601 and relative expressions' +
        ' like \'now-1h\', \'now\', \'now-1d\', etc. are allowed.');
      return;
    }

    // Update the config with the selected time range using the raw dates.
    config.setTimeRange(...range.getRawDates());

    // Reload the metrics using the new time range.
    reloadMetrics();
  });

  // On change in the refresh interval, report the new value to all the charts.
  document.getElementById('refresh-interval').addEventListener('change', () => {
    let value = getRefreshInterval();
    document.getElementById('clusters').querySelectorAll('.chart').forEach((chartDiv) => {
      chartDiv.chart.setRefreshInterval(value);
    });
  });

  // On click in the refresh button, request all the charts to refresh asap.
  document.getElementById('refresh').addEventListener('click', () => {
    document.getElementById('clusters').querySelectorAll('.chart').forEach((chartDiv) => {
      chartDiv.chart.refresh();
    });
  });

  // On change in the filter, verbosity or columns widgets, update the search
  // results accordingly. This is a lightweight operation, as it only adjusts
  // the visibility and arranging of the charts and clusters already fetched.
  document.getElementById('filter').addEventListener('input', updateSearchResults);
  document.getElementById('verbosity').addEventListener('change', updateSearchResults);
  document.getElementById('columns').addEventListener('change', updateSearchResults);

  // On change in the aggregator, report the new value to all the charts.
  document.getElementById('aggregator').addEventListener('change', (event) => {
    document.getElementById('clusters').querySelectorAll('.chart').forEach((chartDiv) => {
      chartDiv.chart.setAggregator(event.target.value);
    });
  });

  // On change in the step, report the new value to all the charts.
  document.getElementById('step').addEventListener('change', () => {
    const value = getStep();
    document.getElementById('clusters').querySelectorAll('.chart').forEach((chartDiv) => {
      chartDiv.chart.setStep(value);
    });
  });

  // On click in the reset button, reset the config in the local storage and
  // reload the page.
  document.getElementById('reset').addEventListener('click', () => {
    config.reset();
    location.reload();
  });

  // On click in the collapse-all button, collapse all clusters.
  document.getElementById('collapse-all').addEventListener('click', () => {
    document.getElementById('clusters').querySelectorAll('.cluster').forEach(cluster => {
      Collapse.getInstance(cluster.querySelector('.accordion-collapse')).hide();
      cluster.querySelector('.accordion-button').classList.add('collapsed');
    });
  });

  // On click in the expand-all button, expand all clusters.
  document.getElementById('expand-all').addEventListener('click', () => {
    document.getElementById('clusters').querySelectorAll('.cluster').forEach(cluster => {
      Collapse.getInstance(cluster.querySelector('.accordion-collapse')).show();
      cluster.querySelector('.accordion-button').classList.remove('collapsed');
    });
  });
}

async function reloadMetrics() {
  // Discard any previous search results and show a spinner while fetching the
  // new metrics.
  const clustersSelector = document.getElementById('clusters');
  clustersSelector.querySelectorAll('.chart').forEach((chartDiv) => {
    chartDiv.chart.destroy();
  });
  clustersSelector.innerHTML = '';
  clustersSelector.appendChild(document.getElementById('spinner-template').
    content.cloneNode(true).firstElementChild);

  // Fetch values from some widgets.
  const rangeFactory = document.getElementById('range').timeRangePicker.getDatesFactory();
  const refreshInterval = getRefreshInterval();
  const aggregator = document.getElementById('aggregator').value;
  const step = getStep();

  // Fetch metrics from the storage.
  let metrics;
  try {
    const [from, to] = rangeFactory();
    metrics = await storage.getMetrics(from, to, step);
  } catch (error) {
    clustersSelector.innerHTML = '';
    clustersSelector.appendChild(document.getElementById('metrics-meditation-template').
      content.cloneNode(true).firstElementChild);
    helpers.notify('error', `Failed to fetch metrics: ${error}`);
    return;
  }

  // Notify the user about the number of metrics and clusters fetched.
  const numClusters = metrics.clusters.length;
  const numMetrics = metrics.clusters.reduce((acc, cluster) => acc + cluster.metrics.length, 0);
  helpers.notify('info', `Fetched ${numMetrics} metrics organized in ${numClusters} clusters`);

  // Build the search results from the fetched metrics.
  const clusterTemplateSelector = document.getElementById('cluster-template');
  const chartTemplateSelector = document.getElementById('chart-template');
  clustersSelector.innerHTML = '';
  metrics.clusters.forEach((cluster) => {
    // Build cluster, make it collapsible and append it to the container.
    const clusterDiv = clusterTemplateSelector.content.cloneNode(true).firstElementChild;
    clusterDiv.querySelector('.cluster-name').textContent = cluster.name;
    clustersSelector.appendChild(clusterDiv);
    const clusterCollapsable = new Collapse(clusterDiv.querySelector('.accordion-collapse'));
    clusterDiv.querySelector('.accordion-button').addEventListener('click', (event) => {
      event.currentTarget.classList.toggle('collapsed');
      clusterCollapsable.toggle();
    });

    // Build charts and append them to the cluster.
    const chartsDiv = clusterDiv.querySelector('.charts');
    cluster.metrics.forEach(metric => {
      const chartDiv = chartTemplateSelector.content.cloneNode(true).firstElementChild;
      const chart = new Chart(chartDiv, metric, rangeFactory, refreshInterval, aggregator, step);
      chartDiv.chart = chart;
      chartsDiv.appendChild(chartDiv);
    });
  });

  // Adjust visibility of charts and clusters according to the filtering
  // criteria, the verbosity and the number of columns available.
  updateSearchResults();
}

function updateSearchResults() {
  // Adjust charts according to the filtering criteria, the verbosity and the
  // number of columns available.
  const clustersSelector = document.getElementById('clusters');
  clustersSelector.querySelectorAll('.chart').forEach((chartDiv) => {
    chartDiv.chart.redraw(
      document.getElementById('filter').value,
      document.getElementById('verbosity').value,
      parseInt(document.getElementById('columns').value, 10));
  });

  // Once visibility of charts is adjusted, adjust visibility of clusters: hide
  // clusters with no visible charts, and show clusters with at least one
  // chart visible.
  clustersSelector.querySelectorAll('.cluster').forEach((cluster) => {
    const chartSelectors = cluster.querySelectorAll('.chart:not(.d-none)');
    if (chartSelectors.length === 0) {
      cluster.classList.add('d-none');
    } else {
      cluster.classList.remove('d-none');
    }
  });

  // Once both charts and clusters visibility are adjusted, update the filter
  // stats accordingly.
  const numClusters = clustersSelector.querySelectorAll('.cluster').length;
  const numVisibleClusters = clustersSelector.querySelectorAll('.cluster:not(.d-none)').length;
  const numMetrics = clustersSelector.querySelectorAll('.chart').length;
  const numVisibleMetrics = clustersSelector.querySelectorAll('.chart:not(.d-none)').length;
  document.getElementById('filter-stats').textContent =
    `${numVisibleMetrics} metrics found (${numMetrics-numVisibleMetrics} hidden),` +
    ` organized in ${numVisibleClusters} clusters (${numClusters-numVisibleClusters}` +
    ' hidden)';
}

/******************************************************************************
 * MAIN.
 ******************************************************************************/

document.addEventListener('DOMContentLoaded', async () => {
  // Override Plotly notifications system to use our own.
  helpers.overridePlotlyNotificationsSystem();

  // Set up time pickers.
  document.getElementById('range').timeRangePicker = new TimeRangePicker(
    document.getElementById('range-from'),
    document.getElementById('range-to'));

  // Initialize & keep input widgets in sync with the config in the local
  // storage.
  syncConfigWithUI();

  // Set up event listeners for all the widgets.
  setUpEventListeners();

  // Load metrics.
  reloadMetrics();
});
