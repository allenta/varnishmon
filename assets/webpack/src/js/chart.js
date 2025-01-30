import Plotly from 'plotly.js-dist';
import Tooltip from 'bootstrap/js/dist/tooltip';

import * as storage from './storage';
import * as helpers from './helpers';

// Charts are aware of their visibility to avoid rendering when not visible and
// to start/stop refresh loops accordingly. A global 'IntersectionObserver' is
// used to monitor their visibility status.
const intersectionObserver = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    entry.target.chart.handleVisibilityChange(entry.isIntersecting);
  });
}, { threshold: 0.1 });

// Charts need to be notified about changes in their container size to adjust
// the graph accordingly. This might require a full data refresh (e.g., less
// space is available, resulting in a smaller effective step, which requires
// fetching new down-sampled data), or just a re-render of the graph (e.g., more
// space is available, allowing the use of lines and markers instead of just
// lines).
const resizeObserver = new ResizeObserver((entries) => {
  entries.forEach(entry => {
    entry.target.chart.handleSizeChange();
  });
});

class Chart {
  constructor(container, metric, rangeFactory, refreshInterval, aggregator, step) {
    this.container = container;
    this.metric = metric;
    this.rangeFactory = rangeFactory;
    this.refreshInterval = refreshInterval;
    this.aggregator = aggregator;
    this.step = step;

    this.listeners = {};

    this.initializing = false;
    this.refreshing = false;

    this.debouncedHandleRefresh = helpers.debounce(this.handleRefresh.bind(this), 500);

    this.visible = false;
    this.interval = null;
    this.lastRefresh = 0;
    this.pendingRefresh = false;
    this.error = null;

    this.graph = {
      // A lazily initialized reference to the Plotly graph element.
      element: null,

      // The data currently displayed in the graph, as returned by the storage,
      // but slightly adjusted: hex bitmaps as integers, etc. Currently,
      // keeping this data around is not strictly necessary, but it might be
      // useful in the future to support incremental updates.
      x: null,
      y: null,

      // The step of the data currently displayed in the graph, as returned by
      // the storage. This is currently needed to estimate the optimal mode for
      // the Plotly data depending on the available space (i.e., number of
      // columns, zoom level, etc.).
      step: null,

      // The range in the X axis for the data currently contained in the graph,
      // as returned by the storage, but slightly adjusted: right boundary
      // adjusted one step back, etc.
      range: null,

      // The range for the X axis currently displayed in the graph, if zoomed.
      zoomRange: null,
    };

    intersectionObserver.observe(this.container);
    resizeObserver.observe(this.container);
  }

  async init() {
    // Prevent execution of multiple initializations at the same time, due to
    // the evaluation of the 'getMetric()' promise.
    if (this.initializing) {
      return;
    }
    this.initializing = true;
    try {
      // Fetch metric (adjusting the effective step if necessary, overriding the
      // selected aggregator if the metric is a bitmap, etc.) and immediately
      // process it to prepare the graph for the first render.
      this.processMetric(await this.getMetric());

      // Render the graph for the first time.
      this.graph.element = this.renderGraph();

      // Since the chart is being initialized, it is visible, so we can start
      // the refresh loop to keep the chart up-to-date.
      this.setupInterval();

      // Initialize refresh flags:
      //   - The 'last refresh' is needed because refresh loops are canceled
      //     when the chart is not visible. When it becomes visible again,
      //     the last refresh time is used to decide if the chart needs to be
      //     refreshed right away or if it can wait for the next scheduled
      //     refresh.
      //   - The 'pending refresh' is needed because other properties (e.g, step,
      //     aggregator, etc.) can change while the chart is not visible. When
      //     the chart becomes visible again, it needs to be refreshed right
      //     away to reflect these changes.
      this.lastRefresh = helpers.dateToUnix(new Date());
      this.pendingRefresh = false;
    } catch (error) {
      // On error, show some visual feedback to the user. Failing during the
      // initialization is a bit more complicated to recover automatically, so
      // we just show the error and let the user try to manually refresh the
      // chart.
      this.setError(`Failed to fetch samples of a metric: ${error}`);
    } finally {
      this.initializing = false;
    }
  }

