import PropTypes from 'prop-types';

export function FilterStats({ filterStats }) {
  return (
    <div className="col align-content-center text-muted">{filterStats}</div>
  );
}

FilterStats.propTypes = {
  filterStats: PropTypes.string.isRequired,
};
