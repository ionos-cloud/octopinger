package octopinger

import (
	"context"
	"time"

	"github.com/go-ping/ping"
	"github.com/montanaflynn/stats"
	"golang.org/x/sync/errgroup"
)

type prober struct {
	opts *Opts
}

// NewProber ...
func NewProber(opt ...Opt) *prober {
	options := new(Opts)
	options.Configure(opt...)

	p := new(prober)
	p.opts = options

	return p
}

// Prober ...
type Prober interface {
	// Do ...
	Do(ctx context.Context, probe Probe) (*Stats, error)
}

// Do ...
func (p *prober) Do(ctx context.Context, probe Probe) func() error {
	return func() error {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
			case <-ticker.C:
				nodeList := NewNodeList()

				err := nodeList.Load(p.opts.configPath, "nodes")
				if err != nil {
					return err
				}

				g, gctx := errgroup.WithContext(ctx)
				g.SetLimit(10)

				samples := NewSamples()
				gather := make(chan *Stats)

				samples.Gather(ctx, gather)

				nodes := nodeList.Nodes()
				health := true
				for _, n := range nodes {
					node := n
					g.Go(func() error {
						stats, err := probe.Do(gctx, node)
						if err != nil {
							return err
						}

						gather <- stats

						return nil
					})
				}

				err = g.Wait()
				if err != nil {
					health = false
				}

				p.opts.monitor.SetProbeHealth(p.opts.nodeName, probe.Name(), health)
				p.opts.monitor.SetProbeRttMax(p.opts.nodeName, probe.Name(), samples.MaxRtt())
				p.opts.monitor.SetProbeRttMin(p.opts.nodeName, probe.Name(), samples.MinRtt())
				p.opts.monitor.SetProbeRttMean(p.opts.nodeName, probe.Name(), samples.MeanRtt())
				p.opts.monitor.SetProbeRttStddev(p.opts.nodeName, probe.Name(), samples.StdDevRtt())
				p.opts.monitor.SetProbePacketLossMax(p.opts.nodeName, probe.Name(), samples.PacketLossMax())
				p.opts.monitor.SetProbePacketLossMean(p.opts.nodeName, probe.Name(), samples.PacketLossMean())
				p.opts.monitor.SetProbePacketlossMin(p.opts.nodeName, probe.Name(), samples.PacketLossMin())
				p.opts.monitor.SetProbeNodesTotal(p.opts.nodeName, probe.Name(), float64(len(nodeList.Nodes())))
				p.opts.monitor.SetProbeNodesReports(p.opts.nodeName, probe.Name(), samples.Reports())

				continue
			}
		}
	}
}

// Probe ...
type Probe interface {
	// Do ...
	Do(ctx context.Context, addr string) (*Stats, error)

	// Name ...
	Name() string
}

type icmpProbe struct {
	opts *Opts
	name string
}

// Samples ...
type Samples struct {
	maxRtt     []float64
	minRtt     []float64
	meanRtt    []float64
	packetloss []float64
	num        int
}

func (s *Samples) addMaxRtt(rtt time.Duration) {
	s.maxRtt = append(s.maxRtt, float64(rtt.Microseconds()))
}

func (s *Samples) addMinRtt(rtt time.Duration) {
	s.minRtt = append(s.minRtt, float64(rtt.Microseconds()))
}

func (s *Samples) addMeanRtt(rtt time.Duration) {
	s.meanRtt = append(s.meanRtt, float64(rtt.Microseconds()))
}

func (s *Samples) addPacketLoss(percentage float64) {
	s.meanRtt = append(s.packetloss, percentage)
}

// MeanRtt ...
func (s *Samples) MeanRtt() float64 {
	m, err := stats.Mean(s.meanRtt)
	if err != nil {
		return 0
	}

	return m
}

// MaxRtt ...
func (s *Samples) MaxRtt() float64 {
	max, err := stats.Max(s.maxRtt)
	if err != nil {
		return 0
	}

	return max
}

// MinRtt ...
func (s *Samples) MinRtt() float64 {
	min, err := stats.Min(s.minRtt)
	if err != nil {
		return 0
	}

	return min
}

// StdDevRtt ...
func (s *Samples) StdDevRtt() float64 {
	stdDev, err := stats.StdDevS(s.meanRtt)
	if err != nil {
		return 0
	}

	return stdDev
}

// PacketLossMean ...
func (s *Samples) PacketLossMean() float64 {
	m, err := stats.Mean(s.packetloss)
	if err != nil {
		return 0
	}

	return m
}

// PacketLossMax ...
func (s *Samples) PacketLossMax() float64 {
	m, err := stats.Max(s.packetloss)
	if err != nil {
		return 0
	}

	return m
}

// PacketLossMin ...
func (s *Samples) PacketLossMin() float64 {
	m, err := stats.Min(s.packetloss)
	if err != nil {
		return 0
	}

	return m
}

// Reports ...
func (s *Samples) Reports() float64 {
	return float64(s.num)
}

// NewSamples ...
func NewSamples() *Samples {
	return &Samples{}
}

// Gather ...
func (s *Samples) Gather(ctx context.Context, ch <-chan *Stats) {
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
			case stat := <-ch:
				s.addMaxRtt(stat.MaxRtt)
				s.addMinRtt(stat.MinRtt)
				s.addMeanRtt(stat.AvgRtt)
				s.addPacketLoss(stat.PacketLoss)
				s.num += 1
			}
		}
	}(ctx)
}

// Stats ...
type Stats struct {
	MaxRtt     time.Duration
	MinRtt     time.Duration
	AvgRtt     time.Duration
	PacketLoss float64
}

// NewICMPProbe ...
func NewICMPProbe(opts ...Opt) *icmpProbe {
	options := new(Opts)
	options.Configure(opts...)

	p := new(icmpProbe)
	p.opts = options

	p.name = "icmp"

	return p
}

// Name ...
func (p *icmpProbe) Name() string {
	return p.name
}

// Do ...
func (i *icmpProbe) Do(ctx context.Context, addr string) (*Stats, error) {
	pinger, err := ping.NewPinger(addr)
	if err != nil {
		return nil, err
	}
	pinger.SetPrivileged(true)

	go func() {
		<-ctx.Done()
		pinger.Stop()
	}()

	pinger.Count = 5
	err = pinger.Run()
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		MaxRtt:     pinger.Statistics().MaxRtt,
		MinRtt:     pinger.Statistics().MinRtt,
		AvgRtt:     pinger.Statistics().AvgRtt,
		PacketLoss: pinger.Statistics().PacketLoss,
	}

	return stats, nil
}
