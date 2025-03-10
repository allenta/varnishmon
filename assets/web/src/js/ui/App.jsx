import React from 'react';

import * as config from '../config';
import * as helpers from '../helpers';

import { useReload } from './useReload';
import { useFilter } from './useFilter';
import { Host } from './components/Host';
import { TimeRange } from './components/TimeRange';
import { Refresh } from './components/Refresh';
import { Filter } from './components/Filter';
import { Verbosity } from './components/Verbosity';
import { Columns } from './components/Columns';
import { Aggregator } from './components/Aggregator';
import { Step } from './components/Step';
import { FilterStats } from './components/FilterStats';
import { Actions } from './components/Actions';
import { Clusters } from './components/Clusters';

export function App() {
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
            <Host />
            <TimeRange
              timeRangePicker={timeRangePicker}
              setTimeRangePicker={setTimeRangePicker}
              reload={reload}
            />
            <Refresh
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
              <Filter filter={filter} setFilter={setFilter} />
            </div>
            <div className="col-md-1">
              <Verbosity verbosity={verbosity} setVerbosity={setVerbosity} />
            </div>
            <div className="col-md-1">
              <Columns columns={columns} setColumns={setColumns} />
            </div>
            <div className="col-md-1">
              <Aggregator
                aggregator={aggregator}
                setAggregator={setAggregator}
              />
            </div>
            <div className="col-md-1">
              <Step step={step} setStep={setStep} />
            </div>
          </div>

          <div className="row mb-2">
            <FilterStats filterStats={filterStats} />
            <div className="col text-end">
              <Actions />
            </div>
          </div>

          {loadProgress == null && (
            <Clusters
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
}
