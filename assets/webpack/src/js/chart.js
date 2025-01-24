import Plotly from 'plotly.js-dist';
import Tooltip from 'bootstrap/js/dist/tooltip';

import * as storage from './storage';

// Charts are self-conscious about their visibility to avoid being rendered
// if they are not visible to the user. To accomplish this, a global
// 'IntersectionObserver' is used to monitor the visibility of the charts and
// update their status accordingly.
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    entry.target.chart.setVisible(entry.isIntersecting);
  });
}, { threshold: 0.1 });

class Chart {
  constructor(container, metric, rangeFactory, refreshInterval, aggregator, step) {
    this.container = container;
    this.metric = metric;
    this.rangeFactory = rangeFactory;
    this.refreshInterval = refreshInterval;
    this.aggregator = aggregator;
    this.step = step;

    this.graph = null;
    this.visible = false;
    this.interval = null;

    this.x = [];
    this.y = [];

    observer.observe(this.container);
  }

  async init() {
    let metric;
    try {
      const [from, to] = this.rangeFactory();
      metric = await storage.getMetric(this.metric.id, from, to, this.step, this.aggregator);
    } catch (error) {
      this.setError(`Failed to fetch samples of a metric: ${error}`);
      return;
    }

    metric.samples.forEach(sample => {
      this.x.push(sample[0]);
      this.y.push(sample[1]);
    });

    const data = [
      {
        x: this.x,
        y: this.y,
        mode: this.x.length < 32 ? 'lines+markers' : 'lines', // TODO: dynamic, based on pixels available -- ! step changes won't be reflected!
        type: 'scatter',
        marker: { size: 4 },
        hovertemplate: '<b>X:</b> %{x|%Y-%m-%d %H:%M:%S}<br><b>Y:</b> %{y:.1f}<extra></extra>',
        connectgaps: false, // TODO: review
      }
    ];

    const layout = {
      autosize: true,
      title: {
        text: this.metric.name,
        font: {
          size: 14,
        },
        subtitle: {
          text: this.metric.description,
        },
      },
      margin: {
        l: 60,
        r: 10,
        b: 40,
        t: 40,
        pad: 5,
      },
      xaxis: {
        fixedrange: true,
        griddash: 'dash',
        range: [metric.from, new Date(metric.to.getTime() - metric.step * 1000)],
        autorange: false,
      },
      yaxis: {
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
        fixedrange: true,
        rangemode: 'normal',
        griddash: 'dash',
      },
    };

    const config = {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToRemove: [
        'zoom2d', 'pan2d', 'select2d', 'lasso2d', 'zoomIn2d', 'zoomOut2d',
        'autoScale2d', 'resetScale2d',
      ],
      toImageButtonOptions: {
        filename: `${varnishmon.storage.hostname} - ${this.metric.name}`,
        format: 'png',
      },
    };

    this.graph = this.container.querySelector('.graph');
    Plotly.newPlot(this.graph, data, layout, config);

    // Since the chart is being initialized, it is visible, so we can start the
    // refresh loop immediately.
    this.setupInterval();
  }

  //
  // Handlers.
  //

  async handleRefresh() {
    // If the chart is not visible, there is no need to update it. Refresh loops
    // are canceled when the chart is not visible, but this is a safety check.
    // Also, manual refreshes would be pointless for invisible charts.
    if (this.graph != null && this.visible) {
      let metric;
      try {
        const [from, to] = this.rangeFactory();
        metric = await storage.getMetric(this.metric.id, from, to, this.step, this.aggregator);
      } catch (error) {
        this.setError(`Failed to fetch samples of a metric: ${error}`);
        return;
      }

      this.x = [];
      this.y = [];
      metric.samples.forEach(sample => {
        this.x.push(sample[0]);
        this.y.push(sample[1]);
      });

      const update = {
        x: [this.x],
        y: [this.y],
      };

      const layout = {
        xaxis: {
          range: [metric.from, new Date(metric.to.getTime() - metric.step * 1000)],
        },
      };

      Plotly.update(this.graph, update, layout);
      // Plotly.extendTraces(this.graph, { x: [[newX]], y: [[newY]] }, [0]);

      // All good, clear any previous error.
      this.clearError();
    }
  }

  //
  // Public API.
  //

  setVisible(visible) {
    this.visible = visible;
    if (this.visible) {
      if (this.graph == null) {
        this.init();
      } else if (this.interval == null) {
        this.setupInterval();
      }
    } else {
      this.stopInterval();
    }
  }

  redraw(filter, verbosity, columns) {
    const hidden = !this.metric.name.includes(filter) || (verbosity === 'normal' && this.metric.debug);
    this.container.classList.toggle('d-none', hidden);

    this.container.classList.forEach(className => {
      if (className.startsWith('col-')) {
        this.container.classList.remove(className);
      }
    });
    this.container.classList.add(`col-${12 / columns}`);

    if (this.graph != null) {
      Plotly.Plots.resize(this.graph);
    }
  }

  setRefreshInterval(interval) {
    this.refreshInterval = interval;
    this.setupInterval();
  }

  refresh() {
    this.setupInterval();
    this.handleRefresh();
  }

  setAggregator(aggregator) {
    this.aggregator = aggregator;
    this.setupInterval();
    this.handleRefresh();
  }

  setStep(step) {
    this.step = step;
    this.setupInterval();
    this.handleRefresh();
  }

  destroy() {
    observer.unobserve(this.container);
    this.stopInterval();
    this.clearError();
    if (this.graph != null) {
      Plotly.purge(this.graph);
    }
  }

  //
  // Private helpers.
  //

  setupInterval() {
    this.stopInterval();

    if (this.visible) {
      if (this.graph != null && this.refreshInterval > 0) {
        this.interval = setInterval(this.handleRefresh.bind(this), this.refreshInterval * 1000);
      } else if (this.graph == null) {
        this.init();
      }
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
}

export default Chart;
