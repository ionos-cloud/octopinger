package octopinger

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type token struct{}

type dnsProbe struct {
	opts *Opts

	dnsError   *dnsError
	dnsSuccess *dnsSuccess

	nodeName string
	server   string
	names    []string

	maxConcurrency int
	resolver       *net.Resolver

	sem chan token
	wg  sync.WaitGroup
	mux sync.Mutex
}

var (
	// ErrResolveHost ...
	ErrResolveHost = errors.New("could not resolve host")
)

// NewDNSProbe ...
func NewDNSProbe(nodeName, server string, names []string, opts ...Opt) *dnsProbe {
	options := new(Opts)
	options.Configure(opts...)

	d := new(dnsProbe)
	d.opts = options
	d.nodeName = nodeName
	d.server = server
	d.names = names
	d.maxConcurrency = 100
	d.sem = make(chan token, d.maxConcurrency)

	d.configureResolver()
	d.Reset()

	return d
}

// Reset ...
func (d *dnsProbe) Reset() {
	d.dnsError = NewDNSError(d.nodeName)
	d.dnsSuccess = NewDNSSuccess(d.nodeName)
}

// Collect ...
func (d *dnsProbe) Collect(ch chan<- Metric) {
	d.dnsError.Collect(ch)
	d.dnsSuccess.Collect(ch)
}

type dnsSuccess struct {
	value    float64
	nodeName string

	Metric
	Collector
}

// Write ...
func (d *dnsSuccess) Write(monitor *Monitor) error {
	monitor.SetProbeDNSSuccess(d.nodeName, d.value)

	return nil
}

// Collect ...
func (d *dnsSuccess) Collect(ch chan<- Metric) {
	ch <- d
}

// NewDNSSuccess ...
func NewDNSSuccess(nodeName string) *dnsSuccess {
	return &dnsSuccess{
		nodeName: nodeName,
	}
}

type dnsError struct {
	value    float64
	nodeName string

	Metric
	Collector
}

// Write ...
func (d *dnsError) Write(monitor *Monitor) error {
	monitor.SetProbeDNSError(d.nodeName, d.value)

	return nil
}

// Collect ...
func (d *dnsError) Collect(ch chan<- Metric) {
	ch <- d
}

// NewDNSError ...
func NewDNSError(nodeName string) *dnsError {
	return &dnsError{
		nodeName: nodeName,
	}
}

// IncSuccess ...
func (d *dnsProbe) IncSuccess() {
	d.mux.Lock()
	defer d.mux.Unlock()

	d.dnsSuccess.value += 1
}

// IncError ...
func (d *dnsProbe) IncError() {
	d.mux.Lock()
	defer d.mux.Unlock()

	d.dnsError.value += 1
}

// Do ...
func (d *dnsProbe) Do(ctx context.Context, metrics Gatherer) func() error {
	return func() error {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
			case <-ticker.C:
				d.do(ctx, d.names...)

				metrics.Gather(d)
				ticker.Reset(1 * time.Second)

				continue
			}
		}
	}
}

func (d *dnsProbe) do(ctx context.Context, hosts ...string) {
	d.Reset()

	for _, host := range hosts {
		host := host

		d.wg.Add(1)
		go func() {
			defer d.wg.Done()

			d.sem <- token{}

			err := d.resolve(ctx, host)
			if err != nil {
				d.IncError()
			} else {
				d.IncSuccess()
			}

			<-d.sem
		}()
	}

	d.wg.Wait()
}

func (d *dnsProbe) resolve(ctx context.Context, host string) error {
	ctx, cancel := context.WithTimeout(ctx, d.opts.timeout)
	defer cancel()

	ips, err := d.resolver.LookupHost(ctx, host)
	if len(ips) == 0 {
		return ErrResolveHost
	}

	return err
}

func (dp *dnsProbe) configureResolver() {
	r := &net.Resolver{
		PreferGo: true,
	}

	if dp.server != "" {
		r.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: dp.opts.timeout,
			}

			return d.DialContext(ctx, network, fmt.Sprintf("%s:53", dp.server))
		}
	}

	dp.resolver = r
}
