import * as config from '../../config';
import { EVENTS } from '../events';

export function Actions() {
  const onResetClick = () => {
    config.reset();
    location.reload();
  };

  const onCollapseClick = () => {
    document.dispatchEvent(new Event(EVENTS.COLLAPSE_ALL_CLUSTERS));
  };

  const onExpandClick = () => {
    document.dispatchEvent(new Event(EVENTS.EXPAND_ALL_CLUSTERS));
  };

  return (
    <>
      <a
        className="btn btn-link"
        href="/metrics"
        role="button"
        title="View internal Prometheus metrics"
      >
        internal metrics
      </a>{' '}
      |
      <button
        type="button"
        className="btn btn-link"
        title="Discard saved state & reload"
        onClick={onResetClick}
      >
        reset
      </button>{' '}
      |
      <button
        type="button"
        className="btn btn-link"
        title="Collapse all clusters"
        onClick={onCollapseClick}
      >
        collapse
      </button>{' '}
      |
      <button
        type="button"
        className="btn btn-link"
        title="Expand all clusters"
        onClick={onExpandClick}
      >
        expand
      </button>
    </>
  );
}
