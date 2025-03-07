import PropTypes from 'prop-types';
import React from 'react';
import Collapse from 'bootstrap/js/dist/collapse';

import * as config from './config';
import * as helpers from './helpers';
import { TimeRangePicker } from './time-picker';
import Chart from './chart';

/******************************************************************************
 * Host.
 ******************************************************************************/

export function Host() {
  return (
    <div className="me-4 align-self-center">
      <span className="navbar-text font-monospace text-white">
        <i className="fa-solid fa-computer"></i> {varnishmon.storage.hostname}
      </span>
    </div>
  );
}

/******************************************************************************
 * TimeRange.
 ******************************************************************************/

export function TimeRange({ timeRangePicker, setTimeRangePicker, reload }) {
  const containerRef = React.useRef(null);

  React.useEffect(() => {
    if (containerRef.current != null) {
      // Beware the time range is synchronized to the config only when the user
      // clicks the apply button and the selected range is valid.
      const from = containerRef.current.querySelector('input.time-range-from');
      const to = containerRef.current.querySelector('input.time-range-to');
      const newTimeRangePicker = new TimeRangePicker(from, to);
      try {
        newTimeRangePicker.setDates(...config.getTimeRange());
      } catch {
        // If whatever comes from the local storage is invalid, skip it and use
        // default values.
        newTimeRangePicker.setDates(...config.getTimeRange(true));
      }
      setTimeRangePicker(newTimeRangePicker);

      return () => {
        newTimeRangePicker.destroy();
      };
    }
  }, [containerRef, setTimeRangePicker]);

  // On click in the apply time range button, the search results must be rebuilt
  // from scratch because a different time range might lead to a different set
  // of metrics.
  const onClick = () => {
    // Validate the selected time range.
    if (!timeRangePicker.hasValidDates()) {
      helpers.notify(
        'error',
        'The selected time range is invalid. ISO 8601 and relative expressions' +
          " like 'now-1h', 'now', 'now-1d', etc. are allowed.",
      );
      return;
    }

    // Update the config with the selected time range using the raw dates.
    config.setTimeRange(...timeRangePicker.getRawDates());

    // Reload the metrics using the new time range.
    reload();
  };

  const onKeyDown = (event) => {
    if (event.key === 'Enter') {
      onClick();
    }
  };

  return (
    <div ref={containerRef} style={{ display: 'contents' }}>
      <div className="me-2">
        <div className="input-group">
          <span className="input-group-text">
            <i className="fas fa-calendar-alt"></i>
          </span>
          <input
            type="text"
            className="form-control time-range-from"
            placeholder="from"
            onKeyDown={onKeyDown}
          />
        </div>
      </div>
      <div className="me-2 align-self-center text-light">
        <i className="fa-solid fa-arrow-right"></i>
      </div>
      <div className="me-2">
        <div className="input-group">
          <span className="input-group-text">
            <i className="fas fa-calendar-alt"></i>
          </span>
          <input
            type="text"
            className="form-control time-range-to"
            placeholder="to"
            onKeyDown={onKeyDown}
          />
        </div>
      </div>
      <div className="me-4 align-self-center align-self-end">
        <button
          className="btn btn-primary"
          title="Apply the selected time range"
          onClick={onClick}
        >
          <i className="fa-solid fa-play"></i>
        </button>
      </div>
    </div>
  );
}

TimeRange.propTypes = {
  timeRangePicker: PropTypes.object,
  setTimeRangePicker: PropTypes.func.isRequired,
  reload: PropTypes.func.isRequired,
};

/******************************************************************************
 * Refresh.
 ******************************************************************************/

export function Refresh({ refreshInterval, setRefreshInterval }) {
  const onChange = (event) => {
    const value = parseInt(event.target.value, 10);
    config.setRefreshInterval(value);
    setRefreshInterval(value);
  };

  const onClick = () => {
    document.dispatchEvent(new Event('onRefresh'));
  };

  return (
    <>
      <div className="me-2">
        <select
          className="form-select"
          value={refreshInterval}
          onChange={onChange}
        >
          {config.getRefreshIntervalValues().map((value) => (
            <option key={value[0]} value={value[0]}>
              {value[1]}
            </option>
          ))}
        </select>
      </div>
      <div className="align-self-center align-self-end">
        <button
          className="btn btn-primary"
          title="Trigger refresh now"
          onClick={onClick}
        >
          <i className="fa-solid fa-sync"></i>
        </button>
      </div>
    </>
  );
}

