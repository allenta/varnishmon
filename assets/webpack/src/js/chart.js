import Plotly from 'plotly.js-dist';

import * as helpers from './helpers';

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
    this.date = new Date();

    observer.observe(this.container);
  }

  init() {
    for (let i = 0; i < 50; i++) {
      this.x.push(this.date.toISOString());
      this.y.push(Math.random() * 10);
      this.date = new Date(this.date.getTime() + 5000);
    }

    const data = [
      {
        x: this.x,
        y: this.y,
        type: 'scatter',
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
        l: 20,
        r: 20,
        b: 40,
        t: 25,
        pad: 4,
      },
      xaxis: {
        fixedrange: true,
      },
      yaxis: {
        title: 'eps',
        fixedrange: true,
      },
    };

    const config = {
      responsive: true,
      displaylogo: false,
      modeBarButtonsToRemove: [
        'zoom2d', 'pan2d', 'select2d', 'lasso2d', 'zoomIn2d', 'zoomOut2d',
        'autoScale2d', 'resetScale2d',
      ],
    };

    this.graph = this.container.querySelector('.graph');
    Plotly.newPlot(this.graph, data, layout, config);

    // Since the chart is being initialized, it is visible, so we can start the
    // refresh loop immediately.
    this.interval = setInterval(this.handleRefresh.bind(this), this.refreshInterval * 1000);
  }

  //
  // Handlers.
  //

  handleRefresh() {
    // If the chart is not visible, there is no need to update it. Refresh loops
    // are canceled when the chart is not visible, but this is a safety check.
    if (this.visible) {
      const [from, to] = this.rangeFactory();
      console.log('updating chart ' + this.metric.id + ': ' + from + ' => ' + to);

      this.date = new Date(this.date.getTime() + 5000);
      const newX = this.date.toISOString();
      const newY = Math.random() * 10;

      this.x.push(newX);
      this.y.push(newY);

      Plotly.extendTraces(this.graph, { x: [[newX]], y: [[newY]] }, [0]);
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
        this.interval = setInterval(this.handleRefresh.bind(this), this.refreshInterval * 1000);
      }
    } else {
      if (this.graph != null && this.interval != null) {
        clearInterval(this.interval);
      }
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
    console.log('setting refresh interval ' + interval + ' for chart ' + this.metric.id);
    this.refreshInterval = interval;
    if (this.graph != null && this.visible) {
      if (this.interval != null) {
        clearInterval(this.interval);
      }
      this.interval = setInterval(this.handleRefresh.bind(this), this.refreshInterval * 1000);
    }
  }

  refresh() {
    console.log('refreshing asap chart ' + this.metric.id);
    if (this.graph != null && this.visible) {
      if (this.interval != null) {
        clearInterval(this.interval);
      }
      this.interval = setInterval(this.handleRefresh.bind(this), this.refreshInterval * 1000);
      this.handleRefresh();
    }
  }

  setAggregator(aggregator) {
    console.log('setting aggregator ' + aggregator + ' for chart ' + this.metric.id);
    // TODO.
  }

  setStep(step) {
    console.log('setting step ' + step + ' for chart ' + this.metric.id);
    // TODO.
  }

  destroy() {
    console.log('destroying chart ' + this.metric.id);
    // TODO: what else?
    observer.unobserve(this.container);
    if (this.graph != null) {
      if (this.interval != null) {
        clearInterval(this.interval);
      }
      Plotly.purge(this.graph);
    }
  }

  //
  // Private helpers.
  //
}

export default Chart;
