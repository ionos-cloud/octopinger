package octopinger

import (
	"context"
	"time"

	"github.com/ionos-cloud/octopinger/api/v1alpha1"
	srv "github.com/ionos-cloud/octopinger/internal/server"
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
	hostIP     string
	timeout    time.Duration
	config     *v1alpha1.Config
}

// Configure ...
func (o *Opts) Configure(opts ...Opt) {
	for _, opt := range opts {
		opt(o)
	}
}

// Opt ...
type Opt func(*Opts)

// WithConfig ...
func WithConfig(c *v1alpha1.Config) Opt {
	return func(o *Opts) {
		o.config = c
	}
}

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

// WithHostIP ...
func WithHostIP(ip string) Opt {
	return func(o *Opts) {
		o.hostIP = ip
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
		cfg, err := Config().Load(s.opts.configPath)
		if err != nil {
			return err
		}

		if cfg.ICMP.Enable {
			icmp := NewICMPProbe(
				s.opts.nodeName,
				WithConfigPath(s.opts.configPath),
				WithNodeName(s.opts.nodeName),
				WithLogger(s.opts.logger),
				WithPodIP(s.opts.podIP),
				WithHostIP(s.opts.hostIP),
				WithConfig(cfg),
			)

			run(icmp.Do(ctx, s.opts.monitor))
		}

		if cfg.DNS.Enable && len(cfg.DNS.Names) > 0 {
			timeout := 3 * time.Second
			if cfg.DNS.Timeout != "" {
				timeout, err = time.ParseDuration(cfg.DNS.Timeout)
				if err != nil {
					return err
				}
			}

			dns := NewDNSProbe(
				s.opts.nodeName,
				cfg.DNS.Server,
				cfg.DNS.Names,
				WithConfigPath(s.opts.configPath),
				WithNodeName(s.opts.nodeName),
				WithLogger(s.opts.logger),
				WithPodIP(s.opts.podIP),
				WithHostIP(s.opts.hostIP),
				WithTimeout(timeout),
				WithConfig(cfg),
			)

			run(dns.Do(ctx, s.opts.monitor))
		}

		<-ctx.Done()

		return nil
	}
}
