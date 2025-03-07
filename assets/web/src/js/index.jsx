import React from 'react';
import { createRoot } from 'react-dom/client';

import * as config from './config';
import * as helpers from './helpers';
import * as storage from './storage';
import * as components from './components';

import '../scss/main.scss';

const useReload = ({ timeRangePicker, initialRange, step }) => {
  const [metrics, setMetrics] = React.useState(null);
  const [loadProgress, setLoadProgress] = React.useState(null);

  const reload = React.useCallback(() => {
    // Flush previous state.
    initialRange.current = null;
    setMetrics(null);

    // Fetch metrics from the storage, exposing load state to show a loading
    // spinner in the meantime, etc.
    setLoadProgress('loading');
    const rangeFactory = timeRangePicker.getDatesFactory();
    const [from, to] = rangeFactory();
    storage
      .getMetrics(from, to, step)
      .then((fetchedMetrics) => {
        const numClusters = fetchedMetrics.clusters.length;
        const numMetrics = fetchedMetrics.clusters.reduce(
          (acc, cluster) => acc + cluster.metrics.length,
          0,
        );
        helpers.notify(
          'info',
          `Fetched ${numMetrics} metrics organized in ${numClusters} clusters`,
        );
        setLoadProgress(null);

        setMetrics(fetchedMetrics);
      })
      .catch((error) => {
        helpers.notify('error', `Failed to fetch metrics: ${error}`);
        setLoadProgress('error');
      });
  }, [timeRangePicker, initialRange, step]);

  return [metrics, loadProgress, reload];
};

const useFilter = ({ metrics }) => {
  const [filter, setFilter] = React.useState(config.getFilter());
  const [filterStats, setFilterStats] = React.useState('');
  const [verbosity, setVerbosity] = React.useState(config.getVerbosity());
  const [filteredMetrics, setFilteredMetrics] = React.useState(metrics);

  React.useEffect(() => {
    if (metrics == null) {
      setFilterStats('');
      setFilteredMetrics(null);
      return;
    }

    const newFilteredMetrics = { ...metrics, clusters: [] };
    const filterTerms = filter.split(/\s+/).filter((term) => term.length > 0);
    let numClusters = 0,
      numVisibleClusters = 0,
      numMetrics = 0,
      numVisibleMetrics = 0;
    metrics.clusters.forEach((cluster) => {
      numClusters += 1;
      numMetrics += cluster.metrics.length;

      const newCluster = { ...cluster, metrics: [] };
      newCluster.metrics = cluster.metrics.filter((metric) => {
        if (verbosity === 'normal' && metric.debug) {
          return false;
        }

        if (filterTerms.length > 0) {
          return filterTerms.some((term) => metric.name.includes(term));
        }

        return true;
      });

      if (newCluster.metrics.length > 0) {
        numVisibleClusters += 1;
        numVisibleMetrics += newCluster.metrics.length;
        newFilteredMetrics.clusters.push(newCluster);
      }
    });

    setFilterStats(
      `${numVisibleMetrics} metrics found (${numMetrics - numVisibleMetrics} hidden),` +
        ` organized in ${numVisibleClusters} clusters (${numClusters - numVisibleClusters}` +
        ' hidden)',
    );
    setFilteredMetrics(newFilteredMetrics);
  }, [metrics, filter, verbosity]);

  return [
    filteredMetrics,
    filter,
    setFilter,
    filterStats,
    verbosity,
    setVerbosity,
  ];
};

const App = () => {
  const [timeRangePicker, setTimeRangePicker] = React.useState(null);
  const [refreshInterval, setRefreshInterval] = React.useState(
    config.getRefreshInterval(),
  );
  const [columns, setColumns] = React.useState(config.getColumns());
  const [aggregator, setAggregator] = React.useState(config.getAggregator());
  const [step, setStep] = React.useState(config.getStep());

  const initialRange = React.useRef(null);
  const initialLoadDone = React.useRef(false);

  const [metrics, loadProgress, reload] = useReload({
    timeRangePicker,
    initialRange,
    step,
  });
  const [
    filteredMetrics,
    filter,
    setFilter,
    filterStats,
    verbosity,
    setVerbosity,
  ] = useFilter({ metrics });

  // Override Plotly notifications system to use our own. This will be executed
  // only once, when the component is mounted.
  React.useEffect(() => {
    helpers.overridePlotlyNotificationsSystem();
  }, []);

  // Execute the initial load of metrics when all required parameters are set.
  // This will trigger the load only once, when the component is mounted and all
  // the required parameters are set.
  React.useEffect(() => {
    if (
      !initialLoadDone.current &&
      timeRangePicker != null &&
      aggregator != null &&
      step != null
    ) {
      reload();
      initialLoadDone.current = true;
    }
  }, [initialLoadDone, timeRangePicker, aggregator, step, reload]);

  return (
    <>
      <nav className="navbar navbar-expand-lg navbar-dark bg-dark sticky-top">
        <div className="container-fluid">
          <a className="navbar-brand" href="/">
            varnishmon
          </a>
          <div className="d-flex ms-auto">
            <components.Host />
            <components.TimeRange
              timeRangePicker={timeRangePicker}
              setTimeRangePicker={setTimeRangePicker}
              reload={reload}
            />
            <components.Refresh
              refreshInterval={refreshInterval}
              setRefreshInterval={setRefreshInterval}
            />
          </div>
        </div>
      </nav>

      <main className="flex-grow-1 d-flex flex-column">
        <div className="container-fluid py-md-4 flex-grow-1 d-flex flex-column">
          <div className="row mb-2">
            <div className="col-md-8">
              <components.Filter filter={filter} setFilter={setFilter} />
            </div>
            <div className="col-md-1">
              <components.Verbosity
                verbosity={verbosity}
                setVerbosity={setVerbosity}
              />
            </div>
            <div className="col-md-1">
              <components.Columns columns={columns} setColumns={setColumns} />
            </div>
            <div className="col-md-1">
              <components.Aggregator
                aggregator={aggregator}
                setAggregator={setAggregator}
              />
            </div>
            <div className="col-md-1">
              <components.Step step={step} setStep={setStep} />
            </div>
          </div>

          <div className="row mb-2">
            <components.FilterStats filterStats={filterStats} />
            <div className="col text-end">
              <components.Actions />
            </div>
          </div>

          {loadProgress == null && (
            <components.Clusters
              timeRangePicker={timeRangePicker}
              initialRange={initialRange}
              refreshInterval={refreshInterval}
              columns={columns}
              aggregator={aggregator}
              step={step}
              filteredMetrics={filteredMetrics}
            />
          )}
          {loadProgress === 'loading' && (
            <div className="d-flex justify-content-center flex-grow-1 align-items-center">
              <div className="spinner-border fs-2 opacity-50" role="status">
                <span className="visually-hidden">Loading...</span>
              </div>
            </div>
          )}
          {loadProgress === 'error' && (
            <div className="d-flex flex-column text-center justify-content-center flex-grow-1">
              <h2 className="mt-4">
                <i className="fa-regular fa-face-sad-tear fa-3x"></i>
              </h2>
              <h2 className="mt-2">Metrics Meditation</h2>
              <p className="mt-4 text-muted fs-5 w-25 mx-auto">
                Oops! Something went wrong while fetching metrics. Please, make
                sure
                <span className="font-monospace">varnishmon</span> is up and
                reachable
              </p>
            </div>
          )}
        </div>
      </main>
    </>
  );
};

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
