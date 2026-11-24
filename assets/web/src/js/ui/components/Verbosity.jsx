import React from 'react';

import * as config from '../../config';

/**
 * @param {object} props - Component props.
 * @param {string} props.verbosity
 * @param {Function} props.setVerbosity
 */
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