Refresh.propTypes = {
  refreshInterval: PropTypes.number.isRequired,
  setRefreshInterval: PropTypes.func.isRequired,
};

/******************************************************************************
 * Filter.
 ******************************************************************************/

export function Filter({ filter, setFilter }) {
  const [localFilter, setLocalFilter] = React.useState(filter);
  const [history, setHistory] = React.useState(config.getFilterHistory());
  const filterId = React.useId();

  const debouncedSetFilter = React.useCallback(
    (value) => {
      return helpers.debounce(setFilter, 500)(value);
    },
    [setFilter],
  );

  const updateLocalState = (newFilter, updateHistory) => {
    config.setFilter(newFilter);
    setLocalFilter(newFilter);
    debouncedSetFilter(newFilter);

    if (updateHistory && newFilter) {
      const newHistory = [...history];
      const index = newHistory.indexOf(newFilter);
      if (index !== 0) {
        if (index !== -1) {
          newHistory.splice(index, 1);
        }
        newHistory.unshift(newFilter);
        if (newHistory.length > 10) {
          newHistory.pop();
        }
        config.setFilterHistory(newHistory);
        setHistory(newHistory);
      }
    }
  };

  const onChange = (event) => {
    updateLocalState(event.target.value, false);
  };

  const onBlur = (event) => {
    updateLocalState(event.target.value, true);
  };

  const onHistoryClick = (event) => {
    updateLocalState(event.target.textContent, true);
  };

  return (
    <>
      <label className="form-label" htmlFor={filterId}>
        Filter
      </label>
      <div className="input-group">
        <span className="input-group-text">
          <i className="fa-solid fa-magnifying-glass"></i>
        </span>
        <input
          type="text"
          className="form-control"
          placeholder="type here to filter metrics by name"
          onBlur={onBlur}
          onChange={onChange}
          value={localFilter}
          id={filterId}
        />
        <button
          className="btn border-secondary-subtle bg-body-tertiary dropdown-toggle"
          type="button"
          data-bs-toggle="dropdown"
          aria-expanded="false"
        ></button>
        <ul
          className="dropdown-menu dropdown-menu-end w-100"
          aria-labelledby="historyDropdown"
        >
          {history.map((item) => (
            <li key={item} className="dropdown-item" onClick={onHistoryClick}>
              {item}
            </li>
          ))}
        </ul>
      </div>
    </>
  );
}

Filter.propTypes = {
  filter: PropTypes.string.isRequired,
  setFilter: PropTypes.func.isRequired,
};

/******************************************************************************
 * Verbosity.
 ******************************************************************************/

export function Verbosity({ verbosity, setVerbosity }) {
  const verbosityId = React.useId();

  const onChange = (event) => {
    const value = event.target.value;
    config.setVerbosity(value);
    setVerbosity(value);
  };

  return (
    <>
      <label className="form-label" htmlFor={verbosityId}>
        Verbosity
      </label>
      <div className="input-group">
        <span className="input-group-text">
          <i className="fa-regular fa-comments"></i>
        </span>
        <select
          className="form-select"
          value={verbosity}
          onChange={onChange}
          id={verbosityId}
        >
          {config.getVerbosityValues().map((value) => (
            <option key={value} value={value}>
              {value}
            </option>
          ))}
        </select>
      </div>
    </>
  );
}

Verbosity.propTypes = {
  verbosity: PropTypes.string.isRequired,
  setVerbosity: PropTypes.func.isRequired,
};

/******************************************************************************
 * Columns.
 ******************************************************************************/

