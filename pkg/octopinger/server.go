package octopinger

import (
	"context"
	"os"
	"strings"

	srv "github.com/katallaxie/pkg/server"
	"go.uber.org/zap"
)

type server struct {
	nodeList string
	logger   *zap.Logger
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
		nodes, err := os.ReadFile(s.nodeList)
		if err != nil {
			return err
		}

		nn := strings.Split(string(nodes), ",")

		for _, node := range nn {
			s.logger.Sugar().Infof("configuring %s", node)
		}

		<-ctx.Done()

		return nil
	}
}
