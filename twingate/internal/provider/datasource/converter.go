package datasource

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider"
)

func convertConnectorsToTerraform(connectors []*model.Connector) []interface{} {
	out := make([]interface{}, 0, len(connectors))

	for _, connector := range connectors {
		out = append(out, convertConnectorToTerraform(connector))
	}

	return out
}

func convertConnectorToTerraform(connector *model.Connector) map[string]interface{} {
	return map[string]interface{}{
		"id":                connector.ID,
		"name":              connector.Name,
		"remote_network_id": connector.NetworkID,
	}
}

func convertGroupsToTerraform(groups []*model.Group) []interface{} {
	out := make([]interface{}, 0, len(groups))

	for _, group := range groups {
		out = append(out, map[string]interface{}{
			"id":        group.ID,
			"name":      group.Name,
			"type":      group.Type,
			"is_active": group.IsActive,
		})
	}

	return out
}

func convertResourcesToTerraform(resources []*model.Resource) []interface{} {
	out := make([]interface{}, 0, len(resources))

	for _, res := range resources {
		rawData := convertResourceToTerraform(res)
		if rawData == nil {
			continue
		}

		out = append(out, rawData)
	}

	return out
}

func convertResourceToTerraform(resource *model.Resource) interface{} {
	if resource == nil {
		return nil
	}

	return map[string]interface{}{
		"id":                resource.ID,
		"name":              resource.Name,
		"address":           resource.Address,
		"remote_network_id": resource.RemoteNetworkID,
		"protocols":         provider.ConvertProtocolsToTerraform(resource.Protocols),
	}
}

func convertUsersToTerraform(users []*model.User) []interface{} {
	out := make([]interface{}, 0, len(users))
	for _, user := range users {
		out = append(out, map[string]interface{}{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"is_admin":   user.IsAdmin(),
			"role":       user.Role,
		})
	}

	return out
}
