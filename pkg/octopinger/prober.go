package octopinger

import (
	"context"
	"time"

	"github.com/go-ping/ping"
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
	Do(ctx context.Context, probe Probe) error
}

// Do ...
func (p *prober) Do(ctx context.Context, probe Probe) func() error {
	return func() error {
		ticker := time.NewTicker(time.Second)
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

				healthy := true
				for _, n := range nodeList.Nodes() {
					node := n
					g.Go(func() error {
						err := probe.Do(gctx, node)
						if err != nil {
							return err
						}

						return nil
					})
				}

				err = g.Wait()
				if err != nil {
					healthy = false
				}

				p.opts.monitor.SetProbeHealth(p.opts.nodeName, probe.Name(), healthy)

				continue
			}
		}
	}
}

// Probe ...
type Probe interface {
	// Do ...
	Do(ctx context.Context, addr string) error

	// Name ...
	Name() string
}

type icmpProbe struct {
	opts *Opts
	name string
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
func (i *icmpProbe) Do(ctx context.Context, addr string) error {
	pinger, err := ping.NewPinger(addr)
	if err != nil {
		return err
	}
	pinger.SetPrivileged(true)

	go func() {
		<-ctx.Done()
		pinger.Stop()
	}()

	pinger.Count = 3
	err = pinger.Run()
	if err != nil {
		return err
	}

	return nil
}
