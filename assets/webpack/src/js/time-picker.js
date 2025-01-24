import flatpickr from 'flatpickr';
import 'flatpickr/dist/flatpickr.min.css';

import * as helpers from './helpers';

/**
* A basic time picker that wraps Flatpickr to support very simple relative
* expressions such as 'now', 'now-1h', 'now-30m', etc.
*/
class TimePicker {
  constructor(element) {
    this.element = element;
    this.flatpickr = null;
    this.expression = null;
    this.init();
  }

  init() {
    this.flatpickr = flatpickr(this.element, {
      enableTime: true,
      dateFormat: 'Y-m-d H:i',
      allowInput: true,
      allowInvalidPreload: true,
      clickOpens: true,
      parseDate: this.handleParseDate.bind(this),
      onValueUpdate: this.handleValueUpdate.bind(this),
      onChange: this.handleChange.bind(this),
      errorHandler: this.handleError.bind(this),
    });
  }

  //
  // Handlers for the wrapped Flatpickr instance. This is where the dirty hacks
  // to support relative expressions are.
  //

  handleParseDate(datestr, format) {
    const parsed = this.parseExpression(datestr);
    if (parsed != null) {
      this.expression = datestr;
      return null;
    }
    this.expression = null;
    return flatpickr.parseDate(datestr, format);
  }

  handleValueUpdate(_selectedDates, _dateStr, _instance) {
  }

  handleChange(_selectedDates, _dateStr, _instance) {
    this.element.classList.remove('is-invalid');
  }

  handleError(_error) {
    this.flatpickr.close();
    if (this.expression == null) {
      this.element.classList.add('is-invalid');
    } else {
      this.element.classList.remove('is-invalid');
    }
  }

  //
  // Public API.
  //

  // Tries to set the date of the picker. The date can be a Date object, a
  // string in ISO format, or a relative expression like 'now-1h'.
  setDate(date) {
    if (typeof date === 'string') {
      const parsed = this.parseExpression(date);
      if (parsed != null) {
        this.expression = date;
        this.element.value = date;
        return;
      }
    }
    this.flatpickr.setDate(date, false);
  }

  // Returns the selected date as a Date object, or null if no date is selected.
  // For relative expressions, the returned date is the result of evaluating the
  // expression.
  getDate() {
    if (this.flatpickr.selectedDates[0] != null) {
      return this.flatpickr.selectedDates[0];
    }
    if (this.expression != null) {
      return this.parseExpression(this.expression);
    }
    return null;
  }

  // Returns the selected date as a Date object, or null if no date is selected.
  // For relative expressions, the returned date is the raw expression string.
  getRawDate() {
    if (this.flatpickr.selectedDates[0] != null) {
      return this.flatpickr.selectedDates[0];
    }
    if (this.expression != null) {
      return this.expression;
    }
    return null;
  }

  // Returns true if the selected date is a relative expression, false otherwise.
  isRelativeDate() {
    return this.flatpickr.selectedDates[0] == null && this.expression != null;
  }

  // Returns a factory function that returns the currently selected date as a
  // Date object, or null if no date is selected. The factory function is not
  // affected by changes to the picker state: for absolute dates, it always
  // returns the same value; for relative expressions, it always evaluates the
  // expression relative to the current time when called.
  getDateFactory() {
    if (this.flatpickr.selectedDates[0] != null) {
      const date = this.flatpickr.selectedDates[0];
      return () => date;
    }
    if (this.expression != null) {
      const expression = this.expression;
      return () => this.parseExpression(expression);
    }
    return null;
  }

  //
  // Private helpers.
  //

  parseExpression(expression) {
    const now = new Date();
    now.setMilliseconds(0);

    if (expression.toLowerCase() === 'now') {
      return now;
    }

    // Match expressions like 'now + 5h', 'now - 30m', 'now + 45s', 'now - 3d',
    // etc.
    const match = expression.match(/^\s*now\s*(-|\+)\s*(\d+)([dhms])\s*$/i);
    if (match) {
      const [, operation, offset, unit] = match;
      let offsetInSeconds;
      switch (unit.toLowerCase()) {
      case 'd':
        offsetInSeconds = offset * 60 * 60 * 24;
        break;
      case 'h':
        offsetInSeconds = offset * 60 * 60;
        break;
      case 'm':
        offsetInSeconds = offset * 60;
        break;
      case 's':
        offsetInSeconds = offset;
        break;
      default:
        return null;
      }
      if (operation === '-') {
        offsetInSeconds = -offsetInSeconds;
      }
      return helpers.unixToDate(helpers.dateToUnix(now) + offsetInSeconds);
    }

    return null;
  }
}

/**
* A time range picker that wraps two TimePicker instances to provide a
* convenient way to select a time range.
*/
class TimeRangePicker {
  constructor(fromElement, toElement) {
    this.fromPicker = new TimePicker(fromElement);
    this.toPicker = new TimePicker(toElement);
  }

  setDates(from, to) {
    this.fromPicker.setDate(from);
    this.toPicker.setDate(to);
  }

  getDates() {
    return [this.fromPicker.getDate(), this.toPicker.getDate()];
  }

  getRawDates() {
    return [this.fromPicker.getRawDate(), this.toPicker.getRawDate()];
  }

  getDatesFactory() {
    const fromFactory = this.fromPicker.getDateFactory();
    const toFactory = this.toPicker.getDateFactory();
    return () => [fromFactory(), toFactory()];
  }

  hasValidDates() {
    const from = this.fromPicker.getDate();
    const to = this.toPicker.getDate();
    return from != null && to != null && from <= to;
  }
}

export { TimePicker, TimeRangePicker };
