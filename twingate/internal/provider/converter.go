package provider

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"

func ConvertProtocolsToTerraform(protocols *model.Protocols) []interface{} {
	if protocols == nil {
		return nil
	}

	rawMap := make(map[string]interface{})
	rawMap["allow_icmp"] = protocols.AllowIcmp

	if protocols.TCP != nil {
		rawMap["tcp"] = convertPortToTerraform(protocols.TCP)
	}

	if protocols.UDP != nil {
		rawMap["udp"] = convertPortToTerraform(protocols.UDP)
	}

	return []interface{}{rawMap}
}

func convertPortToTerraform(protocol *model.Protocol) []interface{} {
	if protocol == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"policy": protocol.Policy,
			"ports":  protocol.PortsToString(),
		},
	}
}
