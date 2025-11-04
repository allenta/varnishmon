import React from 'react';

import * as helpers from '../helpers';
import * as storage from '../storage';

export function useReload({ timeRangePicker, initialRangeRef, step }) {
  const [metrics, setMetrics] = React.useState(null);
  const [loadProgress, setLoadProgress] = React.useState(null);

  const reload = React.useCallback(() => {
    // Flush previous state.
    initialRangeRef.current = null;
    setMetrics(null);

    // Fetch metrics from the storage, exposing load state to show a loading
    // spinner in the meantime, etc.
    setLoadProgress('loading');
    const rangeFactory = timeRangePicker.getDatesFactory();
    const [from, to] = rangeFactory();
    storage
      .getMetrics(from, to, step)
      .then((fetchedMetrics) => {
        const numClusters = fetchedMetrics.clusters.length;
        const numMetrics = fetchedMetrics.clusters.reduce(
          (acc, cluster) => acc + cluster.metrics.length,
          0,
        );
        helpers.notify(
          'info',
          `Fetched ${numMetrics} metrics organized in ${numClusters} clusters`,
        );
        setLoadProgress(null);

        setMetrics(fetchedMetrics);
      })
      .catch((error) => {
        helpers.notify('error', `Failed to fetch metrics: ${error}`);
        setLoadProgress('error');
      });
  }, [timeRangePicker, initialRangeRef, step]);

  return [metrics, loadProgress, reload];
}
