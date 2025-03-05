import AirDatepicker from 'air-datepicker';
import 'air-datepicker/air-datepicker.css';
import localeEn from 'air-datepicker/locale/en';

import * as helpers from './helpers';

/**
* A basic time picker that wraps Air Datepicker to support very simple relative
* expressions such as 'now', 'now-1h', 'now-30m', etc.
*/
class TimePicker {
  constructor(element) {
    this.element = element;
    this.datepicker = null;
    this.init();
  }

  init() {
    this.element.addEventListener('keydown', this.handleKeyDown.bind(this));
    this.element.addEventListener('blur', this.handleOnBlur.bind(this));

    this.datepicker = new AirDatepicker(this.element, {
      timepicker: true,
      dateFormat: 'yyyy-MM-dd',
      timeFormat: 'HH:mm',
      locale: localeEn,
      firstDay: 1,
      autoClose: false,
      keyboardNav: false,
      // onSelect: this.handleOnSelect.bind(this),
    });
  }

  //
  // Handlers for the wrapped Air Datepicker instance & input field. This
  // contains the main dirty hack to support relative expressions.
  //

  handleKeyDown(event) {
    if (event.key === 'Enter') {
      this.element.blur();
    }
  }

  handleOnBlur() {
    this.datepicker.selectedDates = [];

    if (this.parseExpression(this.element.value) != null) {
      this.element.classList.remove('is-invalid');
    } else {
      const value = this.element.value.trim();
      if (value) {
        const date = new Date(value);
        if (!Number.isNaN(date.valueOf())) {
          this.datepicker.setViewDate(date);
          this.datepicker.selectDate([date], { silent: true, updateTime: true });
          this.element.classList.remove('is-invalid');
        } else {
          this.element.classList.add('is-invalid');
        }
      } else {
        this.element.classList.remove('is-invalid');
      }
    }
  }

  //
  // Public API.
  //

  // Tries to set the date of the picker. The date can be a Date object, a
  // string in ISO format, or a relative expression like 'now-1h'.
  setDate(date) {
    if (date == null) {
      this.datepicker.clear();
      this.element.classList.remove('is-invalid');
    } else if (typeof date === 'string' && this.parseExpression(date) != null) {
      this.element.value = date;
      this.element.classList.remove('is-invalid');
    } else {
      this.datepicker.selectDate([date], { silent: true, updateTime: true});
      if (this.datepicker.selectedDates.length > 0) {
        this.element.classList.remove('is-invalid');
      } else {
        this.element.classList.add('is-invalid');
      }
    }
  }

  // Returns the selected date as a Date object, or null if no date is selected.
  // For relative expressions, the returned date is the result of evaluating the
  // expression.
  getDate() {
    let date = this.parseExpression(this.element.value);
    if (date == null && this.datepicker.selectedDates.length > 0) {
      date = this.datepicker.selectedDates[0];
    }
    return date;
  }

  // Returns the selected date as a Date object, or null if no date is selected.
  // For relative expressions, the returned date is the raw expression string.
  getRawDate() {
    if (this.parseExpression(this.element.value) != null) {
      return this.element.value;
    }
    if (this.datepicker.selectedDates.length > 0) {
      return this.datepicker.selectedDates[0];
    }
    return null;
  }

  // Returns true if the selected date is a relative expression, false otherwise.
  isRelativeDate() {
    return this.parseExpression(this.element.value) != null;
  }

  // Returns a factory function that returns the currently selected date as a
  // Date object, or null if no date is selected. The factory function is not
  // affected by changes to the picker state: for absolute dates, it always
  // returns the same value; for relative expressions, it always evaluates the
  // expression relative to the current time when called.
  getDateFactory() {
    if (this.parseExpression(this.element.value) != null) {
      const expression = this.element.value;
      return () => this.parseExpression(expression);
    }
    if (this.datepicker.selectedDates.length > 0) {
      const date = this.datepicker.selectedDates[0];
      return () => date;
    }
    return null;
  }

  destroy() {
    this.element.removeEventListener('keydown', this.handleKeyDown);
    this.element.removeEventListener('blur', this.handleOnBlur);
    this.datepicker.destroy();
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

  destroy() {
    this.fromPicker.destroy();
    this.toPicker.destroy();
  }
}

export { TimePicker, TimeRangePicker };