  //
  // Handlers.
  //

  async handleRefresh() {
    // Prevent execution of multiple refreshes at the same time, due to the
    // evaluation of the 'getMetric()' promise.
    if (this.refreshing) {
      return;
    }
    this.refreshing = true;
    try {
      // Charts that are still pending initialization are not refreshed. This
      // will be done when they are created for the first time.
      if (this.graph.element != null) {
        // Refreshing an already initialized chart that is not visible is a
        // no-op, but a mark is used to trigger the pending refreshed as soon
        // as it becomes visible again. Same for failed refreshes of visible
        // charts.
        this.pendingRefresh = true;
        if (this.visible) {
          try {
            // Fetch metric (adjusting the effective step if necessary,
            // overriding the selected aggregator if the metric is a bitmap,
            // etc.) and immediately process it to prepare the graph for the
            // update.
            this.processMetric(await this.getMetric());

            // Update the graph with the new samples. For now we are just
            // re-fetching and re-rendering the whole graph, but in the future
            // we should be able to me more efficient. The server-side already
            // supports this kind of incremental updates.
            this.updateGraph(false);

            // All good, clear any previous error message and update the refresh
            // flags.
            this.clearError();
            this.lastRefresh = helpers.dateToUnix(new Date());
            this.pendingRefresh = false;
          } catch (error) {
            // On error, show some visual feedback to the user. The next refresh,
            // manual or automatic, will try to fetch the metric again and it
            // will eventually succeed and clear the error.
            this.setError(`Failed to fetch samples of a metric: ${error}`);
          }
        }
      }
    } finally {
      this.refreshing = false;
    }
  }

  handleVisibilityChange(visible) {
    this.visible = visible;
    if (this.visible) {
      if (this.graph.element == null) {
        this.init();
      } else {
        if (this.interval == null) {
          this.setupInterval();
        }
        if (this.pendingRefresh ||
            (this.refreshInterval > 0 && helpers.dateToUnix(new Date()) - this.lastRefresh > this.refreshInterval)) {
          // Note that the debounced refresh handler is used here.
          this.debouncedHandleRefresh();
        }
      }
    } else {
      this.stopInterval();
    }
  }

  handleSizeChange() {
    if (this.graph.element != null) {
      // Note that the debounced refresh handler is used here.
      this.debouncedHandleRefresh();
    }
  }

  handleGraphRelayout(event) {
    // Extract the zoom range of the graph, if any. This includes detecting the
    // no-zoom / zoom-reset event.
    if (event['xaxis.range[0]'] && event['xaxis.range[1]']) {
      this.graph.zoomRange = [
        new Date(event['xaxis.range[0]']),
        new Date(event['xaxis.range[1]']),
      ];
    } else if (event['xaxis.range'] && Array.isArray(event['xaxis.range']) && event['xaxis.range'].length === 2) {
      this.graph.zoomRange = null;
    } else {
      this.graph.zoomRange = null;
    }

    // If the zoom range goes beyond the current range, adjust it to the
    // current range.
    if (this.graph.zoomRange != null) {
      if (helpers.dateToUnix(this.graph.zoomRange[0]) < helpers.dateToUnix(this.graph.range[0])) {
        this.graph.zoomRange[0] = this.graph.range[0];
      }
      if (helpers.dateToUnix(this.graph.zoomRange[1]) > helpers.dateToUnix(this.graph.range[1])) {
        this.graph.zoomRange[1] = this.graph.range[1];
      }
    }

    // If the zoom range is too small, adjust it to a reasonable range.
    if (this.graph.zoomRange != null) {
      const difference = helpers.dateToUnix(this.graph.zoomRange[1]) - helpers.dateToUnix(this.graph.zoomRange[0]);
      if (difference < this.graph.step) {
        const center = helpers.dateToUnix(this.graph.zoomRange[0]) + Math.round(difference / 2);
        this.graph.zoomRange = [
          helpers.unixToDate(center - this.graph.step),
          helpers.unixToDate(center + this.graph.step),
        ];
      }
    }

    // Force a re-render of the graph. The zoom range is already applied, but we
    // might want to adjust other properties of the graph (e.g., the data mode),
    // or the adjusted zoom range itself.
    this.updateGraph(true);

    // Inform listeners about the zoom event.
    this.notifyEventListeners('zoom', {
      target: this,
      range: this.graph.zoomRange, // null if no zoom.
    });
  }

