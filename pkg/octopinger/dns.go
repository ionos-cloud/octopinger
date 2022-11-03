package octopinger

import (
	"context"
	"net"
	"sync"
	"time"
)

type token struct{}

type dnsProbe struct {
	maxConcurrency int

	dnsFailed  *dnsFailed
	dnsTotal   *dnsTotal
	dnsError   *dnsError
	dnsSuccess *dnsSuccess

	names []string

	nodeName string

	sem chan token
	wg  sync.WaitGroup
	mux sync.Mutex
}

// NewDNSProbe ...
func NewDNSProbe(nodeName string, names ...string) *dnsProbe {
	d := new(dnsProbe)
	d.nodeName = nodeName
	d.names = names
	d.maxConcurrency = 100
	d.sem = make(chan token, d.maxConcurrency)

	d.Reset()

	return d
}

// Reset ...
func (d *dnsProbe) Reset() {
	d.dnsFailed = NewDNSFailed(d.nodeName)
	d.dnsError = NewDNSError(d.nodeName)
	d.dnsSuccess = NewDNSSuccess(d.nodeName)
	d.dnsTotal = NewDNSTotal(d.nodeName)
}

// Collect ...
func (d *dnsProbe) Collect(ch chan<- Metric) {
	d.dnsError.Collect(ch)
	d.dnsFailed.Collect(ch)
	d.dnsSuccess.Collect(ch)
	d.dnsTotal.Collect(ch)
}

type dnsFailed struct {
	value    float64
	nodeName string

	Metric
	Collector
}

// Write ...
func (d *dnsFailed) Write(monitor *Monitor) error {
	monitor.SetProbeDNSFailure(d.nodeName, d.value)

	return nil
}

// Collect ...
func (d *dnsFailed) Collect(ch chan<- Metric) {
	ch <- d
}

// NewDNSFailed ...
func NewDNSFailed(nodeName string) *dnsFailed {
	return &dnsFailed{
		nodeName: nodeName,
	}
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

type dnsTotal struct {
	value    float64
	nodeName string

	Metric
	Collector
}

// Write ...
func (d *dnsTotal) Write(monitor *Monitor) error {
	monitor.SetProbeDNSTotal(d.nodeName, d.value)

	return nil
}

// Collect ...
func (d *dnsTotal) Collect(ch chan<- Metric) {
	ch <- d
}

// NewDNSTotal ...
func NewDNSTotal(nodeName string) *dnsTotal {
	return &dnsTotal{
		nodeName: nodeName,
	}
}

// SetTotal ...
func (d *dnsProbe) SetTotal(value float64) {
	d.mux.Lock()
	defer d.mux.Unlock()

	d.dnsTotal.value = value
}

// IncFailure ...
func (d *dnsProbe) IncFailure() {
	d.mux.Lock()
	defer d.mux.Unlock()

	d.dnsFailed.value += 1
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
				d.resolve(ctx, d.names...)

				metrics.Gather(d)
				ticker.Reset(1 * time.Second)

				continue
			}
		}
	}
}

func (d *dnsProbe) resolve(ctx context.Context, names ...string) {
	resolver := net.Resolver{}

	d.Reset()
	d.SetTotal(float64(len(names)))

	for _, name := range names {
		name := name

		d.wg.Add(1)
		go func() {
			defer d.wg.Done()

			d.sem <- token{}

			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()

			ips, err := resolver.LookupHost(ctx, name)
			if err != nil {
				d.IncError()

				return
			}

			if len(ips) == 0 {
				d.IncFailure()

				return
			}

			d.IncSuccess()

			<-d.sem
		}()
	}
}
