package datasource

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func convertConnectorsToTerraform(connectors []*model.Connector) []connectorModel {
	return utils.Map(connectors, func(connector *model.Connector) connectorModel {
		return connectorModel{
			Name:                 types.StringValue(connector.Name),
			RemoteNetworkID:      types.StringValue(connector.NetworkID),
			StatusUpdatesEnabled: types.BoolPointerValue(connector.StatusUpdatesEnabled),
		}
	})
}

func convertGroupsToTerraform(groups []*model.Group) []groupModel {
	return utils.Map(groups, func(group *model.Group) groupModel {
		return groupModel{
			ID:               types.StringValue(group.ID),
			Name:             types.StringValue(group.Name),
			Type:             types.StringValue(group.Type),
			IsActive:         types.BoolValue(group.IsActive),
			SecurityPolicyID: types.StringValue(group.SecurityPolicyID),
		}
	})
}

func convertResourcesToTerraform(resources []*model.Resource) []interface{} {
	out := make([]interface{}, 0, len(resources))

	for _, res := range resources {
		out = append(out, res.ToTerraform())
	}

	return out
}

func convertUsersToTerraform(users []*model.User) []interface{} {
	out := make([]interface{}, 0, len(users))
	for _, user := range users {
		out = append(out, user.ToTerraform())
	}

	return out
}

func convertServicesToTerraform(services []*model.ServiceAccount) []interface{} {
	out := make([]interface{}, 0, len(services))

	for _, service := range services {
		out = append(out, service.ToTerraform())
	}

	return out
}

func convertSecurityPoliciesToTerraform(securityPolicies []*model.SecurityPolicy) []interface{} {
	out := make([]interface{}, 0, len(securityPolicies))
	for _, policy := range securityPolicies {
		out = append(out, policy.ToTerraform())
	}

	return out
}

func convertRemoteNetworksToTerraform(networks []*model.RemoteNetwork) []remoteNetworkModel {
	return utils.Map(networks, func(network *model.RemoteNetwork) remoteNetworkModel {
		return remoteNetworkModel{
			ID:       types.StringValue(network.ID),
			Name:     types.StringValue(network.Name),
			Location: types.StringValue(network.Location),
		}
	})
}
