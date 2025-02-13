import * as helpers from './helpers';

const PREFIX = 'varnishmon.';

/******************************************************************************
* TIME RANGE.
******************************************************************************/

const TIME_RANGE_FROM = `${PREFIX}time-range-from`;
const TIME_RANGE_TO = `${PREFIX}time-range-to`;
const DEFAULT_RELATIVE_TIME_RANGE = ['now-1h', 'now'];

export function getTimeRange(skipLocalStorage = false) {
  if (!skipLocalStorage) {
    const from = getTimeRangeValue(TIME_RANGE_FROM);
    const to = getTimeRangeValue(TIME_RANGE_TO);
    if (from != null && to != null) {
      // The 'from' and 'to' values retrieved from local storage can be
      // either ISO strings or relative date expressions (or a mix of both).
      // In either case, they are returned as-is, allowing the time picker
      // component to handle the conversion.
      return [from, to];
    }
  }

  if (varnishmon.config.scraper.enabled) {
    return DEFAULT_RELATIVE_TIME_RANGE;
  }

  return [
    helpers.unixToDate(varnishmon.storage.earliest),
    helpers.unixToDate(varnishmon.storage.latest),
  ];
}

export function setTimeRange(from, to) {
  setTimeRangeValue(TIME_RANGE_FROM, from);
  setTimeRangeValue(TIME_RANGE_TO, to);
}

function getTimeRangeValue(key) {
  try {
    let value = localStorage.getItem(key);
    if (value != null) {
      return value;
    }
  } catch (error) {
    console.error(`Failed to read '${key}' from local storage!`, error);
  }

  return null;
}

function setTimeRangeValue(key, value) {
  if (value instanceof Date) {
    value = value.toISOString();
  }

  try {
    localStorage.setItem(key, value);
  } catch (error) {
    console.error(`Failed to write '${key}' to local storage!`, error);
  }
}

/******************************************************************************
* REFRESH INTERVAL.
******************************************************************************/

const REFRESH_INTERVAL = `${PREFIX}refresh-interval`;
const REFRESH_INTERVAL_VALUES = [
  [-1, 'auto'],
  [0, 'disabled'],
  [1, '1s'],
  [2, '2s'],
  [3, '3s'],
  [4, '4s'],
  [5, '5s'],
  [10, '10s'],
  [15, '15s'],
  [30, '30s'],
  [60, '1m'],
];

export function getRefreshInterval() {
  try {
    let value = localStorage.getItem(REFRESH_INTERVAL);
    if (value != null) {
      value = parseInt(value, 10);
      if (isValidRefreshIntervalValue(value)) {
        return value;
      }
    }
  } catch (error) {
    console.error(`Failed to read '${REFRESH_INTERVAL}' from local storage!`, error);
  }

  return varnishmon.config.scraper.enabled ? -1 : 0;
}

export function setRefreshInterval(value) {
  if (!isValidRefreshIntervalValue(value)) {
    console.error('Invalid refresh interval value!', value);
    return;
  }

  try {
    localStorage.setItem(REFRESH_INTERVAL, value);
  } catch (error) {
    console.error(`Failed to write '${REFRESH_INTERVAL}' to local storage!`, error);
  }
}

export function getRefreshIntervalValues() {
  return REFRESH_INTERVAL_VALUES;
}

function isValidRefreshIntervalValue(value) {
  return Number.isInteger(value) && REFRESH_INTERVAL_VALUES.map(v => v[0]).includes(value);
}

/******************************************************************************
* FILTER.
******************************************************************************/

const FILTER = `${PREFIX}filter`;
const FILTER_HISTORY = `${PREFIX}filter-history`;

export function getFilter() {
  try {
    let value = localStorage.getItem(FILTER);
    if (value != null) {
      return value;
    }
  } catch (error) {
    console.error(`Failed to read '${FILTER}' from local storage!`, error);
  }

  return '';
}

export function setFilter(value) {
  try {
    localStorage.setItem(FILTER, value);
  } catch (error) {
    console.error(`Failed to write '${FILTER}' to local storage!`, error);
  }
}

export function getFilterHistory() {
  try {
    let value = localStorage.getItem(FILTER_HISTORY);
    if (value != null) {
      return JSON.parse(value);
    }
  } catch (error) {
    console.error(`Failed to read '${FILTER_HISTORY}' from local storage!`, error);
  }

  return [];
}

export function setFilterHistory(value) {
  try {
    localStorage.setItem(FILTER_HISTORY, JSON.stringify(value));
  } catch (error) {
    console.error(`Failed to write '${FILTER_HISTORY}' to local storage!`, error);
  }
}

