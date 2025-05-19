package opts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestConfig_NewDefaultOpts(t *testing.T) {
	cond := []struct {
		desc string
		in   Opt
		out  interface{}
	}{
		{desc: "", in: Verbose, out: false},
	}

	for _, tt := range cond {
		t.Run(tt.desc, func(t *testing.T) {
			o := NewDefaultOpts()

			v, err := o.Get(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.out, v)
		})
	}
}

func TestConfig_WithLogger(t *testing.T) {
	logger, err := zap.NewProduction()
	defer func() { _ = logger.Sync() }()
	assert.NoError(t, err)

	o := New(WithLogger(logger))
	v, err := o.Get(Logger)
	assert.NoError(t, err)
	assert.NotNil(t, v)
}
