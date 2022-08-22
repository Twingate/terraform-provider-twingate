package client

import (
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func TestProtocolsInput(t *testing.T) {
	t.Run("Test Twingate Resource : ProtocolsInput", func(t *testing.T) {
		pi := transport.NewEmptyProtocols()

		assert.EqualValues(t, model.PolicyAllowAll, pi.TCP.Policy)
		assert.EqualValues(t, model.PolicyAllowAll, pi.UDP.Policy)
		assert.NotNil(t, pi.UDP.Ports)
		assert.NotNil(t, pi.TCP.Ports)

		pi.AllowIcmp = graphql.Boolean(true)
		pri := &transport.PortRangeInput{Start: graphql.Int(1), End: graphql.Int(18000)}
		pi.TCP.Ports = append(pi.TCP.Ports, pri)
		pi.UDP.Ports = append(pi.UDP.Ports, pri)
		udpPorts, udpPolicy := pi.UDP.BuildPortsRange()
		tcpPorts, tcpPolicy := pi.TCP.BuildPortsRange()
		assert.EqualValues(t, model.PolicyAllowAll, udpPolicy)
		assert.EqualValues(t, model.PolicyAllowAll, tcpPolicy)
		assert.EqualValues(t, "1-18000", tcpPorts[0])
		assert.EqualValues(t, "1-18000", udpPorts[0])
	})
}

func TestConvertIDNameToString(t *testing.T) {
	t.Run("Test Twingate Resource : Convert ID Name To String", func(t *testing.T) {
		in := &transport.IDName{ID: graphql.ID("id"), Name: graphql.String("name")}
		id := in.StringID()
		name := in.StringName()
		assert.Equal(t, "name", name)
		assert.Equal(t, "id", id)
	})
}
