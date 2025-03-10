import PropTypes from 'prop-types';
import React from 'react';
import Collapse from 'bootstrap/js/dist/collapse';

import { Metric } from './Metric';

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
        'onCollapseAllClusters',
        onCollapseAllClustersListener,
      );

      const onExpandAllClustersListener = () => {
        setIsAccordionCollapsed(false);
      };
      document.addEventListener(
        'onExpandAllClusters',
        onExpandAllClustersListener,
      );

      return () => {
        document.removeEventListener(
          'onCollapseAllClusters',
          onCollapseAllClustersListener,
        );
        document.removeEventListener(
          'onExpandAllClusters',
          onExpandAllClustersListener,
        );
        // When mounting and then unmounting the component too quickly (e.g.,
        // due to fast reloads of metrics, strict mode, etc.), calling
        // 'newAccordion.dispose()' immediately after 'newAccordion' is created
        // causes errors to be logged in the console. UI transitions are
        // executed asynchronously and that doesn't play well with immediate
        // disposal of the instance. A hardcoded delay is used here to mitigate
        // this issue.
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

Cluster.propTypes = {
  timeRangePicker: PropTypes.object.isRequired,
  initialRange: PropTypes.object.isRequired,
  refreshInterval: PropTypes.number.isRequired,
  columns: PropTypes.number.isRequired,
  aggregator: PropTypes.string.isRequired,
  step: PropTypes.number.isRequired,
  cluster: PropTypes.object.isRequired,
};
