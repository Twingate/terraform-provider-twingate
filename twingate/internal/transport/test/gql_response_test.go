package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func TestProtocolsInput(t *testing.T) {
	t.Run("Test Twingate Resource : ProtocolsInput", func(t *testing.T) {
		pi := newEmptyProtocols()

		assert.EqualValues(t, policyAllowAll, pi.TCP.Policy)
		assert.EqualValues(t, policyAllowAll, pi.UDP.Policy)
		assert.NotNil(t, pi.UDP.Ports)
		assert.NotNil(t, pi.TCP.Ports)

		pi.AllowIcmp = graphql.Boolean(true)
		pri := &PortRangeInput{Start: graphql.Int(1), End: graphql.Int(18000)}
		pi.TCP.Ports = append(pi.TCP.Ports, pri)
		pi.UDP.Ports = append(pi.UDP.Ports, pri)
		udpPorts, udpPolicy := pi.UDP.buildPortsRange()
		tcpPorts, tcpPolicy := pi.TCP.buildPortsRange()
		assert.EqualValues(t, policyAllowAll, udpPolicy)
		assert.EqualValues(t, policyAllowAll, tcpPolicy)
		assert.EqualValues(t, "1-18000", tcpPorts[0])
		assert.EqualValues(t, "1-18000", udpPorts[0])
	})
}

func TestConvertIDNameToString(t *testing.T) {
	t.Run("Test Twingate Resource : Convert ID Name To String", func(t *testing.T) {
		in := &IDName{ID: graphql.ID("id"), Name: graphql.String("name")}
		id := in.StringID()
		name := in.StringName()
		assert.Equal(t, "name", name)
		assert.Equal(t, "id", id)
	})
}
