package octopinger

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/chenjiandongx/pinger"
	"github.com/ionos-cloud/octopinger/api/v1alpha1"
	"github.com/montanaflynn/stats"
)

const (
	defaultPacketLossThreshold = 0.05
	defaultTimeout             = 5 * time.Second
)

type maxRtt struct {
	values []float64

	probeName string
	nodeName  string

	Metric
	Collector
}

// Write ...
func (m *maxRtt) Write(monitor *Monitor) error {
	max, err := stats.Max(m.values)
	if err != nil {
		return err
	}

	monitor.SetProbeRttMax(m.nodeName, m.probeName, max)

	return nil
}

// NewMaxRtt ...
func NewMaxRtt(probeName, nodeName string) *maxRtt {
	return &maxRtt{
		probeName: probeName,
		nodeName:  nodeName,
	}
}

type meanRtt struct {
	values []float64

	probeName string
	nodeName  string

	Metric
	Collector
}

// Write ...
func (m *meanRtt) Write(monitor *Monitor) error {
	mean, err := stats.Mean(m.values)
	if err != nil {
		return err
	}

	monitor.SetProbeRttMean(m.nodeName, m.probeName, mean)

	return nil
}

// NewMeanRtt ...
func NewMeanRtt(probeName, nodeName string) *meanRtt {
	return &meanRtt{
		probeName: probeName,
		nodeName:  nodeName,
	}
}

type minRtt struct {
	values []float64

	probeName string
	nodeName  string

	Metric
	Collector
}

// Write ...
func (m *minRtt) Write(monitor *Monitor) error {
	min, err := stats.Min(m.values)
	if err != nil {
		return err
	}

	monitor.SetProbeRttMin(m.nodeName, m.probeName, min)

	return nil
}

// NewMinRtt ...
func NewMinRtt(probeName, nodeName string) *minRtt {
	return &minRtt{
		probeName: probeName,
		nodeName:  nodeName,
	}
}

type totalNumber struct {
	value float64

	probeName string
	nodeName  string

	Metric
	Collector
}

// Write ...
func (m *totalNumber) Write(monitor *Monitor) error {
	monitor.SetProbeNodesTotal(m.nodeName, m.probeName, m.value)

	return nil
}

// NewTotalNumber ...
func NewTotalNumber(probeName, nodeName string) *totalNumber {
	return &totalNumber{
		probeName: probeName,
		nodeName:  nodeName,
	}
}

type reportNumber struct {
	value float64

	probeName string
	nodeName  string

	Metric
	Collector
}

// Write ...
func (m *reportNumber) Write(monitor *Monitor) error {
	monitor.SetProbeNodesReports(m.nodeName, m.probeName, m.value)

	return nil
}

// NewReportNumber ...
func NewReportNumber(probeName, nodeName string) *reportNumber {
	return &reportNumber{
		probeName: probeName,
		nodeName:  nodeName,
	}
}

type packetLoss struct {
	values []float64

	probeName string
	nodeName  string

	Metric
	Collector
}

// Write ...
func (m *packetLoss) Write(monitor *Monitor) error {
	mean, err := stats.Mean(m.values)
	if err != nil {
		return err
	}
	monitor.SetProbePacketLossMean(m.nodeName, m.probeName, mean)

	max, err := stats.Mean(m.values)
	if err != nil {
		return err
	}
	monitor.SetProbePacketLossMax(m.nodeName, m.probeName, max)

	min, err := stats.Min(m.values)
	if err != nil {
		return err
	}
	monitor.SetProbePacketLossMin(m.nodeName, m.probeName, min)

	sum, err := stats.Sum(m.values)
	if err != nil {
		return err
	}
	monitor.SetProbePacketLossTotal(m.nodeName, m.probeName, sum)

	return nil
}

// NewPacketLoss ...
func NewPacketLoss(probeName, nodeName string) *packetLoss {
	return &packetLoss{
		probeName: probeName,
		nodeName:  nodeName,
	}
}

// AddMaxRtt ...
func (i *icmpProbe) AddMaxRtt(value float64) {
	i.Lock()
	defer i.Unlock()

	i.maxRtt.values = append(i.maxRtt.values, value)
}

// AddMinRtt ...
func (i *icmpProbe) AddMinRtt(value float64) {
	i.Lock()
	defer i.Unlock()

	i.minRtt.values = append(i.minRtt.values, value)
}

// AddMeanRtt ...
func (i *icmpProbe) AddMeanRtt(value float64) {
	i.Lock()
	defer i.Unlock()

	i.meanRtt.values = append(i.meanRtt.values, value)
}

// AddPacketLoss ...
func (i *icmpProbe) AddPacketLoss(value float64) {
	i.Lock()
	defer i.Unlock()

	i.packetLoss.values = append(i.packetLoss.values, value)
}

