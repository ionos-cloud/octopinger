package octopinger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	assert.NotNil(t, m.probeDNSError)
	assert.NotNil(t, m.probeDNSSuccess)
	assert.NotNil(t, m.probeNodesReports)
	assert.NotNil(t, m.probeNodesTotal)
	assert.NotNil(t, m.probePacketLossMax)
	assert.NotNil(t, m.probePacketLossMean)
	assert.NotNil(t, m.probePacketLossMin)
	assert.NotNil(t, m.probePacketLossTotal)
	assert.NotNil(t, m.probeRttMax)
	assert.NotNil(t, m.probeRttMean)
	assert.NotNil(t, m.probeRttMin)
}
