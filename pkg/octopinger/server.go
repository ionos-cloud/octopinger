package octopinger

import (
	"context"

	srv "github.com/katallaxie/pkg/server"
)

type server struct {
	srv.Listener
}

// NewServer ...
func NewServer() *server {
	s := new(server)

	return s
}

// Start ...
func (s *server) Start(ctx context.Context, ready srv.ReadyFunc, run srv.RunFunc) func() error {
	return func() error {
		return nil
	}
}
