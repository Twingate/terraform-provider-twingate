package datasource

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

func convertConnectorsToTerraform(connectors []*model.Connector) []interface{} {
	out := make([]interface{}, 0, len(connectors))

	for _, connector := range connectors {
		out = append(out, connector.ToTerraform())
	}

	return out
}

func convertGroupsToTerraform(groups []*model.Group) []interface{} {
	out := make([]interface{}, 0, len(groups))

	for _, group := range groups {
		out = append(out, group.ToTerraform())
	}

	return out
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
