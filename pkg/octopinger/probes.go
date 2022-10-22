package octopinger

import (
	"context"
	"time"

	"github.com/go-ping/ping"
)

// Probe ...
type Probe interface {
	// Do ...
	Do(context.Context, string, time.Duration) error
}

type icmpProbe struct{}

// NewICMPProbe ...
func NewICMPProbe() *icmpProbe {
	i := new(icmpProbe)

	return i
}

// Do ...
func (i *icmpProbe) Do(ctx context.Context, addr string, timeout time.Duration) error {
	pinger, err := ping.NewPinger(addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		pinger.Stop()
	}()

	pinger.Count = 3
	err = pinger.Run()
	if err != nil {
		return err
	}
	// stats := pinger.Statistics()

	return nil
}