export function Columns({ columns, setColumns }) {
  const columnsId = React.useId();

  const onChange = (event) => {
    const value = parseInt(event.target.value, 10);
    config.setColumns(value);
    setColumns(value);
  };

  return (
    <>
      <label className="form-label" htmlFor={columnsId}>
        Columns
      </label>
      <div className="input-group">
        <span className="input-group-text">
          <i className="fa-solid fa-table-cells-large"></i>
        </span>
        <select
          className="form-select"
          value={columns}
          onChange={onChange}
          id={columnsId}
        >
          {config.getColumnsValues().map((value) => (
            <option key={value} value={value}>
              {value}
            </option>
          ))}
        </select>
      </div>
    </>
  );
}

Columns.propTypes = {
  columns: PropTypes.number.isRequired,
  setColumns: PropTypes.func.isRequired,
};

/******************************************************************************
 * Aggregator.
 ******************************************************************************/

export function Aggregator({ aggregator, setAggregator }) {
  const aggregatorId = React.useId();

  const onChange = (event) => {
    const value = event.target.value;
    config.setAggregator(value);
    setAggregator(value);
  };

  return (
    <>
      <label className="form-label" htmlFor={aggregatorId}>
        Aggregator
      </label>
      <div className="input-group">
        <span className="input-group-text">
          <i className="fa-solid fa-filter"></i>
        </span>
        <select
          className="form-select"
          value={aggregator}
          onChange={onChange}
          id={aggregatorId}
        >
          {config.getAggregatorValues().map((value) => (
            <option key={value} value={value}>
              {value}
            </option>
          ))}
        </select>
      </div>
    </>
  );
}

Aggregator.propTypes = {
  aggregator: PropTypes.string.isRequired,
  setAggregator: PropTypes.func.isRequired,
};

/******************************************************************************
 * Step.
 ******************************************************************************/

export function Step({ step, setStep }) {
  const stepId = React.useId();
  const [localStep, setLocalStep] = React.useState(step);

  const onChange = (event) => {
    const value = event.target.value;
    setLocalStep(value);
  };

  const onBlur = (event) => {
    let value = event.target.value;
    const minimum = config.getMinimumStep();

    if (value === '' || parseInt(value, 10) < minimum) {
      helpers.notify('error', `Step must be at least ${minimum} seconds`);
      value = minimum;
    } else {
      value = parseInt(value, 10);
    }

    config.setStep(value);
    setLocalStep(value);
    setStep(value);
  };

  return (
    <>
      <label className="form-label" htmlFor={stepId}>
        Step
      </label>
      <div className="input-group">
        <span className="input-group-text">
          <i className="fa-solid fa-arrows-left-right-to-line"></i>
        </span>
        <input
          type="number"
          className="form-control"
          id={stepId}
          min={config.getMinimumStep()}
          value={localStep}
          onChange={onChange}
          onBlur={onBlur}
        />
      </div>
    </>
  );
}

Step.propTypes = {
  step: PropTypes.number.isRequired,
  setStep: PropTypes.func.isRequired,
};

/******************************************************************************
 * FilterStats.
 ******************************************************************************/

export function FilterStats({ filterStats }) {
  return (
    <div className="col align-content-center text-muted">{filterStats}</div>
  );
}

FilterStats.propTypes = {
  filterStats: PropTypes.string.isRequired,
};

/******************************************************************************
 * Actions.
 ******************************************************************************/

export function Actions() {
  const onResetClick = () => {
    config.reset();
    location.reload();
  };

  const onCollapseClick = () => {
    document.dispatchEvent(new Event('onCollapseAllClusters'));
  };

  const onExpandClick = () => {
    document.dispatchEvent(new Event('onExpandAllClusters'));
  };

  return (
    <>
      <a
        className="btn btn-link"
        href="/metrics"
        role="button"
        title="View internal Prometheus metrics"
      >
        internal metrics
      </a>{' '}
      |
      <button
        type="button"
        className="btn btn-link"
        title="Discard saved state & reload"
        onClick={onResetClick}
      >
        reset
      </button>{' '}
      |
      <button
        type="button"
        className="btn btn-link"
        title="Collapse all clusters"
        onClick={onCollapseClick}
      >
        collapse
      </button>{' '}
      |
      <button
        type="button"
        className="btn btn-link"
        title="Expand all clusters"
        onClick={onExpandClick}
      >
        expand
      </button>
    </>
  );
}

