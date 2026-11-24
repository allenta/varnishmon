import React from 'react';

import * as config from '../../config';

/**
 * @param {object} props - Component props.
 * @param {number} props.columns
 * @param {Function} props.setColumns
 */
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
