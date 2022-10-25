package octopinger

import (
	"context"
	"time"

	srv "github.com/katallaxie/pkg/server"
	"go.uber.org/zap"
)

type server struct {
	configPath string
	nodeName   string

	monitor *Monitor
	logger  *zap.Logger
	srv.Listener
}

// Opt ...
type Opt func(*server)

// WithLogger ...
func WithLogger(logger *zap.Logger) Opt {
	return func(s *server) {
		s.logger = logger
	}
}

// WithConfigPath ...
func WithConfigPath(path string) Opt {
	return func(s *server) {
		s.configPath = path
	}
}

// WithMonitor ...
func WithMonitor(m *Monitor) Opt {
	return func(s *server) {
		s.monitor = m
	}
}

// WithNodeName ...
func WithNodeName(nodeName string) Opt {
	return func(s *server) {
		s.nodeName = nodeName
	}
}

// NewServer ...
func NewServer(opts ...Opt) *server {
	s := new(server)

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start ...
func (s *server) Start(ctx context.Context, ready srv.ReadyFunc, run srv.RunFunc) func() error {
	return func() error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
			case <-ticker.C:
				cfg, err := Config{}.Load(s.configPath)
				if err != nil {
					return err
				}

				for _, n := range cfg.Nodes {
					s.logger.Sugar().Info("processing node: %s", n)

					if cfg.ICMP.Enabled {
						icmp := NewICMPProbe()
						err := icmp.Do(ctx, n, cfg.ICMP.Timeout)
						if err != nil {
							s.logger.Sugar().Error("could not ping: %w", err)
						}

						s.logger.Sugar().Infof("successfully pinged: %s", n)
					}
				}

				s.monitor.SetClusterHealthy(s.nodeName, true)

				ticker.Reset(cfg.ICMP.Interval)
			}
		}
	}
}
