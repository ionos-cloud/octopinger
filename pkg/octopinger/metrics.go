package octopinger

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	DefaultRegistry                         = NewRegistry()
	DefaultRegisterer prometheus.Registerer = DefaultRegistry
	DefaultGatherer   prometheus.Gatherer   = DefaultRegistry
)

// Registry ...
type Registry struct {
	*prometheus.Registry
}

// NewRegistry ...
func NewRegistry() *Registry {
	r := prometheus.NewRegistry()

	r.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
	)

	return &Registry{Registry: r}
}

// Handler returns a HTTP handler for this registry. Should be registered at "/metrics".
func (r *Registry) Handler() http.Handler {
	return promhttp.InstrumentMetricHandler(r, promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
}

var (
	// DefaultMetrics ...
	DefaultMetrics = NewMetrics()
)

type Metrics struct {
	nodesHealthGauge   *prometheus.GaugeVec
	errorsCounter      *prometheus.CounterVec
	icmpErrorsCounter  *prometheus.CounterVec
	clusterHealthGauge *prometheus.GaugeVec
}

// NewMetrics ...
func NewMetrics() *Metrics {
	m := new(Metrics)

	m.clusterHealthGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_cluster_health_total",
			Help: "1 if all check pass, 0 otherwise",
		},
		[]string{
			"instance",
		},
	)

	m.nodesHealthGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_nodes_health_total",
			Help: "Number of nodes seen as healthy/unhealthy from this instance",
		},
		[]string{
			"instance",
			"status",
		},
	)

	m.errorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "octopinger_errors_total",
			Help: "The total number of errors per instance",
		},
		[]string{
			"instance",
			"type",
		},
	)

	m.icmpErrorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "octopinger_icmp_errors_total",
			Help: "The total number of ICMP probe errors per instance",
		},
		[]string{
			"instance",
			"type",
		},
	)

	return m
}

// Collect ...
func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.nodesHealthGauge.Collect(ch)
	m.errorsCounter.Collect(ch)
	m.icmpErrorsCounter.Collect(ch)
	m.clusterHealthGauge.Collect(ch)
}

// Describe ...
func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.nodesHealthGauge.Describe(ch)
	m.errorsCounter.Describe(ch)
	m.icmpErrorsCounter.Describe(ch)
	m.clusterHealthGauge.Describe(ch)
}

// Monitor ...
type Monitor struct {
	metrics *Metrics
}

// NewMonitor ...
func NewMonitor(metrics *Metrics) (*Monitor, error) {
	m := new(Monitor)
	m.metrics = metrics

	_, err := m.metrics.nodesHealthGauge.GetMetricWith(prometheus.Labels{"instance": "", "status": ""})
	if err != nil {
		return nil, err
	}

	m.metrics.errorsCounter.GetMetricWith(prometheus.Labels{"instance": "", "type": ""})
	if err != nil {
		return nil, err
	}

	m.metrics.clusterHealthGauge.GetMetricWith(prometheus.Labels{"instance": ""})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// SetClusterHealth ...
func (m *Monitor) SetClusterHealthy(instance string, healthy bool) {
	value := 1.0
	if !healthy {
		value = 0
	}

	m.metrics.clusterHealthGauge.WithLabelValues(instance).Set(value)
}

// CountErrors ...
func (m *Monitor) CountError(instance, errorType string) {
	m.metrics.errorsCounter.WithLabelValues(instance, errorType).Inc()
}

// CountICMPErrors ...
func (m *Monitor) CountICMPErrors(instance, errorType string) {
	m.metrics.icmpErrorsCounter.WithLabelValues(instance, errorType).Inc()
}
