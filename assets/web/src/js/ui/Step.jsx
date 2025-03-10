import PropTypes from 'prop-types';
import React from 'react';

import * as config from '../config';
import * as helpers from '../helpers';

export function Step({ step, setStep }) {
  const stepId = React.useId();
  const [localStep, setLocalStep] = React.useState(step);

  const onChange = (event) => {
    const value = event.target.value;
    setLocalStep(value);
  };

  const onBlur = (event) => {
    let value = event.target.value;
    const minimum = config.getMinimumStep();

    if (value === '' || parseInt(value, 10) < minimum) {
      helpers.notify('error', `Step must be at least ${minimum} seconds`);
      value = minimum;
    } else {
      value = parseInt(value, 10);
    }

    config.setStep(value);
    setLocalStep(value);
    setStep(value);
  };

  return (
    <>
      <label className="form-label" htmlFor={stepId}>
        Step
      </label>
      <div className="input-group">
        <span className="input-group-text">
          <i className="fa-solid fa-arrows-left-right-to-line"></i>
        </span>
        <input
          type="number"
          className="form-control"
          id={stepId}
          min={config.getMinimumStep()}
          value={localStep}
          onChange={onChange}
          onBlur={onBlur}
        />
      </div>
    </>
  );
}

Step.propTypes = {
  step: PropTypes.number.isRequired,
  setStep: PropTypes.func.isRequired,
};
