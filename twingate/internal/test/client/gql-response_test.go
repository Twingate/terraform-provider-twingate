package client

import (
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func TestProtocols(t *testing.T) {
	t.Run("Test Twingate Resource : Protocols", func(t *testing.T) {
		protocols := model.DefaultProtocols()

		assert.EqualValues(t, model.PolicyAllowAll, protocols.TCP.Policy)
		assert.EqualValues(t, model.PolicyAllowAll, protocols.UDP.Policy)
		assert.Nil(t, protocols.UDP.Ports)
		assert.Nil(t, protocols.TCP.Ports)

		port := &model.PortRange{Start: 1, End: 18000}
		protocols.TCP.Ports = append(protocols.TCP.Ports, port)
		protocols.UDP.Ports = append(protocols.UDP.Ports, port)
		udpPorts := protocols.UDP.PortsToString()
		tcpPorts := protocols.TCP.PortsToString()
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
