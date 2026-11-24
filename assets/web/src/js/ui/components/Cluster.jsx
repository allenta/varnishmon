import React from 'react';
import Collapse from 'bootstrap/js/dist/collapse';

import { Metric } from './Metric';
import { EVENTS } from '../events';

/**
 * @param {object} props - Component props.
 * @param {object} props.timeRangePicker
 * @param {object} props.initialRangeRef
 * @param {number} props.refreshInterval
 * @param {number} props.columns
 * @param {string} props.aggregator
 * @param {number} props.step
 * @param {object} props.cluster
 */
export function Cluster({ cluster, ...props }) {
  const containerRef = React.useRef(null);
  const [accordion, setAccordion] = React.useState(null);
  const [isAccordionCollapsed, setIsAccordionCollapsed] = React.useState(false);

  React.useEffect(() => {
    if (containerRef.current != null) {
      const accordionCollapse = containerRef.current.querySelector(
        '.accordion-collapse',
      );
      const newAccordion = new Collapse(accordionCollapse);
      setAccordion(newAccordion);

      const onCollapseAllClustersListener = () => {
        setIsAccordionCollapsed(true);
      };
      document.addEventListener(
        EVENTS.COLLAPSE_ALL_CLUSTERS,
        onCollapseAllClustersListener,
      );

      const onExpandAllClustersListener = () => {
        setIsAccordionCollapsed(false);
      };
      document.addEventListener(
        EVENTS.EXPAND_ALL_CLUSTERS,
        onExpandAllClustersListener,
      );

      return () => {
        document.removeEventListener(
          EVENTS.COLLAPSE_ALL_CLUSTERS,
          onCollapseAllClustersListener,
        );
        document.removeEventListener(
          EVENTS.EXPAND_ALL_CLUSTERS,
          onExpandAllClustersListener,
        );
        // When mounting and then unmounting the component too quickly (e.g.,
        // due to fast reloads of metrics, strict mode, etc.), calling
        // 'newAccordion.dispose()' immediately after 'newAccordion' is created
        // causes errors to be logged in the console. UI transitions are
        // executed asynchronously and that doesn't play well with immediate
        // disposal of the instance. A hardcoded delay is used here to mitigate
        // this issue. The timeout is intentionally not tracked: it is the
        // cleanup itself, so there is nothing left to cancel it.
        // eslint-disable-next-line @eslint-react/web-api-no-leaked-timeout
        setTimeout(() => newAccordion.dispose(), 1000);
      };
    }
  }, [containerRef]);

  React.useEffect(() => {
    if (accordion != null) {
      if (isAccordionCollapsed) {
        accordion.hide();
      } else {
        accordion.show();
      }
    }
  }, [accordion, isAccordionCollapsed]);

  const onClick = () => {
    setIsAccordionCollapsed(!isAccordionCollapsed);
  };

  return (
    <div ref={containerRef}>
      <div className="accordion-header">
        <button
          className={`accordion-button bg-light text-dark fs-5 border-0 font-monospace ${isAccordionCollapsed ? 'collapsed' : ''}`}
          type="button"
          onClick={onClick}
        >
          {cluster.name}
        </button>
      </div>
      <div className="accordion-collapse">
        <div className="row g-4 py-4">
          {cluster.metrics.map((metric) => (
            <Metric key={metric.id} {...props} metric={metric} />
          ))}
        </div>
      </div>
    </div>
  );
}