  //
  // Event listeners.
  //

  addEventListener(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }

  removeEventListener(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(listener => listener !== callback);
    }
  }

  notifyEventListeners(event, data) {
    if (this.listeners[event]) {
      this.listeners[event].forEach(callback => callback(data));
    }
  }

  //
  // Public API.
  //

  redraw(filter, verbosity, columns) {
    // Hide the chart if the metric name does not match the filter & verbosity.
    let hidden = verbosity === 'normal' && this.metric.debug;
    if (!hidden) {
      const terms = filter.split(/\s+/).filter(term => term.length > 0);
      if (terms.length > 0) {
        hidden = !terms.some(term => this.metric.name.includes(term));
      }
    }
    this.container.classList.toggle('d-none', hidden);

    // Rearrange CSS classes to adjust to the number of columns.
    this.container.classList.forEach(className => {
      if (className.startsWith('col-')) {
        this.container.classList.remove(className);
      }
    });
    this.container.classList.add(`col-${12 / columns}`);

    // No need to call 'handleRefresh()' here: visibility changes are managed by
    // 'handleVisibilityChange()', and size changes are managed by
    // 'handleSizeChange()'.
  }

  setRefreshInterval(interval) {
    this.refreshInterval = interval;
    if (this.visible && this.graph.element != null) {
      this.setupInterval();
    }
  }

  refresh() {
    if (this.graph.element != null) {
      if (this.visible) {
        this.setupInterval();
      }
      this.handleRefresh();
    }
  }

  setAggregator(aggregator) {
    this.aggregator = aggregator;
    if (this.graph.element != null) {
      if (this.visible) {
        this.setupInterval();
      }
      this.handleRefresh();
    }
  }

  setStep(step) {
    this.step = step;
    if (this.graph.element != null) {
      if (this.visible) {
        this.setupInterval();
      }
      this.handleRefresh();
    }
  }

  setZoomRange(range) {
    this.graph.zoomRange = range;

    // If the chart has already been initialized (whether visible or not),
    // update it to reflect the new range, the potential new data mode, etc.
    // Beware 'handleRefresh()' is not called here, as we are not fetching new
    // samples, just updating the graph with the current data.
    if (this.graph.element != null) {
      this.updateGraph(true);
    }
  }

  destroy() {
    intersectionObserver.unobserve(this.container);
    resizeObserver.unobserve(this.container);
    this.stopInterval();
    this.clearError();
    if (this.graph.element != null) {
      Plotly.purge(this.graph.element);
    }
  }

  //
  // Private helpers.
  //

  setupInterval() {
    this.stopInterval();
    if (this.refreshInterval > 0) {
      this.interval = setInterval(
        this.handleRefresh.bind(this), this.refreshInterval * 1000);
    }
  }

  stopInterval() {
    if (this.interval != null) {
      clearInterval(this.interval);
      this.interval = null;
    }
  }

  setError(error) {
    const card = this.container.querySelector('.card');
    const errorIcon = card.querySelector('.error-icon');

    if (this.error == null) {
      // Highlight card & show error icon.
      card.classList.add('border-danger');
      errorIcon.classList.remove('d-none');
    }

    // Create / update tooltip.
    const tooltipMessage = `${new Date().toISOString()}: ${error}`;
    let tooltip = Tooltip.getInstance(errorIcon);
    if (tooltip == null) {
      tooltip = new Tooltip(errorIcon, { title: tooltipMessage });
    } else {
      tooltip.setContent({ '.tooltip-inner': tooltipMessage });
    }

    // Update error variable.
    this.error = error;
  }

  clearError() {
    if (this.error != null) {
      // Unhighlight card & hide error icon.
      const card = this.container.querySelector('.card');
      card.classList.remove('border-danger');
      const errorIcon = card.querySelector('.error-icon');
      errorIcon.classList.add('d-none');

      // Destroy tooltip.
      const tooltip = Tooltip.getInstance(errorIcon);
      tooltip.dispose();

      // Update error variable.
      this.error = null;
    }
  }

  async getMetric() {
    // Fetch metric samples from the storage, adjusting the step if necessary,
    // and ignoring the selected aggregator if the metric is a bitmap. This
    // will be improved in the future adding more flexibility to control the
    // down-sampling of bitmap metrics.
    const loadingIcon = this.container.querySelector('.card .loading-icon');
    loadingIcon.classList.remove('d-none');
    try {
      const [from, to] = this.rangeFactory();
      const optimalStep = this.estimateOptimalStep(from, to);
      const aggregator = this.metric.flag === 'b' ? 'bit_and' : this.aggregator;
      return await storage.getMetric(this.metric.id, from, to, optimalStep, aggregator);
    } finally {
      loadingIcon.classList.add('d-none');
    }
  }

  estimateOptimalStep(from, to) {
    // Calculate the number of samples that would be required to cover the
    // whole 'from' - 'to' range with the selected step.
    const samples = (helpers.dateToUnix(to) - helpers.dateToUnix(from)) / this.step;

    // Estimate the number of samples that would fit reasonably within the
    // graph, estimated as 90% of the container width. Ideally we'd prefer to
    // let Plotly to decide this, but apparently it's not possible.
    const containerWidth = this.container.clientWidth;
    const maxSamples = Math.floor(0.9 * containerWidth);

    // Stop here if the maximum number of samples cannot be calculated (most
    // likely because the container is not visible due to a race condition).
    if (maxSamples <= 0) {
      throw new Error('Failed to estimate the optimal step');
    }

    // If the number of samples required to cover the whole range is less than
    // the number of samples that would fit in the graph, we can use the current
    // step. Otherwise, we need to calculate a new step that would fit the
    // graph, letting the server-side to apply the down-sampling using the
    // current aggregator.
    if (samples <= maxSamples) {
      return this.step;
    }

    // Calculate the optimal step as a multiple of the current step that would
    // fit the graph.
    return Math.ceil(samples / maxSamples) * this.step;
  }

  processMetric(metric) {
    // Prepare X & Y data for Plotly.
    this.graph.x = [];
    this.graph.y = [];
    metric.samples.forEach(sample => {
      this.graph.x.push(sample[0]);
      // Bitmap metrics are returned as an hex string. For now, we represent the
      // number of bits set to 1 in the bitmap as the Y value. This will be
      // improved in the future using a different visualization for bitmap
      // metrics.
      this.graph.y.push(
        this.metric.flag === 'b' ?
          BigInt(`0x${sample[1]}`).toString(2).split('').filter(bit => bit === '1').length :
          sample[1]);
    });

    // Store the step (already adjusted to be optimal for the space available)
    // as returned by the storage.
    this.graph.step = metric.step;

    // Calculate & store the range for the X axis. This may change during
    // zoom events, and we need to know the original range to reset it.
    this.graph.range = [
      metric.from,
      new Date(metric.to.getTime() - metric.step * 1000),
    ];

    // If the effective step is different from the selected step, show in the
    // card the factor by which the selected step has been adjusted.
    const stepFactor = this.container.querySelector('.card .step-factor');
    if (this.graph.step !== this.step) {
      const factor = Math.round(this.graph.step / this.step);
      stepFactor.innerHTML =
        `<i class="fa-solid fa-arrows-left-right-to-line"></i> ${factor}x`;
    } else {
      stepFactor.innerHTML = '';
    }
  }

  renderGraph() {
    // Decide range to be used in the X axis.
    const range = this.graph.zoomRange != null ? this.graph.zoomRange : this.graph.range;

    // Prepare data for Plotly.
    const data = [
      {
        x: this.graph.x,
        y: this.graph.y,
        type: 'scatter',
        mode: this.estimatePlotlyDataMode(...range, this.graph.step),
        marker: { size: 4 },
        hovertemplate: '<b>X:</b> %{x|%Y-%m-%d %H:%M:%S}<br><b>Y:</b> %{y:,.1f}<extra></extra>',
        connectgaps: false,
        line: { shape: 'linear', width: 2 },
      }
    ];

    // Prepare layout for Plotly.
    const layout = {
      autosize: true,
      title: {
        text: this.metric.name,
        font: { size: 14 },
        subtitle: { text: this.metric.description },
      },
      margin: { l: 60, r: 10, b: 40, t: 40, pad: 5 },
      xaxis: {
        fixedrange: false,
        griddash: 'dash',
        range: Array.from(range), // Beware the array needs to be cloned.
        autorange: false,
      },
      yaxis: {
        fixedrange: true,
        griddash: 'dash',
        rangemode: 'normal',
        // tickformat: '.1s',
        title: (() => {
          if (this.metric.flag === 'c') {
            if (this.metric.format === 'd') {
              return 'seconds';
            } else if (this.metric.format === 'B') {
              return 'Bps';
            }
            return 'eps';
          } else if (this.metric.flag === 'g') {
            if (this.metric.format === 'd') {
              return 'seconds';
            } else if (this.metric.format === 'B') {
              return 'bytes';
            }
          }
          return '';
        })(),
      },
    };

    // Prepare configuration for Plotly.
    const config = {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToRemove: [
        'zoom2d', 'pan2d', 'select2d', 'lasso2d', 'autoScale2d',
      ],
      toImageButtonOptions: {
        filename: `${varnishmon.storage.hostname} - ${this.metric.name}`,
        format: 'png',
      },
      scrollZoom: false,
    };

    // Render the graph.
    const graphElement = this.container.querySelector('.graph');
    Plotly.newPlot(graphElement, data, layout, config);

    // Handle Plotly events.
    graphElement.on('plotly_relayout', this.handleGraphRelayout.bind(this));

    // Done!
    return graphElement;
  }

  updateGraph(sameData) {
    // Decide range to be used in the X axis.
    const range = this.graph.zoomRange != null ? this.graph.zoomRange : this.graph.range;

    // Prepare data for Plotly.
    const data = {
      mode: this.estimatePlotlyDataMode(...range, this.graph.step),
    };
    if (!sameData) {
      data.x = [this.graph.x];
      data.y = [this.graph.y];
    }

    // Prepare layout for Plotly.
    const layout = {
      xaxis: {
        range: Array.from(range), // Beware the array needs to be cloned.
      },
    };

    // Update the graph!
    Plotly.update(this.graph.element, data, layout);
  }

  estimatePlotlyDataMode(from, to, step) {
    const samples = (helpers.dateToUnix(to) - helpers.dateToUnix(from)) / step;
    const containerWidth = this.container.clientWidth;
    const minSpacing = 6;
    const maxSamples = Math.floor(0.9 * containerWidth / minSpacing);
    if (samples > maxSamples) {
      return 'lines';
    }
    return 'lines+markers';
  }
}

export default Chart;
