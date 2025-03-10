import PropTypes from 'prop-types';
import React from 'react';

import * as config from '../../config';

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
