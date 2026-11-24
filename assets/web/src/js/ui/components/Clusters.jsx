import { Cluster } from './Cluster';

/**
 * @param {object} props - Component props.
 * @param {object} props.timeRangePicker
 * @param {object} props.initialRangeRef
 * @param {number} props.refreshInterval
 * @param {number} props.columns
 * @param {string} props.aggregator
 * @param {number} props.step
 * @param {?object} [props.filteredMetrics]
 */
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
