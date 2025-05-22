package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupName(t *testing.T) {
	assert.Equal(t, "", Service.Name())

	_ = os.Setenv("NAME", "test")

	env := ServiceEnv{"NAME"}

	Service.Lookup(env)
	assert.Equal(t, "test", Service.Name())

	_ = os.Setenv("NAME", "foo")
	Service.Lookup(env)
	assert.NotEqual(t, "foo", Service.Name())
}
