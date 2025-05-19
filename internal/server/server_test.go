package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithContext(t *testing.T) {
	srv, ctx := WithContext(context.Background())
	assert.Implements(t, (*Server)(nil), srv)
	assert.NotNil(t, ctx)
	assert.NotNil(t, srv)
	assert.Nil(t, srv.sem)
}

func TestSetLimit(t *testing.T) {
	srv, ctx := WithContext(context.Background())
	assert.Implements(t, (*Server)(nil), srv)
	assert.NotNil(t, srv)
	assert.NotNil(t, ctx)

	srv.SetLimit(1)
	assert.NotNil(t, srv.sem)
}

func TestSetLimitZero(t *testing.T) {
	srv, ctx := WithContext(context.Background())
	assert.Implements(t, (*Server)(nil), srv)
	assert.NotNil(t, srv)
	assert.NotNil(t, ctx)

	srv.SetLimit(0)
	assert.NotNil(t, srv.sem)
}

func TestSetLimitNegative(t *testing.T) {
	srv, ctx := WithContext(context.Background())
	assert.Implements(t, (*Server)(nil), srv)
	assert.NotNil(t, srv)
	assert.NotNil(t, ctx)

	srv.SetLimit(-1)
	assert.Nil(t, srv.sem)
}

func TestUnimplemented(t *testing.T) {
	srv, ctx := WithContext(context.Background())
	assert.Implements(t, (*Server)(nil), srv)
	assert.NotNil(t, srv)
	assert.NotNil(t, ctx)

	l := &Unimplemented{}
	assert.Implements(t, (*Listener)(nil), l)

	srv.Listen(l, false)
	err := srv.Wait()
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnimplemented)
}

func TestNewError(t *testing.T) {
	err := NewError(ErrUnimplemented)
	assert.Implements(t, (*error)(nil), err)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnimplemented)
}
