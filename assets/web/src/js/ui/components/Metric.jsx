import PropTypes from 'prop-types';
import React from 'react';

import Chart from '../chart';

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
