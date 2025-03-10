import PropTypes from 'prop-types';
import React from 'react';

import * as config from '../../config';
import * as helpers from '../../helpers';

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
