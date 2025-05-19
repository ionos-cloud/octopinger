package server

import (
	"context"
	"maps"
	"net/http"
	"net/http/pprof"
)

var _ Listener = (*debug)(nil)

// DefaultRoues are the default routes for the debug listener.
var DefaultRoues = map[string]http.Handler{
	"/debug/pprof/trace":   http.HandlerFunc(pprof.Trace),
	"/debug/pprof/":        http.HandlerFunc(pprof.Index),
	"/debug/pprof/cmdline": http.HandlerFunc(pprof.Cmdline),
	"/debug/pprof/profile": http.HandlerFunc(pprof.Profile),
	"/debug/pprof/symbol":  http.HandlerFunc(pprof.Symbol),
}

type debug struct {
	opts    *DebugOpts
	mux     *http.ServeMux
	handler *http.Server
}

// DebugOpts are the options for the debug listener.
type DebugOpts struct {
	// Addr is the address to listen on.
	Addr string
	// Routes configures the routes for the debug listener.
	Routes map[string]http.Handler
}

// Configure is a method that configures the debug options.
func (o *DebugOpts) Configure(opts ...DebugOpt) {
	for _, opt := range opts {
		opt(o)
	}
}

// DefaultOpts returns the default options for the debug listener.
func DefaultOpts() *DebugOpts {
	return &DebugOpts{
		Addr:   ":8443",
		Routes: map[string]http.Handler{},
	}
}

// DebugOpt is a function that configures the debug options.
type DebugOpt func(*DebugOpts)

// NewDebug is a function that creates a new debug listener.
func NewDebug(opts ...DebugOpt) *debug {
	options := DefaultOpts()
	options.Configure(opts...)

	d := new(debug)
	d.opts = options

	// create the mux
	d.mux = http.NewServeMux()

	configureMux(d)

	d.handler = new(http.Server)
	d.handler.Addr = d.opts.Addr
	d.handler.Handler = d.mux

	return d
}

// Start is a method that starts the debug listener.
func (d *debug) Start(ctx context.Context, ready ReadyFunc, run RunFunc) func() error {
	return func() error {
		// noop, call to be ready
		ready()

		if err := d.handler.ListenAndServe(); err != nil {
			return err
		}

		return nil
	}
}

// WithAddr is adding this status addr as an option.
func WithAddr(addr string) DebugOpt {
	return func(opts *DebugOpts) {
		opts.Addr = addr
	}
}

// WithPprof is adding this pprof routes as an option.
func WithPprof() DebugOpt {
	return func(opts *DebugOpts) {
		maps.Copy(opts.Routes, DefaultRoues)
	}
}

// WithPrometheusHandler is adding this prometheus http handler as an option.
func WithPrometheusHandler(handler http.Handler) DebugOpt {
	return func(opts *DebugOpts) {
		maps.Copy(opts.Routes, map[string]http.Handler{
			"/debug/metrics": handler,
		})
	}
}

func configureMux(d *debug) {
	for route, handler := range d.opts.Routes {
		d.mux.Handle(route, handler)
	}
}
