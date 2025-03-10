import React from 'react';

import * as config from '../config';

export function useFilter({ metrics }) {
  const [filter, setFilter] = React.useState(config.getFilter());
  const [filterStats, setFilterStats] = React.useState('');
  const [verbosity, setVerbosity] = React.useState(config.getVerbosity());
  const [filteredMetrics, setFilteredMetrics] = React.useState(metrics);

  React.useEffect(() => {
    if (metrics == null) {
      setFilterStats('');
      setFilteredMetrics(null);
      return;
    }

    const newFilteredMetrics = { ...metrics, clusters: [] };
    const filterTerms = filter.split(/\s+/).filter((term) => term.length > 0);
    let numClusters = 0,
      numVisibleClusters = 0,
      numMetrics = 0,
      numVisibleMetrics = 0;
    metrics.clusters.forEach((cluster) => {
      numClusters += 1;
      numMetrics += cluster.metrics.length;

      const newCluster = { ...cluster, metrics: [] };
      newCluster.metrics = cluster.metrics.filter((metric) => {
        if (verbosity === 'normal' && metric.debug) {
          return false;
        }

        if (filterTerms.length > 0) {
          return filterTerms.some((term) => metric.name.includes(term));
        }

        return true;
      });

      if (newCluster.metrics.length > 0) {
        numVisibleClusters += 1;
        numVisibleMetrics += newCluster.metrics.length;
        newFilteredMetrics.clusters.push(newCluster);
      }
    });

    setFilterStats(
      `${numVisibleMetrics} metrics found (${numMetrics - numVisibleMetrics} hidden),` +
        ` organized in ${numVisibleClusters} clusters (${numClusters - numVisibleClusters}` +
        ' hidden)',
    );
    setFilteredMetrics(newFilteredMetrics);
  }, [metrics, filter, verbosity]);

  return [
    filteredMetrics,
    filter,
    setFilter,
    filterStats,
    verbosity,
    setVerbosity,
  ];
}
