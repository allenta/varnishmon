import PropTypes from 'prop-types';
import React from 'react';

export function FilterStats({ filterStats }) {
  return (
    <div className="col align-content-center text-muted">{filterStats}</div>
  );
}

FilterStats.propTypes = {
  filterStats: PropTypes.string.isRequired,
};
