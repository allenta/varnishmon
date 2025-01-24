import * as helpers from './helpers';

// Debug metrics are explicitly tagged as such. Otherwise they are assumed to be
// regular metrics.
const debugMetrics = [
  /^MGT[.](?!(?:uptime)$)/,
  /^ACCG_DIAG[.]/,
  /^VBE[.](?!.*[.](?:bereq_bodybytes|bereq_hdrbytes|beresp_bodybytes|beresp_hdrbytes|happy|is_healthy|req)$)/,
  /^MEMPOOL[.]/,
  /^LCK[.]/,
];

// Clustering of metrics is based on the longest prefix just before the last
// dot, unless explicitly overridden here using a regex + a capture group.
const adhocClusteringPrefixes = [
  /^(MAIN[.]backend)/,
  /^(MAIN[.]bans)_?/,
  /^(MAIN[.]cache)/,
  /^(MAIN[.]client)/,
  /^(MAIN[.]esi_)/,
  /^(MAIN[.]fetch)/,
  /^(MAIN[.]g_mem)/,
  /^(MAIN[.]s_)/,
  /^(MAIN[.]sc_)/,
  /^(MAIN[.]sess_)/,
  /^(MAIN[.]shm_)/,
  /^(MAIN[.]thread)s?_?/,
  /^(MAIN[.]vgs_)/,
  /^(MAIN[.]ws_)/,
  /^(MEMPOOL[.])/,
  /^(LCK[.])/,
];

// Clusters are sorted by name, unless explicitly overridden here using a regex.
const orderOfClusters = [
  /^MGT[.]/,
  /^MAIN[.][*]$/,
  /^MAIN[.]/,
  /^MSE[.]/,
  /^MSE_/,
  /^MSE4[.]/,
  /^MSE4_/,
  /^SMA[.]/,
  /^SMF[.]/,
  /^BROTLI[.]/,
  /^SLICER[.]/,
  /^VMOD_/,
  /^KVSTORE[.]/,
  /^ACCG[.]/,
  /^ACCG_DIAG[.]/,
  /^VBE[.]/,
  /^MEMPOOL[.]/,
  /^LCK[.]/,
];

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
      metric.debug = debugMetrics.findIndex(regex => regex.test(metric.name)) !== -1;
    });
  }

  function clusterByPrefix(metrics) {
    // Clusters are defined by the longest prefix just before the last dot,
    // unless explicitly overridden by a regex in 'adhocClusteringPrefixes'.
    const clusters = {};
    metrics.forEach(metric => {
      let prefix = '';

      // Check against the list of ad-hoc clustering prefixes.
      for (let regex of adhocClusteringPrefixes) {
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
      const indexA = orderOfClusters.findIndex(regex => regex.test(a.name));
      const indexB = orderOfClusters.findIndex(regex => regex.test(b.name));

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
    samples: preprocessSamples(data.samples),
  };
}

/**
 * Sorts the samples by timestamp and converts the timestamps to Date objects.
 *
 * @param {Array} samples - The samples to process, as returned by the storage
 * API.
 * @returns {Array} The processed samples.
 */
function preprocessSamples(samples) {
  return samples.sort((a, b) => a[0] - b[0]).map(sample => ([
    helpers.unixToDate(sample[0]),
    sample[1],
  ]));
}
