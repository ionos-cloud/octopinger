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
	probeHealthGauge    *prometheus.GaugeVec
	probeRttMax         *prometheus.GaugeVec
	probeRttMin         *prometheus.GaugeVec
	probeRttMean        *prometheus.GaugeVec
	probeRttStddev      *prometheus.GaugeVec
	probePacketLossMin  *prometheus.GaugeVec
	probePacketLossMax  *prometheus.GaugeVec
	probePacketLossMean *prometheus.GaugeVec
	probeNodesTotal     *prometheus.GaugeVec
	probeNodesReports   *prometheus.GaugeVec
	nodesHealthGauge    *prometheus.GaugeVec
	errorsCounter       *prometheus.CounterVec
	icmpErrorsCounter   *prometheus.CounterVec
	clusterHealthGauge  *prometheus.GaugeVec
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

	m.probeRttStddev = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_probe_rtt_stddev",
			Help: "Standard deviation in round-trip time of the probe in this instance",
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

	m.nodesHealthGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "octopinger_nodes_health_total",
			Help: "Number of nodes seen as healthy/unhealthy from this instance",
		},
		[]string{
			"octopinger_node",
			"octopinger_status",
		},
	)

	m.errorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "octopinger_errors_total",
			Help: "The total number of errors per instance",
		},
		[]string{
			"octopinger_node",
			"octopinger_type",
		},
	)

	m.icmpErrorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "octopinger_icmp_errors_total",
			Help: "The total number of ICMP probe errors per instance",
		},
		[]string{
			"octopinger_node",
			"octopinger_type",
		},
	)

	return m
}

// Collect ...
func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.nodesHealthGauge.Collect(ch)
	m.probeRttMax.Collect(ch)
	m.probeRttMin.Collect(ch)
	m.probeRttStddev.Collect(ch)
	m.probeRttMean.Collect(ch)
	m.probePacketLossMax.Collect(ch)
	m.probePacketLossMin.Collect(ch)
	m.probePacketLossMean.Collect(ch)
	m.probeNodesTotal.Collect(ch)
	m.probeNodesReports.Collect(ch)
	m.errorsCounter.Collect(ch)
	m.icmpErrorsCounter.Collect(ch)
	m.clusterHealthGauge.Collect(ch)
	m.probeHealthGauge.Collect(ch)
}

// Describe ...
func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.nodesHealthGauge.Describe(ch)
	m.probeRttMax.Describe(ch)
	m.probeRttMin.Describe(ch)
	m.probeRttStddev.Describe(ch)
	m.probePacketLossMax.Describe(ch)
	m.probePacketLossMin.Describe(ch)
	m.probePacketLossMean.Describe(ch)
	m.probeRttMean.Describe(ch)
	m.probeNodesTotal.Describe(ch)
	m.probeNodesReports.Describe(ch)
	m.errorsCounter.Describe(ch)
	m.icmpErrorsCounter.Describe(ch)
	m.clusterHealthGauge.Describe(ch)
	m.probeHealthGauge.Describe(ch)
}

// Monitor ...
type Monitor struct {
	metrics *Metrics
}

// NewMonitor ...
func NewMonitor(metrics *Metrics) (*Monitor, error) {
	m := new(Monitor)
	m.metrics = metrics

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

// SetProbeRttStddev ...
func (m *Monitor) SetProbeRttStddev(instance, probe string, rtt float64) {
	m.metrics.probeRttStddev.WithLabelValues(instance, probe).Set(rtt)
}

// SetProbeRttMean ...
func (m *Monitor) SetProbeRttMean(instance, probe string, rtt float64) {
	m.metrics.probeRttMean.WithLabelValues(instance, probe).Set(rtt)
}

// SetProbePacketLossMin ...
func (m *Monitor) SetProbePacketlossMin(instance, probe string, percentage float64) {
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
