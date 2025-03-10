import PropTypes from 'prop-types';
import React from 'react';

import * as config from '../config';
import * as helpers from '../helpers';
import { TimeRangePicker } from '../time-picker';

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
