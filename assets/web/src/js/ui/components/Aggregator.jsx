import PropTypes from 'prop-types';
import React from 'react';

import * as config from '../../config';

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
