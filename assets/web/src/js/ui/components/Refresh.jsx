import * as config from '../../config';
import { EVENTS } from '../events';

/**
 * @param {object} props - Component props.
 * @param {number} props.refreshInterval
 * @param {Function} props.setRefreshInterval
 */
export function Refresh({ refreshInterval, setRefreshInterval }) {
  const onChange = (event) => {
    const value = parseInt(event.target.value, 10);
    config.setRefreshInterval(value);
    setRefreshInterval(value);
  };

  const onClick = () => {
    document.dispatchEvent(new Event(EVENTS.REFRESH));
  };

  return (
    <>
      <div className="me-2">
        <select
          className="form-select"
          value={refreshInterval}
          onChange={onChange}
        >
          {config.getRefreshIntervalValues().map((value) => (
            <option key={value[0]} value={value[0]}>
              {value[1]}
            </option>
          ))}
        </select>
      </div>
      <div className="align-self-center align-self-end">
        <button
          className="btn btn-primary"
          title="Trigger refresh now"
          onClick={onClick}
        >
          <i className="fa-solid fa-sync"></i>
        </button>
      </div>
    </>
  );
}
