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
	nodesHealthGauge  *prometheus.GaugeVec
	errorsCounter     *prometheus.CounterVec
	icmpErrorsCounter *prometheus.CounterVec
}

// NewMetrics ...
func NewMetrics() *Metrics {
	m := new(Metrics)

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
}

// Describe ...
func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.nodesHealthGauge.Describe(ch)
	m.errorsCounter.Describe(ch)
	m.icmpErrorsCounter.Describe(ch)
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

	return m, nil
}

// CountErrors ...
func (m *Metrics) CountError(instance, errorType string) {
	m.errorsCounter.WithLabelValues(instance, errorType).Inc()
}

// CountICMPErrors ...
func (m *Metrics) CountICMPErrors(instance, errorType string) {
	m.icmpErrorsCounter.WithLabelValues(instance, errorType).Inc()
}
