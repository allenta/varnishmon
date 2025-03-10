import PropTypes from 'prop-types';
import React from 'react';

import { Cluster } from './Cluster';

export function Clusters({ filteredMetrics, ...props }) {
  return (
    <div className="accordion accordion-flush flex-grow-1 d-flex flex-column">
      {filteredMetrics != null &&
        filteredMetrics.clusters.map((cluster) => (
          <Cluster key={cluster.name} {...props} cluster={cluster} />
        ))}
    </div>
  );
}

Clusters.propTypes = {
  timeRangePicker: PropTypes.object.isRequired,
  initialRange: PropTypes.object.isRequired,
  refreshInterval: PropTypes.number.isRequired,
  columns: PropTypes.number.isRequired,
  aggregator: PropTypes.string.isRequired,
  step: PropTypes.number.isRequired,
  filteredMetrics: PropTypes.object,
};
