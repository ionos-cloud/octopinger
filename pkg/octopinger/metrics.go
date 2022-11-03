package octopinger

import (
	"net/http"
	"sync"

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

// Metrics ...
type Metrics struct {
	probeHealthGauge    *prometheus.GaugeVec
	probeRttMax         *prometheus.GaugeVec
	probeRttMin         *prometheus.GaugeVec
	probeRttMean        *prometheus.GaugeVec
	probePacketLossMin  *prometheus.GaugeVec
	probePacketLossMax  *prometheus.GaugeVec
	probePacketLossMean *prometheus.GaugeVec
	probeNodesTotal     *prometheus.GaugeVec
	probeNodesReports   *prometheus.GaugeVec
	probeDNSTotal       *prometheus.GaugeVec
	probeDNSSuccess     *prometheus.GaugeVec
	probeDNSFailure     *prometheus.GaugeVec
	probeDNSError       *prometheus.GaugeVec
}

// NewMetrics ...
func NewMetrics() *Metrics {
	m := new(Metrics)

	m.probeDNSTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_dns_total",
			Help: "Total number of DNS records to probe.",
		},
		[]string{
			"octopinger_node",
		},
	)

	m.probeDNSSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_dns_success",
			Help: "Number of successful probed DNS records.",
		},
		[]string{
			"octopinger_node",
		},
	)

	m.probeDNSError = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_dns_error",
			Help: "Number of errored probed DNS records.",
		},
		[]string{
			"octopinger_node",
		},
	)

	m.probeDNSFailure = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_dns_failure",
			Help: "Number of failed probed DNS records.",
		},
		[]string{
			"octopinger_node",
		},
	)

	m.probeDNSError = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_dns_success",
			Help: "Total number of DNS records to probe.",
		},
		[]string{
			"octopinger_node",
		},
	)

	m.probeNodesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_nodes_total",
			Help: "Total number of probed nodes",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	m.probeNodesReports = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_nodes_reports",
			Help: "Number of nodes that reported results",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	m.probeRttMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_rtt_min",
			Help: "Min round-trip time of the probe in this instance",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	m.probeRttMean = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_rtt_mean",
			Help: "Mean round-trip time of the probe in this instance",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	m.probeRttMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_rtt_max",
			Help: "Max round-trip time of the probe in this instance",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	m.probePacketLossMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_loss_min",
			Help: "Min percentage of lost packets",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	m.probePacketLossMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_loss_max",
			Help: "Max percentage of lost packets",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	m.probePacketLossMean = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_loss_mean",
			Help: "Mean percentage of lost packets",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	m.probeHealthGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_health_total",
			Help: "Health based on individual probes",
		},
		[]string{
			"octopinger_node",
			"octopinger_probe",
		},
	)

	return m
}

// Collect ...
func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.probeRttMax.Collect(ch)
	m.probeRttMin.Collect(ch)
	m.probeRttMean.Collect(ch)
	m.probePacketLossMax.Collect(ch)
	m.probePacketLossMin.Collect(ch)
	m.probePacketLossMean.Collect(ch)
	m.probeNodesTotal.Collect(ch)
	m.probeNodesReports.Collect(ch)
	m.probeHealthGauge.Collect(ch)
	m.probeDNSTotal.Collect(ch)
	m.probeDNSSuccess.Collect(ch)
	m.probeDNSError.Collect(ch)
}

// Describe ...
func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.probeRttMax.Describe(ch)
	m.probeRttMin.Describe(ch)
	m.probePacketLossMax.Describe(ch)
	m.probePacketLossMin.Describe(ch)
	m.probePacketLossMean.Describe(ch)
	m.probeRttMean.Describe(ch)
	m.probeNodesTotal.Describe(ch)
	m.probeNodesReports.Describe(ch)
	m.probeHealthGauge.Describe(ch)
	m.probeDNSTotal.Describe(ch)
	m.probeDNSSuccess.Describe(ch)
	m.probeDNSError.Describe(ch)
}

// Monitor ...
type Monitor struct {
	metrics *Metrics

	sync.Mutex
}

// Gatherer ...
type Gatherer interface {
	// Gather ...
	Gather(collector Collector)
}

// NewMonitor ...
func NewMonitor(metrics *Metrics) *Monitor {
	m := new(Monitor)
	m.metrics = metrics

	return m
}

// Gather ...
func (m *Monitor) Gather(collector Collector) {
	m.Lock()
	defer m.Unlock()

	ch := make(chan Metric)
	defer func() { close(ch) }()

	go func() {
		for metric := range ch {
			_ = metric.Write(m)
		}
	}()

	collector.Collect(ch)
}

// SetProbeHealth ...
func (m *Monitor) SetProbeHealth(instance, probe string, healthy bool) {
	value := 1.0
	if !healthy {
		value = 0
	}
	m.metrics.probeHealthGauge.WithLabelValues(instance, probe).Set(value)
}

// SetProbeNodesTotal ...
func (m *Monitor) SetProbeNodesTotal(instance, probe string, num float64) {
	m.metrics.probeNodesTotal.WithLabelValues(instance, probe).Set(num)
}

// SetProbeNodesReports ...
func (m *Monitor) SetProbeNodesReports(instance, probe string, num float64) {
	m.metrics.probeNodesReports.WithLabelValues(instance, probe).Set(num)
}

// SetProbeRttMax ...
func (m *Monitor) SetProbeRttMax(instance, probe string, rtt float64) {
	m.metrics.probeRttMax.WithLabelValues(instance, probe).Set(rtt)
}

// SetProbeRttMin ...
func (m *Monitor) SetProbeRttMin(instance, probe string, rtt float64) {
	m.metrics.probeRttMin.WithLabelValues(instance, probe).Set(rtt)
}

// SetProbeRttMean ...
func (m *Monitor) SetProbeRttMean(instance, probe string, rtt float64) {
	m.metrics.probeRttMean.WithLabelValues(instance, probe).Set(rtt)
}

// SetProbePacketLossMin ...
func (m *Monitor) SetProbePacketLossMin(instance, probe string, percentage float64) {
	m.metrics.probePacketLossMin.WithLabelValues(instance, probe).Set(percentage)
}

// SetProbePacketLossMax ...
func (m *Monitor) SetProbePacketLossMax(instance, probe string, percentage float64) {
	m.metrics.probePacketLossMax.WithLabelValues(instance, probe).Set(percentage)
}

// SetProbePacketLossMean ...
func (m *Monitor) SetProbePacketLossMean(instance, probe string, percentage float64) {
	m.metrics.probePacketLossMax.WithLabelValues(instance, probe).Set(percentage)
}

// SetProbeDNSTotal ...
func (m *Monitor) SetProbeDNSTotal(instance string, float float64) {
	m.metrics.probeDNSTotal.WithLabelValues(instance).Set(float)
}

// SetProbeDNSError ...
func (m *Monitor) SetProbeDNSError(instance string, float float64) {
	m.metrics.probeDNSError.WithLabelValues(instance).Set(float)
}

// SetProbeDNSSuccess ...
func (m *Monitor) SetProbeDNSSuccess(instance string, float float64) {
	m.metrics.probeDNSSuccess.WithLabelValues(instance).Set(float)
}

// SetProbeDNSFailure ...
func (m *Monitor) SetProbeDNSFailure(instance string, float float64) {
	m.metrics.probeDNSFailure.WithLabelValues(instance).Set(float)
}
