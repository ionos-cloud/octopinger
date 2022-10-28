package octopinger

import (
	"context"
	"time"

	srv "github.com/katallaxie/pkg/server"
	"go.uber.org/zap"
)

type server struct {
	opts *Opts
	srv.Listener
}

// Opts ...
type Opts struct {
	configPath string
	interval   time.Duration
	logger     *zap.Logger
	monitor    *Monitor
	nodeName   string
	podIP      string
	timeout    time.Duration
}

// Configure ...
func (o *Opts) Configure(opts ...Opt) {
	for _, opt := range opts {
		opt(o)
	}
}

// Opt ...
type Opt func(*Opts)

// WithMonitor ...
func WithMonitor(m *Monitor) Opt {
	return func(o *Opts) {
		o.monitor = m
	}
}

// WithLogger ...
func WithLogger(logger *zap.Logger) Opt {
	return func(o *Opts) {
		o.logger = logger
	}
}

// WithConfigPath ...
func WithConfigPath(path string) Opt {
	return func(o *Opts) {
		o.configPath = path
	}
}

// WithNodeName ...
func WithNodeName(nodeName string) Opt {
	return func(o *Opts) {
		o.nodeName = nodeName
	}
}

// WithTimeout ...
func WithTimeout(time time.Duration) Opt {
	return func(o *Opts) {
		o.timeout = time
	}
}

// WithInterval ...
func WithInterval(time time.Duration) Opt {
	return func(o *Opts) {
		o.interval = time
	}
}

// WithPodIP ...
func WithPodIP(ip string) Opt {
	return func(o *Opts) {
		o.podIP = ip
	}
}

// NewServer ...
func NewServer(opts ...Opt) *server {
	options := new(Opts)
	options.Configure(opts...)

	s := new(server)
	s.opts = options

	return s
}

// Start ...
func (s *server) Start(ctx context.Context, ready srv.ReadyFunc, run srv.RunFunc) func() error {
	return func() error {
		cfg, err := Config{}.Load(s.opts.configPath)
		if err != nil {
			return err
		}

		opts := []Opt{
			WithConfigPath(s.opts.configPath),
			WithNodeName(s.opts.nodeName),
			WithLogger(s.opts.logger),
			WithPodIP(s.opts.podIP),
		}

		if cfg.ICMP.Enabled {
			run(NewICMPProbe(s.opts.nodeName, opts...).Do(ctx, s.opts.monitor))
		}

		<-ctx.Done()

		return nil
	}
}
