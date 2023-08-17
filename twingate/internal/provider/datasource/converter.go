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

func convertUsersToTerraform(users []*model.User) []userModel {
	return utils.Map(users, func(user *model.User) userModel {
		return userModel{
			ID:        types.StringValue(user.ID),
			FirstName: types.StringValue(user.FirstName),
			LastName:  types.StringValue(user.LastName),
			Email:     types.StringValue(user.Email),
			IsAdmin:   types.BoolValue(user.IsAdmin()),
			Role:      types.StringValue(user.Role),
			Type:      types.StringValue(user.Type),
		}
	})
}

func convertServicesToTerraform(accounts []*model.ServiceAccount) []serviceAccountModel {
	return utils.Map(accounts, func(account *model.ServiceAccount) serviceAccountModel {
		return serviceAccountModel{
			ID:          types.StringValue(account.ID),
			Name:        types.StringValue(account.Name),
			ResourceIDs: utils.Map(account.Resources, types.StringValue),
			KeyIDs:      utils.Map(account.Keys, types.StringValue),
		}
	})
}

func convertSecurityPoliciesToTerraform(policies []*model.SecurityPolicy) []securityPolicyModel {
	return utils.Map(policies, func(policy *model.SecurityPolicy) securityPolicyModel {
		return securityPolicyModel{
			ID:   types.StringValue(policy.ID),
			Name: types.StringValue(policy.Name),
		}
	})
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
