package config

import "github.com/prometheus/client_golang/prometheus"

func (cfg *Config) Metrics() *Metrics {
	return cfg.metrics
}

type Metrics struct {
	Registry *prometheus.Registry
}

func NewMetrics() *Metrics {
	result := &Metrics{
		Registry: prometheus.NewRegistry(),
	}

	// Metrics can be defined and registered here if useful, but in general is
	// a better practice to define them where they belong. That's generally
	// doable, unless shared state for multiple goroutines needing access to a
	// metric is not available.

	return result
}
