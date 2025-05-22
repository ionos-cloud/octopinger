package server

import (
	"os"
	"path"
	"sync"
)

// ServiceEnv is a list of environment variables to lookup the service name.
type ServiceEnv []string

// Service is used to configure the
type service struct {
	name string

	once sync.Once
}

// Service is used to configure the service.
var Service = &service{}

// Name returns the service name.
func (s *service) Name() string {
	return s.name
}

// Loopkup is used to lookup the service name.
func (s *service) Lookup(env ServiceEnv) string {
	s.once.Do(func() {
		for _, name := range env {
			v, ok := os.LookupEnv(name)
			if ok {
				s.name = v
				break
			}
		}

		if s.name == "" {
			s.name = path.Base(os.Args[0])
		}
	})

	return s.name
}