/******************************************************************************
 * Clusters.
 ******************************************************************************/

export function Clusters({ filteredMetrics, ...props }) {
  return (
    <div className="accordion accordion-flush flex-grow-1 d-flex flex-column">
      {filteredMetrics != null &&
        filteredMetrics.clusters.map((cluster) => (
          <Cluster key={cluster.name} {...props} cluster={cluster} />
        ))}
    </div>
  );
}

Clusters.propTypes = {
  timeRangePicker: PropTypes.object.isRequired,
  initialRange: PropTypes.object.isRequired,
  refreshInterval: PropTypes.number.isRequired,
  columns: PropTypes.number.isRequired,
  aggregator: PropTypes.string.isRequired,
  step: PropTypes.number.isRequired,
  filteredMetrics: PropTypes.object,
};

/******************************************************************************
 * Cluster.
 ******************************************************************************/

export function Cluster({ cluster, ...props }) {
  const containerRef = React.useRef(null);
  const [accordion, setAccordion] = React.useState(null);
  const [isAccordionCollapsed, setIsAccordionCollapsed] = React.useState(false);

  React.useEffect(() => {
    if (containerRef.current != null) {
      const accordionCollapse = containerRef.current.querySelector(
        '.accordion-collapse',
      );
      const newAccordion = new Collapse(accordionCollapse);
      setAccordion(newAccordion);

      const onCollapseAllClustersListener = () => {
        setIsAccordionCollapsed(true);
      };
      document.addEventListener(
        'onCollapseAllClusters',
        onCollapseAllClustersListener,
      );

      const onExpandAllClustersListener = () => {
        setIsAccordionCollapsed(false);
      };
      document.addEventListener(
        'onExpandAllClusters',
        onExpandAllClustersListener,
      );

      return () => {
        document.removeEventListener(
          'onCollapseAllClusters',
          onCollapseAllClustersListener,
        );
        document.removeEventListener(
          'onExpandAllClusters',
          onExpandAllClustersListener,
        );
        // When mounting and then unmounting the component too quickly (e.g.,
        // due to fast reloads of metrics, strict mode, etc.), calling
        // 'newAccordion.dispose()' immediately after 'newAccordion' is created
        // causes errors to be logged in the console. UI transitions are
        // executed asynchronously and that doesn't play well with immediate
        // disposal of the instance. A hardcoded delay is used here to mitigate
        // this issue.
        setTimeout(() => newAccordion.dispose(), 1000);
      };
    }
  }, [containerRef]);

  React.useEffect(() => {
    if (accordion != null) {
      if (isAccordionCollapsed) {
        accordion.hide();
      } else {
        accordion.show();
      }
    }
  }, [accordion, isAccordionCollapsed]);

  const onClick = () => {
    setIsAccordionCollapsed(!isAccordionCollapsed);
  };

  return (
    <div ref={containerRef}>
      <div className="accordion-header">
        <button
          className={`accordion-button bg-light text-dark fs-5 border-0 font-monospace ${isAccordionCollapsed ? 'collapsed' : ''}`}
          type="button"
          onClick={onClick}
        >
          {cluster.name}
        </button>
      </div>
      <div className="accordion-collapse">
        <div className="row g-4 py-4">
          {cluster.metrics.map((metric) => (
            <Metric key={metric.id} {...props} metric={metric} />
          ))}
        </div>
      </div>
    </div>
  );
}

Cluster.propTypes = {
  timeRangePicker: PropTypes.object.isRequired,
  initialRange: PropTypes.object.isRequired,
  refreshInterval: PropTypes.number.isRequired,
  columns: PropTypes.number.isRequired,
  aggregator: PropTypes.string.isRequired,
  step: PropTypes.number.isRequired,
  cluster: PropTypes.object.isRequired,
};

/******************************************************************************
 * Metric.
 ******************************************************************************/

