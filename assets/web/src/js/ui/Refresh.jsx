import PropTypes from 'prop-types';
import React from 'react';

import * as config from '../config';

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
