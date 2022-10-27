package octopinger

import (
	"context"
	"sync"
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

				healthy := true
				for _, n := range nodeList.Nodes() {
					node := n
					g.Go(func() error {
						stats, err := probe.Do(gctx, node)
						if err != nil {
							return err
						}

						samples.AddMeanRtt(stats.AvgRtt)
						samples.AddMaxRtt(stats.MaxRtt)
						samples.AddMinRtt(stats.MinRtt)
						samples.AddPacketLoss(stats.PacketLoss)

						return nil
					})
				}

				err = g.Wait()
				if err != nil {
					healthy = false
				}

				p.opts.monitor.SetProbeHealth(p.opts.nodeName, probe.Name(), healthy)
				p.opts.monitor.SetProbeRttMax(p.opts.nodeName, probe.Name(), samples.MaxRtt())
				p.opts.monitor.SetProbeRttMin(p.opts.nodeName, probe.Name(), samples.MinRtt())
				p.opts.monitor.SetProbeRttMean(p.opts.nodeName, probe.Name(), samples.MeanRtt())
				p.opts.monitor.SetProbeRttStddev(p.opts.nodeName, probe.Name(), samples.StdDevRtt())
				p.opts.monitor.SetProbePacketLossMax(p.opts.nodeName, probe.Name(), samples.PacketLossMax())
				p.opts.monitor.SetProbePacketLossMean(p.opts.nodeName, probe.Name(), samples.PacketLossMean())
				p.opts.monitor.SetProbePacketlossMin(p.opts.nodeName, probe.Name(), samples.PacketLossMin())

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

	sync.Mutex
}

// AddMaxRtt ...
func (s *Samples) AddMaxRtt(rtt time.Duration) {
	s.Lock()
	defer s.Unlock()
	s.maxRtt = append(s.maxRtt, float64(rtt.Milliseconds()))
}

// AddMinxRtt ...
func (s *Samples) AddMinRtt(rtt time.Duration) {
	s.Lock()
	defer s.Unlock()
	s.minRtt = append(s.minRtt, float64(rtt.Milliseconds()))
}

// AddMeanRtt ...
func (s *Samples) AddMeanRtt(rtt time.Duration) {
	s.Lock()
	defer s.Unlock()
	s.meanRtt = append(s.meanRtt, float64(rtt.Milliseconds()))
}

// AddPacketLoss ...
func (s *Samples) AddPacketLoss(percentage float64) {
	s.Lock()
	defer s.Unlock()
	s.meanRtt = append(s.packetloss, percentage)
}

// MeanRtt ...
func (s *Samples) MeanRtt() float64 {
	s.Lock()
	defer s.Unlock()

	m, err := stats.Mean(s.meanRtt)
	if err != nil {
		return 0
	}

	return m
}

// MaxRtt ...
func (s *Samples) MaxRtt() float64 {
	s.Lock()
	defer s.Unlock()

	max, err := stats.Max(s.maxRtt)
	if err != nil {
		return 0
	}

	return max
}

// MinRtt ...
func (s *Samples) MinRtt() float64 {
	s.Lock()
	defer s.Unlock()

	min, err := stats.Min(s.minRtt)
	if err != nil {
		return 0
	}

	return min
}

// StdDevRtt ...
func (s *Samples) StdDevRtt() float64 {
	s.Lock()
	defer s.Unlock()

	stdDev, err := stats.StdDevS(s.meanRtt)
	if err != nil {
		return 0
	}

	return stdDev
}

// PacketLossMean ...
func (s *Samples) PacketLossMean() float64 {
	s.Lock()
	defer s.Unlock()

	m, err := stats.Mean(s.packetloss)
	if err != nil {
		return 0
	}

	return m
}

// PacketLossMax ...
func (s *Samples) PacketLossMax() float64 {
	s.Lock()
	defer s.Unlock()

	m, err := stats.Max(s.packetloss)
	if err != nil {
		return 0
	}

	return m
}

// PacketLossMin ...
func (s *Samples) PacketLossMin() float64 {
	s.Lock()
	defer s.Unlock()

	m, err := stats.Min(s.packetloss)
	if err != nil {
		return 0
	}

	return m
}

// NewSamples ...
func NewSamples() *Samples {
	return &Samples{}
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