// SetTotalNumber ...
func (i *icmpProbe) SetTotalNumber(value float64) {
	i.Lock()
	defer i.Unlock()

	i.totalNumber.value = value
}

// IncReportNumber ...
func (i *icmpProbe) IncReportNumber() {
	i.Lock()
	defer i.Unlock()

	i.reportNumber.value += 1
}

// Collect ...
func (m *maxRtt) Collect(ch chan<- Metric) {
	ch <- m
}

// Collect ...
func (m *minRtt) Collect(ch chan<- Metric) {
	ch <- m
}

// Collect ...
func (m *meanRtt) Collect(ch chan<- Metric) {
	ch <- m
}

// Collect ...
func (m *packetLoss) Collect(ch chan<- Metric) {
	ch <- m
}

// Collect ...
func (m *totalNumber) Collect(ch chan<- Metric) {
	ch <- m
}

// Collect ...
func (m *reportNumber) Collect(ch chan<- Metric) {
	ch <- m
}

type icmpProbe struct {
	opts *Opts

	name     string
	nodeName string

	maxRtt       *maxRtt
	minRtt       *minRtt
	meanRtt      *meanRtt
	totalNumber  *totalNumber
	reportNumber *reportNumber
	packetLoss   *packetLoss

	timeout         time.Duration
	reportThreshold float64

	Collector
	sync.RWMutex
}

func (i *icmpProbe) configure(c *v1alpha1.Config) error {
	if c.ICMP.NodePacketLossThreshold != "" {
		s, err := strconv.ParseFloat(c.ICMP.NodePacketLossThreshold, 64)
		if err != nil {
			return err
		}

		i.reportThreshold = s
	}

	if c.ICMP.Timeout != "" {
		s, err := time.ParseDuration(c.ICMP.Timeout)
		if err != nil {
			return err
		}

		i.timeout = s
	}

	return nil
}

// NewICMPProbe ...
func NewICMPProbe(nodeName string, opts ...Opt) *icmpProbe {
	options := new(Opts)
	options.Configure(opts...)

	p := new(icmpProbe)
	p.opts = options
	p.nodeName = nodeName
	p.name = "icmp"

	p.timeout = defaultTimeout
	p.reportThreshold = defaultPacketLossThreshold

	p.Reset()

	return p
}

// Name ...
func (p *icmpProbe) Name() string {
	return p.name
}

// Reset ...
func (p *icmpProbe) Reset() {
	p.maxRtt = NewMaxRtt(p.name, p.nodeName)
	p.meanRtt = NewMeanRtt(p.name, p.nodeName)
	p.minRtt = NewMinRtt(p.name, p.nodeName)
	p.packetLoss = NewPacketLoss(p.name, p.nodeName)
	p.reportNumber = NewReportNumber(p.name, p.nodeName)
	p.totalNumber = NewTotalNumber(p.name, p.nodeName)
}

// Collect ...
func (i *icmpProbe) Collect(ch chan<- Metric) {
	i.maxRtt.Collect(ch)
	i.meanRtt.Collect(ch)
	i.minRtt.Collect(ch)
	i.packetLoss.Collect(ch)
	i.reportNumber.Collect(ch)
	i.totalNumber.Collect(ch)
}

// Do ...
func (i *icmpProbe) Do(ctx context.Context, metrics Gatherer) func() error {
	return func() error {
		err := i.configure(i.opts.config)
		if err != nil {
			return err
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		loaders := []NodeLoader{
			NodesLoader(i.opts.configPath),
		}

		filters := []NodeFilter{
			FilterIP(i.opts.hostIP),
		}

		nodeList := NewNodeList(loaders, filters...)

		for {
			select {
			case <-ctx.Done():
			case <-ticker.C:
				nodes, err := nodeList.Load()
				if err != nil {
					return err
				}

				opt := *pinger.DefaultICMPPingOpts
				opt.Interval = func() time.Duration { return 100 * time.Millisecond }
				opt.PingCount = 5
				opt.PingTimeout = i.timeout

				i.Reset()
				i.SetTotalNumber(float64(len(nodes)))

				stats, err := pinger.ICMPPing(&opt, nodes...)
				if err != nil {
					return err
				}

				for _, stat := range stats {
					if stat.PktLossRate < i.reportThreshold {
						i.IncReportNumber()
					}

					i.AddMaxRtt(float64(stat.Worst.Microseconds()))
					i.AddMinRtt(float64(stat.Best.Microseconds()))
					i.AddMeanRtt(float64(stat.Mean.Microseconds()))
					i.AddPacketLoss(float64(stat.PktLossRate))
				}

				metrics.Gather(i)
				ticker.Reset(1 * time.Second)

				continue
			}
		}
	}
}
