import * as helpers from './helpers';

// Debug metrics are explicitly tagged as such. Otherwise, they are assumed to
// be regular metrics. The following is strongly opinionated but mostly based on
// *.vsc files in https://github.com/varnishcache/varnish-cache/tree/master/lib/libvcc/.
// None of this would be necessary if the 'varnishstat -j' output provided the level
// (info / debug / diag) for each metric, which should be trivial because the
// information is available, but it doesn't.
const debugMetrics = [
  /^MGT[.](?!(?:uptime)$)/,
  new RegExp('MAIN[.](?!(?:' + [
    // TODO.
  ].join('|') + ')$)'),
  new RegExp('MSE[.](?!.*[.](?:' + [
    'c_fail',
    'c_memcache_hit',
    'c_memcache_miss',
    'c_ykey_purged',
    'g_bytes',
    'g_space',
    'g_sparenode',
    'g_ykey_keys',
    'n_lru_moved',
    'n_lru_nuked',
    'n_vary',
  ].join('|') + ')$)'),
  new RegExp('MSE_BOOK[.](?!.*[.](?:' + [
    'c_insert_timeout',
    'c_waterlevel_purge',
    'c_waterlevel_queue',
    'c_waterlevel_runs',
    'g_banlist_bytes',
    'g_banlist_space',
    'g_bytes',
    'g_space',
    'n_vary',
  ].join('|') + ')$)'),
  new RegExp('MSE_STORE[.](?!.*[.](?:' + [
    'c_aio_finished_bytes_read',
    'c_aio_finished_bytes_write',
    'c_aio_finished_read',
    'c_aio_finished_write',
    'c_waterlevel_purge',
    'c_waterlevel_queue',
    'g_alloc_bytes',
    'g_free_bytes',
    'g_objects',
    'g_ykey_keys',
  ].join('|') + ')$)'),
  new RegExp('MSE4[.](?!(?:' + [
    'g_varyspec',
    'g_ykey_keys',
    'c_ykey_purged',
  ].join('|') + ')$)'),
  new RegExp('MSE4_MEM[.](?!(?:' + [
    'c_allocation',
    'c_allocation_buffer',
    'c_allocation_ephemeral',
    'c_allocation_failure',
    'c_allocation_pass',
    'c_allocation_persisted',
    'c_allocation_reqbody',
    'c_allocation_synthetic',
    'c_eviction',
    'c_eviction_failure',
    'c_eviction_reorder',
    'c_free',
    'c_free_buffer',
    'c_free_ephemeral',
    'c_free_pass',
    'c_free_persisted',
    'c_free_reqbody',
    'c_free_synthetic',
    'c_memcache_hit',
    'c_memcache_miss',
    'g_allocations',
    'g_bytes',
    'g_bytes_buffer',
    'g_bytes_ephemeral',
    'g_bytes_pass',
    'g_bytes_persisted',
    'g_bytes_reqbody',
    'g_bytes_synthetic',
    'g_objects',
    'g_objects_ephemeral',
    'g_objects_pass',
    'g_objects_persisted',
    'g_objects_reqbody',
    'g_objects_synthetic',
    'g_space',
  ].join('|') + ')$)'),
  new RegExp('MSE4_BOOK[.](?!.*[.](?:' + [
    'c_freeslot_queued',
    'c_submitslot_queued',
    'c_ykey_purged',
    'g_freeslot_queue',
    'g_objects',
    'g_slots_unused',
    'g_slots_used',
    'g_submitslot_queue',
    'g_unreachable_objects',
    'g_varyspec',
    'g_ykey_keys',
    'online',
  ].join('|') + ')$)'),
  new RegExp('MSE4_STORE[.](?!.*[.](?:' + [
    'online',
    'g_bytes_used',
    'g_bytes_unused',
    'g_objects',
    'g_allocation_queue',
    'c_allocation_queued',
    'g_io_queued',
    'g_io_queued_read',
    'g_io_queued_write',
    'c_io_finished_read',
    'c_io_finished_write',
    'c_io_finished_bytes_read',
    'c_io_finished_bytes_write',
    'g_io_blocked_read',
    'g_io_blocked_write',
    'c_io_limited',
  ].join('|') + ')$)'),
  new RegExp('MSE4_BANJRN[.](?!.*[.](?:' + [
    'g_ban_bytes',
    'g_bans',
    'g_bytes',
    'g_overflow_ban_bytes',
    'g_overflow_bans',
    'g_space',
  ].join('|') + ')$)'),
  new RegExp('MSE4_CAT[.](?!.*[.](?:' + [
    'c_allocation',
    'c_allocation_ephemeral',
    'c_allocation_pass',
    'c_allocation_persisted',
    'c_eviction',
    'c_eviction_failure',
    'c_eviction_reorder',
    'c_free',
    'c_free_ephemeral',
    'c_free_pass',
    'c_free_persisted',
    'c_memcache_hit',
    'c_memcache_miss',
    'g_allocations',
    'g_bytes',
    'g_bytes_ephemeral',
    'g_bytes_pass',
    'g_bytes_persisted',
    'g_objects',
    'g_objects_ephemeral',
    'g_objects_pass',
    'g_objects_persisted',
  ].join('|') + ')$)'),
  new RegExp('SMA[.](?!.*[.](?:' + [
    'c_bytes',
    'c_fail',
    'c_freed',
    'c_req',
    'g_alloc',
    'g_bytes',
    'g_space',
  ].join('|') + ')$)'),
  new RegExp('SMF[.](?!.*[.](?:' + [
    'c_bytes',
    'c_fail',
    'c_freed',
    'c_req',
    'g_alloc',
    'g_bytes',
    'g_smf_frag',
    'g_smf_large',
    'g_smf',
    'g_space',
  ].join('|') + ')$)'),
  new RegExp('SMU[.](?!.*[.](?:' + [
    'c_bytes',
    'c_fail',
    'c_freed',
    'c_req',
    'g_alloc',
    'g_bytes',
    'g_space',
  ].join('|') + ')$)'),
  /^BROTLI[.]/,
  /^SLICER[.]/,
  new RegExp('VMOD_HTTP[.](?!(?:' + [
    'handle_abandon',
    'handle_completed',
    'handle_internal_error',
    'handle_limited',
    'handle_requests',
  ].join('|') + ')$)'),
  new RegExp('KVSTORE[.](?!.*[.](?:' + [
    // TBD.
  ].join('|') + ')$)'),
  new RegExp('ACCG[.](?!.*[.](?:' + [
    'backend_200_count',
    'backend_2xx_count',
    'backend_304_count',
    'backend_3xx_count',
    'backend_404_count',
    'backend_4xx_count',
    'backend_503_count',
    'backend_5xx_count',
    'backend_req_bodybytes',
    'backend_req_count',
    'backend_req_hdrbytes',
    'backend_resp_bodybytes',
    'backend_resp_hdrbytes',
    'client_200_count',
    'client_2xx_count',
    'client_304_count',
    'client_3xx_count',
    'client_404_count',
    'client_4xx_count',
    'client_503_count',
    'client_5xx_count',
    'client_grace_hit_count',
    'client_hit_count',
    'client_hit_req_bodybytes',
    'client_hit_req_hdrbytes',
    'client_hit_resp_bodybytes',
    'client_hit_resp_hdrbytes',
    'client_miss_count',
    'client_miss_req_bodybytes',
    'client_miss_req_hdrbytes',
    'client_miss_resp_bodybytes',
    'client_miss_resp_hdrbytes',
    'client_pass_count',
    'client_pass_req_bodybytes',
    'client_pass_req_hdrbytes',
    'client_pass_resp_bodybytes',
    'client_pass_resp_hdrbytes',
    'client_pipe_count',
    'client_pipe_req_bodybytes',
    'client_pipe_req_hdrbytes',
    'client_pipe_resp_bodybytes',
    'client_pipe_resp_hdrbytes',
    'client_req_bodybytes',
    'client_req_count',
    'client_req_hdrbytes',
    'client_resp_bodybytes',
    'client_resp_hdrbytes',
    'client_synth_count',
    'client_synth_req_bodybytes',
    'client_synth_req_hdrbytes',
    'client_synth_resp_bodybytes',
    'client_synth_resp_hdrbytes',
  ].join('|') + ')$)'),
  new RegExp('ACCG_DIAG[.](?!.*[.](?:' + [
    'bereq_dropped',
    'create_namespace_failure',
    'key_without_namespace',
    'namespace_already_set',
    'namespace_undefined',
    'out_of_key_slots',
    'req_dropped',
    'set_key_failure',
  ].join('|') + ')$)'),
  new RegExp('VBE[.](?!.*[.](?:' + [
    'bereq_bodybytes',
    'bereq_hdrbytes',
    'beresp_bodybytes',
    'beresp_hdrbytes',
    'busy',
    'conn',
    'fail',
    'happy',
    'is_healthy',
    'pipe_hdrbytes',
    'pipe_in',
    'pipe_out',
    'req',
    'unhealthy',
  ].join('|') + ')$)'),
  /^WAITER[.]/,
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
  /^(WAITER[.])/,
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
  /^SMU[.]/,
  /^BROTLI[.]/,
  /^SLICER[.]/,
  /^VMOD_/,
  /^KVSTORE[.]/,
  /^ACCG[.]/,
  /^ACCG_DIAG[.]/,
  /^VBE[.]/,
  /^WAITER[.]/,
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
