package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	o "github.com/ionos-cloud/octopinger/internal/opts"
)

// ErrUnimplemented is returned when a listener is not implemented.
var ErrUnimplemented = errors.New("unimplemented")

type token struct{}

// ReadyFunc is the function that is called by Listener to signal
// that it is ready and the next Listener can be called.
type ReadyFunc func()

// RunFunc is a function that is called to attach more routines
// to the server.
type RunFunc func(func() error)

// Error ...
type Error struct {
	Err error
}

// Error ...
func (s *Error) Error() string { return fmt.Sprintf("server: %s", s.Err) }

// Unwrap ...
func (s *Error) Unwrap() error { return s.Err }

// NewError ...
func NewError(err error) *Error {
	return &Error{Err: err}
}

// Server is the interface to be implemented
// to run the server.
//
//	s, ctx := WithContext(context.Background())
//	s.Listen(listener, false)
//
//	if err := s.Wait(); err != nil {
//		panic(err)
//	}
type Server interface {
	// Run is running a new go routine
	Listen(listener Listener, ready bool)

	// Waits for the server to fail,
	// or gracefully shutdown if context is canceled
	Wait() error

	// SetLimit ...
	SetLimit(n int)
}

// Unimplemented is the default implementation.
type Unimplemented struct{}

// Start ...
func (s *Unimplemented) Start(context.Context, ReadyFunc, RunFunc) func() error {
	return func() error {
		return ErrUnimplemented
	}
}

// Listener is the interface to a listener,
// so starting and shutdown of a listener,
// or any routine.
type Listener interface {
	// Start is being called on the listener
	Start(context.Context, ReadyFunc, RunFunc) func() error
}

type listeners map[Listener]bool

// server holds the instance info of the server
type server struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg      sync.WaitGroup
	errOnce sync.Once
	err     error

	sem chan token

	listeners map[Listener]bool

	ready chan bool
	sys   chan os.Signal

	opts o.Opts
}

// WithContext ...
func WithContext(ctx context.Context, opts ...o.OptFunc) (*server, context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	// new server
	s := newServer(ctx, opts...)
	s.cancel = cancel
	s.ctx = ctx

	return s, ctx
}

func newServer(ctx context.Context, opts ...o.OptFunc) *server {
	options := o.NewDefaultOpts(opts...)

	s := new(server)
	s.opts = options

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.ctx = ctx

	s.listeners = make(listeners)
	s.ready = make(chan bool, 1)
	s.sys = make(chan os.Signal, 1)

	return s
}

// Listen ...
func (s *server) Listen(listener Listener, ready bool) {
	s.listeners[listener] = ready
}

// Wait ...
//
//nolint:gocyclo
func (s *server) Wait() error {
	// create ticker for interrupt signals
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	term, err := s.opts.Get(o.TermSignal)
	if err != nil {
		return err
	}

	kill, err := s.opts.Get(o.KillSignal)
	if err != nil {
		return err
	}

	signal.Notify(s.sys, term.(syscall.Signal), kill.(syscall.Signal))
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

OUTTER:
	// start all listeners in order
	for l, ready := range s.listeners {
		fn := func() {
			r := ready

			var readyOnce sync.Once
			readyOnce.Do(func() {
				if r {
					s.ready <- true
				}
			})
		}

		goFn := func(f func() error) { _ = s.run(f) }

		// schedule to routines
		_ = s.run(l.Start(s.ctx, fn, goFn))

		// this blocks until ready is called
		if ready {
			select {
			case <-s.ready:
				continue
			case <-s.sys:
				s.cancel()
				break OUTTER
			case <-s.ctx.Done():
				break OUTTER
			}
		}
	}

	// this is the main loop
	for {
		select {
		case <-ticker.C:
		case <-s.sys:
			// if there is sys interrupt
			// cancel the context of the routines
			s.cancel()
		case <-s.ctx.Done():
			if err := s.ctx.Err(); err != nil {
				return NewError(s.err)
			}

			return nil
		}
	}
}

// SetLimit limits the number of active listeners in this server
func (s *server) SetLimit(n int) {
	if n < 0 {
		s.sem = nil
		return
	}

	if len(s.sem) != 0 {
		panic(fmt.Errorf("server: modify limit while %v listeners run", len(s.sem)))
	}

	s.sem = make(chan token, n)
}

func (s *server) run(f func() error) error {
	if s.sem != nil {
		select {
		case s.sem <- token{}:
		default:
			return fmt.Errorf("server: start more then %v listeners", len(s.sem))
		}
	}

	s.wg.Add(1)
	fn := func() {
		defer s.done()

		if err := f(); err != nil {
			s.errOnce.Do(func() {
				s.err = err
				if s.cancel != nil {
					s.cancel()
				}
			})
		}
	}

	go fn()

	return nil
}

func (s *server) done() {
	if s.sem != nil {
		<-s.sem
	}

	s.wg.Done()
}
