package octopinger

import (
	"context"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/montanaflynn/stats"
	"golang.org/x/sync/errgroup"
)

// Collector ...
type Collector interface {
	// Collect ...
	Collect(ch chan<- Metric)
}

type statistics struct {
	maxRtt       *maxRtt
	minRtt       *minRtt
	meanRtt      *meanRtt
	totalNumber  *totalNumber
	reportNumber *reportNumber
	packetLoss   *packetLoss

	probeName string
	nodeName  string

	Collector
	sync.RWMutex
}

// Collect ...
func (s *statistics) Collect(ch chan<- Metric) {
	s.maxRtt.Collect(ch)
	s.minRtt.Collect(ch)
	s.meanRtt.Collect(ch)
	s.totalNumber.Collect(ch)
	s.reportNumber.Collect(ch)
	s.packetLoss.Collect(ch)
}

// NewStatistics ...
func NewStatistics(probeName, nodeName string) *statistics {
	s := new(statistics)

	s.nodeName = nodeName
	s.probeName = probeName
	s.maxRtt = NewMaxRtt(s.probeName, s.nodeName)
	s.minRtt = NewMinRtt(s.probeName, s.nodeName)
	s.meanRtt = NewMeanRtt(s.probeName, s.nodeName)
	s.totalNumber = NewTotalNumber(s.probeName, s.nodeName)
	s.reportNumber = NewReportNumber(s.probeName, s.nodeName)
	s.packetLoss = NewPacketLoss(s.probeName, s.nodeName)

	return s
}

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
func (s *statistics) AddMaxRtt(value float64) {
	s.Lock()
	defer s.Unlock()

	s.maxRtt.values = append(s.maxRtt.values, value)
}

// AddMinRtt ...
func (s *statistics) AddMinRtt(value float64) {
	s.Lock()
	defer s.Unlock()

	s.minRtt.values = append(s.minRtt.values, value)
}

// AddMeanRtt ...
func (s *statistics) AddMeanRtt(value float64) {
	s.Lock()
	defer s.Unlock()

	s.meanRtt.values = append(s.meanRtt.values, value)
}

// AddPacketLoss ...
func (s *statistics) AddPacketLoss(value float64) {
	s.Lock()
	defer s.Unlock()

	s.packetLoss.values = append(s.packetLoss.values, value)
}

// SetTotalNumber ...
func (s *statistics) SetTotalNumber(value float64) {
	s.Lock()
	defer s.Unlock()

	s.totalNumber.value = value
}

// IncReportNumber ...
func (s *statistics) IncReportNumber() {
	s.Lock()
	defer s.Unlock()

	s.reportNumber.value += 1
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

// Metric
type Metric interface {
	// Write ...
	Write(m *Monitor) error
}

// Probe ...
type Probe interface {
	// Do ...
	Do(ctx context.Context, monitor Monitor) error

	Collector
}

type icmpProbe struct {
	opts  *Opts
	stats *statistics

	name     string
	nodeName string
}

// Collect ...
func (i *icmpProbe) Collect(ch chan<- Metric) {
	i.stats.Collect(ch)
}

// NewICMPProbe ...
func NewICMPProbe(nodeName string, opts ...Opt) *icmpProbe {
	options := new(Opts)
	options.Configure(opts...)

	p := new(icmpProbe)
	p.opts = options
	p.nodeName = nodeName
	p.name = "icmp"

	return p
}

// Name ...
func (p *icmpProbe) Name() string {
	return p.name
}

// Do ...
func (i *icmpProbe) Do(ctx context.Context, metrics Gatherer) func() error {
	return func() error {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		loaders := []NodeLoader{
			NodesLoader(i.opts.configPath),
		}

		filters := []NodeFilter{
			FilterIP(i.opts.podIP),
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

				g, gctx := errgroup.WithContext(ctx)
				g.SetLimit(10)

				stats := NewStatistics(i.name, i.nodeName)
				stats.SetTotalNumber(float64(len(nodes)))

				for _, n := range nodes {
					node := n
					g.Go(func() error {
						pinger, err := ping.NewPinger(node)
						if err != nil {
							return nil
						}
						pinger.SetPrivileged(true)

						go func() {
							<-gctx.Done()
							pinger.Stop()
						}()

						pinger.Count = 5
						err = pinger.Run()
						if err != nil {
							return nil
						}

						stats.IncReportNumber()
						stats.AddMaxRtt(float64(pinger.Statistics().MaxRtt.Microseconds()))
						stats.AddMinRtt(float64(pinger.Statistics().MinRtt.Microseconds()))
						stats.AddMeanRtt(float64(pinger.Statistics().AvgRtt.Microseconds()))
						stats.AddMeanRtt(float64(pinger.Statistics().AvgRtt.Microseconds()))
						stats.AddPacketLoss(pinger.Statistics().PacketLoss)

						return nil
					})
				}

				err = g.Wait()
				if err != nil {
					return err
				}

				metrics.Gather(stats)

				ticker.Reset(30 * time.Second)

				continue
			}
		}
	}
}
