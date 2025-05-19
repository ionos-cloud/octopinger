package opts

import (
	"fmt"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

const (
	// DefaultVerbose ...
	DefaultVerbose = false
	// DefaultTermSignal is the signal to term the agent.
	DefaultTermSignal = syscall.SIGTERM
	// DefaultReloadSignal is the default signal for reload.
	DefaultReloadSignal = syscall.SIGHUP
	// DefaultKillSignal is the default signal for termination.
	DefaultKillSignal = syscall.SIGINT
)

// ErrNotFound signals that this option is not set.
var ErrNotFound = fmt.Errorf("option not found")

// Opt ...
type Opt int

const (
	// Verbose ...
	Verbose Opt = iota
	// ReloadSignal ...
	ReloadSignal
	// TermSignal ...
	TermSignal
	// KillSignal ...
	KillSignal
	// Logger ...
	Logger
)

// Opts ...
type Opts interface {
	// Get ...
	Get(Opt) (interface{}, error)
	// Set ...
	Set(Opt, interface{})
	// Configure ...
	Configure(...OptFunc)
}

// DefaultOpts ...
type DefaultOpts interface {
	// Verbose ...
	Verbose() bool
	// ReloadSignal ...
	ReloadSignal() syscall.Signal
	// TermSignal ...
	TermSignal() syscall.Signal
	// KillSignal ...
	KillSignal() syscall.Signal

	Opts
}

// OptFunc is an option
type OptFunc func(Opts)

// Options is default options structure.
type Options struct {
	opts map[Opt]interface{}

	sync.RWMutex
}

// DefaultOptions are a collection of default options.
type DefaultOptions struct {
	Options
}

// New returns a new instance of the options.
func New(opts ...OptFunc) Opts {
	o := new(Options)
	o.Configure(opts...)

	return o
}

// NewDefaultOpts returns options with a default configuration.
func NewDefaultOpts(opts ...OptFunc) DefaultOpts {
	o := new(DefaultOptions)
	o.Configure(opts...)

	o.Set(Verbose, DefaultVerbose)
	o.Set(ReloadSignal, DefaultReloadSignal)
	o.Set(TermSignal, DefaultTermSignal)
	o.Set(KillSignal, DefaultKillSignal)

	return o
}

// Verbose ...
func (o *DefaultOptions) Verbose() bool {
	v, _ := o.Get(Verbose)

	return v.(bool)
}

// ReloadSignal ...
func (o *DefaultOptions) ReloadSignal() syscall.Signal {
	v, _ := o.Get(ReloadSignal)

	return v.(syscall.Signal)
}

// TermSignal ...
func (o *DefaultOptions) TermSignal() syscall.Signal {
	v, _ := o.Get(TermSignal)

	return v.(syscall.Signal)
}

// KillSignal ...
func (o *DefaultOptions) KillSignal() syscall.Signal {
	v, _ := o.Get(KillSignal)

	return v.(syscall.Signal)
}

// WithLogger is setting a logger for options.
func WithLogger(logger *zap.Logger) OptFunc {
	return func(opts Opts) {
		opts.Set(Logger, logger)
	}
}

// Get ...
func (o *Options) Get(opt Opt) (interface{}, error) {
	o.RLock()
	defer o.RUnlock()

	v, ok := o.opts[opt]
	if !ok {
		return nil, ErrNotFound
	}

	return v, nil
}

// Set ...
func (o *Options) Set(opt Opt, v interface{}) {
	o.Lock()
	defer o.Unlock()

	o.opts[opt] = v
}

// Configure os configuring the options.
func (o *Options) Configure(opts ...OptFunc) {
	if o.opts == nil {
		o.opts = make(map[Opt]interface{})
	}

	for _, opt := range opts {
		opt(o)
	}
}
