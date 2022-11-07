package octopinger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMaxRtt(t *testing.T) {
	maxRtt := NewMaxRtt("icmp", "monalisa")

	assert.Equal(t, maxRtt.nodeName, "monalisa")
	assert.Equal(t, maxRtt.probeName, "icmp")
}

func TestNewMinRtt(t *testing.T) {
	maxRtt := NewMinRtt("icmp", "monalisa")

	assert.Equal(t, maxRtt.nodeName, "monalisa")
	assert.Equal(t, maxRtt.probeName, "icmp")
}

func TestNewMeanRtt(t *testing.T) {
	maxRtt := NewMeanRtt("icmp", "monalisa")

	assert.Equal(t, maxRtt.nodeName, "monalisa")
	assert.Equal(t, maxRtt.probeName, "icmp")
}
