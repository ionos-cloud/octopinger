package octopinger

import (
	"context"
	"os"
	"strings"
	"time"

	srv "github.com/katallaxie/pkg/server"
	"go.uber.org/zap"
)

type server struct {
	nodeList string
	probes   []Probe

	timeout  time.Duration
	interval time.Duration

	logger *zap.Logger
	srv.Listener
}

// Opt ...
type Opt func(*server)

// WithNodeList ...
func WithNodeList(nodeList string) Opt {
	return func(s *server) {
		s.nodeList = nodeList
	}
}

// WithLogger ...
func WithLogger(logger *zap.Logger) Opt {
	return func(s *server) {
		s.logger = logger
	}
}

// WithTimeout ...
func WithTimeout(timeout time.Duration) Opt {
	return func(s *server) {
		s.timeout = timeout
	}
}

// WithInterval ...
func WithInterval(interval time.Duration) Opt {
	return func(s *server) {
		s.interval = interval
	}
}

// WithProbes ...
func WithProbes(probes ...Probe) Opt {
	return func(s *server) {
		s.probes = append(s.probes, probes...)
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
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
			case <-ticker.C:
				nodes, err := os.ReadFile(s.nodeList)
				if err != nil {
					return err
				}

				nn := strings.Split(string(nodes), ",")

				for _, node := range nn {
					for _, p := range s.probes {
						err := p.Do(ctx, node, s.timeout)
						if err != nil {
							s.logger.Sugar().Error("could not ping: %w", err)
						}

						s.logger.Sugar().Infof("successfully pinged: %s", node)
					}
				}

				ticker.Reset(s.interval)
			}
		}
	}
}