/******************************************************************************
* VERBOSITY.
******************************************************************************/

const VERBOSITY = `${PREFIX}verbosity`;
const VERBOSITY_VALUES = ['normal', 'debug'];

export function getVerbosity() {
  try {
    let value = localStorage.getItem(VERBOSITY);
    if (value != null && isValidVerbosityValue(value)) {
      return value;
    }
  } catch (error) {
    console.error(`Failed to read '${VERBOSITY}' from local storage!`, error);
  }

  return 'normal';
}

export function setVerbosity(value) {
  if (!isValidVerbosityValue(value)) {
    console.error('Invalid verbosity value!', value);
    return;
  }

  try {
    localStorage.setItem(VERBOSITY, value);
  } catch (error) {
    console.error(`Failed to write '${VERBOSITY}' to local storage!`, error);
  }
}

export function getVerbosityValues() {
  return VERBOSITY_VALUES;
}

function isValidVerbosityValue(value) {
  return VERBOSITY_VALUES.includes(value);
}

/******************************************************************************
* COLUMNS.
******************************************************************************/

const COLUMNS = `${PREFIX}columns`;
const COLUMNS_VALUES = [1, 2, 3, 4, 6, 12];

export function getColumns() {
  try {
    let value = localStorage.getItem(COLUMNS);
    if (value != null) {
      value = parseInt(value, 10);
      if (isValidColumnsValue(value)) {
        return value;
      }
    }
  } catch (error) {
    console.error(`Failed to read '${COLUMNS}' from local storage!`, error);
  }

  return 3;
}

export function setColumns(value) {
  if (!isValidColumnsValue(value)) {
    console.error('Invalid columns value!', value);
    return;
  }

  try {
    localStorage.setItem(COLUMNS, value);
  } catch (error) {
    console.error(`Failed to write '${COLUMNS}' to local storage!`, error);
  }
}

export function getColumnsValues() {
  return COLUMNS_VALUES;
}

function isValidColumnsValue(value) {
  return Number.isInteger(value) && COLUMNS_VALUES.includes(value);
}

/******************************************************************************
* AGGREGATOR.
******************************************************************************/

const AGGREGATOR = `${PREFIX}aggregator`;
const AGGREGATOR_VALUES = ['avg', 'min', 'max', 'first', 'last', 'count'];

export function getAggregator() {
  try {
    let value = localStorage.getItem(AGGREGATOR);
    if (value != null && isValidAggregatorValue(value)) {
      return value;
    }
  } catch (error) {
    console.error(`Failed to read '${AGGREGATOR}' from local storage!`, error);
  }

  return 'avg';
}

export function setAggregator(value) {
  if (!isValidAggregatorValue(value)) {
    console.error('Invalid aggregator value!', value);
    return;
  }

  try {
    localStorage.setItem(AGGREGATOR, value);
  } catch (error) {
    console.error(`Failed to write '${AGGREGATOR}' to local storage!`, error);
  }
}

export function getAggregatorValues() {
  return AGGREGATOR_VALUES;
}

function isValidAggregatorValue(value) {
  return AGGREGATOR_VALUES.includes(value);
}

/******************************************************************************
* STEP.
******************************************************************************/

const STEP = `${PREFIX}step`;
const DEFAULT_STEP = 60;

export function getStep() {
  try {
    let value = localStorage.getItem(STEP);
    if (value != null) {
      value = parseInt(value, 10);
      if (isValidStepValue(value)) {
        return value;
      }
    }
  } catch (error) {
    console.error(`Failed to read '${STEP}' from local storage!`, error);
  }

  return varnishmon.config.scraper.enabled ? varnishmon.config.scraper.period : DEFAULT_STEP;
}

export function setStep(value) {
  if (!isValidStepValue(value)) {
    console.error('Invalid columns value!', value);
    return;
  }

  try {
    localStorage.setItem(STEP, value);
  } catch (error) {
    console.error(`Failed to write '${STEP}' to local storage!`, error);
  }
}

export function getMinimumStep() {
  return varnishmon.config.scraper.enabled ? varnishmon.config.scraper.period : 1;
}

function isValidStepValue(value) {
  return Number.isInteger(value) && value >= getMinimumStep();
}

/******************************************************************************
* RESET.
******************************************************************************/

export function reset() {
  for (let i = localStorage.length - 1; i >= 0; i--) {
    try {
      const key = localStorage.key(i);
      if (key.startsWith(PREFIX)) {
        localStorage.removeItem(key);
      }
    } catch (error) {
      console.error('Failed to remove item from local storage!', error);
    }
  }
}
