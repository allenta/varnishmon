import * as helpers from './helpers';

import * as varnish from './varnish';

/******************************************************************************
 * METRICS.
 ******************************************************************************/

/**
 * Retrieves metrics from the storage API. The returned metrics are sorted,
 * clustered, and tagged according to their verbosity level using a set of
 * simple client-side heuristics.
 *
 * @param {Date} from - The start of the time range, optionally aligned to a
 * step boundary.
 * @param {Date} to - The end of the time range, optionally aligned to a step
 * boundary.
 * @param {number} step - The time step in seconds.
 * @returns {Object} The clustered metrics plus the time range and step
 * parameters adjusted by the storage API (e.g., aligned to step boundaries).
 */
export async function getMetrics(from, to, step) {
  const params = new URLSearchParams({
    from: helpers.dateToUnix(from),
    to: helpers.dateToUnix(to),
    step: step,
  });
  const response = await fetch(`/storage/metrics?${params.toString()}`);
  if (!response.ok) {
    throw new Error(`Unexpected API response (${response.status}): ${response.statusText}`);
  }

  const data = await response.json();
  return {
    from: helpers.unixToDate(data.from),
    to: helpers.unixToDate(data.to),
    step: data.step,
    clusters: preprocessMetrics(data.metrics),
  };
}

/**
 * Sorts, clusters, and tags the metrics according to their verbosity level.
 *
 * @param {Array} metrics - The metrics to process, as returned by the storage
 * API.
 * @returns {Array} The processed metrics.
 */
function preprocessMetrics(metrics) {
  function addDebugField(metrics) {
    metrics.forEach(metric => {
      metric.debug = varnish.DEBUG_METRICS.findIndex(regex => regex.test(metric.name)) !== -1;
    });
  }

  function clusterByPrefix(metrics) {
    // Clusters are defined by the longest prefix just before the last dot,
    // unless explicitly overridden by a regex in 'ADHOC_CLUSTERING_PREFIXES'.
    const clusters = {};
    metrics.forEach(metric => {
      let prefix = '';

      // Check against the list of ad-hoc clustering prefixes.
      for (let regex of varnish.ADHOC_CLUSTERING_PREFIXES) {
        const match = metric.name.match(regex);
        if (match && match[1]) {
          prefix = match[1] + '*';
          break;
        }
      }

      // If no ad-hoc clustering matched, use the default behavior.
      if (prefix === '') {
        const parts = metric.name.split('.');
        if (parts.length > 1) {
          prefix = parts.slice(0, -1).join('.') + '.*';
        }
      }

      // Add the metric to the cluster.
      if (!clusters[prefix]) {
        clusters[prefix] = [];
      }
      clusters[prefix].push(metric);
    });

    // Convert clusters object to an array of objects.
    const result = Object.keys(clusters).map(key => ({
      name: key,
      metrics: clusters[key],
    }));
    return result;
  }

  function sortClusteredMetrics(items) {
    items.sort((a, b) => {
      const indexA = varnish.ORDER_OF_CLUSTERS.findIndex(regex => regex.test(a.name));
      const indexB = varnish.ORDER_OF_CLUSTERS.findIndex(regex => regex.test(b.name));

      if (indexA !== -1 && indexB !== -1) {
        return indexA - indexB;
      } else if (indexA !== -1) {
        return -1;
      } else if (indexB !== -1) {
        return 1;
      } else {
        return a.name.localeCompare(b.name);
      }
    });
  }

  // Sort metrics by name.
  metrics.sort((a, b) => a.name.localeCompare(b.name));

  // Add the 'debug' field to each metric.
  addDebugField(metrics);

  // Cluster metrics by prefix.
  const clusteredMetrics = clusterByPrefix(metrics);

  // Sort clusters by name applying some ad-hoc rules.
  sortClusteredMetrics(clusteredMetrics);

  // Done!
  return clusteredMetrics;
}

/******************************************************************************
 * METRIC.
 ******************************************************************************/

/**
 * Retrieves samples of a metric from the storage API.
 *
 * @param {number} id - The metric identifier.
 * @param {Date} from - The start of the time range, optionally aligned to a
 * step boundary.
 * @param {Date} to - The end of the time range, optionally aligned to a step
 * boundary.
 * @param {number} step - The time step in seconds.
 * @param {string} aggregator - The aggregation function to use.
 * @returns {Object} The metric samples plus the time range and step parameters
 * adjusted by the storage API (e.g., aligned to step boundaries).
 */
export async function getMetric(id, from, to, step, aggregator) {
  const params = new URLSearchParams({
    from: helpers.dateToUnix(from),
    to: helpers.dateToUnix(to),
    step: step,
    aggregator: aggregator,
  });
  const response = await fetch(`/storage/metrics/${id}?${params.toString()}`);
  if (!response.ok) {
    throw new Error(`Unexpected API response (${response.status}): ${response.statusText}`);
  }

  const data = await response.json();
  return {
    from: helpers.unixToDate(data.from),
    to: helpers.unixToDate(data.to),
    step: data.step,
    samples: preprocessSamples(data.samples, data.step),
  };
}

/**
 * Sorts the samples by timestamp, converts the timestamps to Date objects and
 * injects null values in all detected gaps.
 *
 * @param {Array} samples - The samples to process, as returned by the storage
 * API.
 * @param {number} step - The time step in seconds.
 * @returns {Array} The processed samples.
 */
function preprocessSamples(samples, step) {
  // Sort samples by timestamp.
  const sortedSamples = samples.sort((a, b) => a[0] - b[0]);

  // Fill gaps with nulls and convert timestamps to Date objects.
  const preprocessedSamples = [];
  for (let i = 0; i < sortedSamples.length; i++) {
    // Add the current sample to the processed samples.
    const currentSample = sortedSamples[i];
    const currentTime = currentSample[0];
    const currentValue = currentSample[1];
    preprocessedSamples.push([helpers.unixToDate(currentTime), currentValue]);

    // Check if there is a gap to the next sample and fill it with nulls in each
    // missing step.
    if (i < sortedSamples.length - 1) {
      const nextSample = sortedSamples[i + 1];
      const nextTime = nextSample[0];
      for (let j = 1; j < (nextTime - currentTime) / step; j++) {
        preprocessedSamples.push([helpers.unixToDate(currentTime + j * step), null]);
      }
    }
  }

  return preprocessedSamples;
}
