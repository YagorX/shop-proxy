package observability

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	ConnectionsTotal      prometheus.Counter
	ConnectionsActive     prometheus.Gauge
	BytesUpstreamTotal    prometheus.Counter
	BytesDownstreamTotal  prometheus.Counter
	CopyErrorsTotal       *prometheus.CounterVec
	ConfiguredDelayMillis prometheus.Gauge
}

var (
	metricsInstance *Metrics
	metricsOnce     sync.Once
)

func MustMetrics() *Metrics {
	metricsOnce.Do(func() {
		metricsInstance = newMetrics()
	})
	return metricsInstance
}

func newMetrics() *Metrics {
	m := &Metrics{
		ConnectionsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "proxy",
				Subsystem: "tcp",
				Name:      "connections_total",
				Help:      "Total number of accepted proxy TCP connections.",
			},
		),
		ConnectionsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "proxy",
				Subsystem: "tcp",
				Name:      "connections_active",
				Help:      "Current number of active proxy TCP connections.",
			},
		),
		BytesUpstreamTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "proxy",
				Subsystem: "traffic",
				Name:      "bytes_upstream_total",
				Help:      "Total number of bytes forwarded from client to upstream.",
			},
		),
		BytesDownstreamTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: "proxy",
				Subsystem: "traffic",
				Name:      "bytes_downstream_total",
				Help:      "Total number of bytes forwarded from upstream to client.",
			},
		),
		CopyErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "proxy",
				Subsystem: "tcp",
				Name:      "copy_errors_total",
				Help:      "Total number of proxy copy loop errors.",
			},
			[]string{"direction"},
		),
		ConfiguredDelayMillis: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "proxy",
				Subsystem: "faults",
				Name:      "delay_milliseconds",
				Help:      "Currently configured fixed proxy delay in milliseconds.",
			},
		),
	}

	prometheus.MustRegister(
		m.ConnectionsTotal,
		m.ConnectionsActive,
		m.BytesUpstreamTotal,
		m.BytesDownstreamTotal,
		m.CopyErrorsTotal,
		m.ConfiguredDelayMillis,
	)

	return m
}
