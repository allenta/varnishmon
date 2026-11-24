import React from 'react';

import * as config from '../../config';

/**
 * @param {object} props - Component props.
 * @param {string} props.aggregator
 * @param {Function} props.setAggregator
 */
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