export function Metric({
  timeRangePicker,
  initialRange,
  refreshInterval,
  columns,
  aggregator,
  step,
  metric,
}) {
  const containerRef = React.useRef(null);
  const [chart, setChart] = React.useState(null);

  const getRefreshInterval = React.useCallback(
    () => (refreshInterval < 0 ? step : refreshInterval),
    [refreshInterval, step],
  );

  // Beware 'useLayoutEffect' is used here in order to play well with the Chart
  // class, that depends on width calculations to decide the right step factor.
  React.useLayoutEffect(() => {
    if (containerRef.current != null) {
      const rangeFactory = timeRangePicker.getDatesFactory();
      const newChart = new Chart(
        containerRef.current,
        metric,
        rangeFactory,
        getRefreshInterval(),
        aggregator,
        step,
      );
      setChart(newChart);

      newChart.addEventListener('zoom', (event) => {
        // Apply the zoom range to all the charts except the one that triggered
        // the event.
        document.dispatchEvent(
          new CustomEvent('onZoom', {
            detail: { source: newChart, range: event.range },
          }),
        );

        // Update the time range picker with the zoom range.
        if (event.range != null) {
          if (initialRange.current == null) {
            initialRange.current = timeRangePicker.getRawDates();
          }
          timeRangePicker.setDates(...event.range);
        } else {
          if (initialRange.current != null) {
            timeRangePicker.setDates(...initialRange.current);
            initialRange.current = null;
          }
        }
      });

      const onZoomListener = (event) => {
        if (event.detail.source !== newChart) {
          newChart.setZoomRange(event.detail.range);
        }
      };
      document.addEventListener('onZoom', onZoomListener);

      const onRefreshListener = () => {
        newChart.refresh();
      };
      document.addEventListener('onRefresh', onRefreshListener);

      return () => {
        document.removeEventListener('onZoom', onZoomListener);
        document.removeEventListener('onRefresh', onRefreshListener);
        newChart.destroy();
        setChart(null);
      };
    }
    // Beware 'getRefreshInterval', 'aggregator' & 'step' are intentionally
    // omitted from the dependencies array to avoid re-creating the chart when
    // those values change. Those changes are handled by the 'useEffect' hooks
    // below, which update the chart instance accordingly. Remaining dependencies
    // are include to keep the linter happy, but all of them are irrelevant.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [containerRef, setChart, timeRangePicker, initialRange, metric]);

  React.useEffect(() => {
    if (chart != null) {
      chart.setRefreshInterval(getRefreshInterval());
    }
  }, [chart, getRefreshInterval]);

  React.useEffect(() => {
    if (chart != null) {
      chart.setAggregator(aggregator);
    }
  }, [chart, aggregator]);

  React.useEffect(() => {
    if (chart != null) {
      chart.setStep(step);
    }
  }, [chart, step]);

  return (
    <div ref={containerRef} className={`chart col col-${12 / columns}`}>
      <div className="card position-relative">
        <span
          className="loading-icon spinner-grow spinner-grow-sm text-secondary position-absolute top-0 m-2 z-1 d-none"
          role="status"
        >
          <span className="visually-hidden">Loading...</span>
        </span>
        <span className="error-icon text-danger position-absolute top-0 end-0 m-2 z-1 d-none">
          <i className="fas fa-exclamation-circle"></i>
        </span>
        <div className="card-body">
          <div className="graph" style={{ height: '300px' }}></div>
        </div>
        <span
          className="step-factor text-secondary text-opacity-25 position-absolute bottom-0 end-0 me-2 mb-1 z-1 small"
          title="Effective step factor"
        ></span>
      </div>
    </div>
  );
}

Metric.propTypes = {
  timeRangePicker: PropTypes.object.isRequired,
  initialRange: PropTypes.object.isRequired,
  refreshInterval: PropTypes.number.isRequired,
  columns: PropTypes.number.isRequired,
  aggregator: PropTypes.string.isRequired,
  step: PropTypes.number.isRequired,
  metric: PropTypes.object.isRequired,
};
