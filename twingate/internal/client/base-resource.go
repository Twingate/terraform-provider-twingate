package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
)

// baseResourceVars are the mutation variables every resource kind takes.
func baseResourceVars(res model.BaseResource) []gqlVarOption {
	return []gqlVarOption{
		gqlVar(res.Name, "name"),
		gqlVar(res.Address, "address"),
		gqlID(res.RemoteNetworkID, "remoteNetworkId"),
		gqlNullable(res.IsVisible, "isVisible"),
		gqlNullable(res.Alias, "alias"),
		gqlNullableID(res.SecurityPolicyID, "securityPolicyId"),
		gqlVar(newTagInputs(res.Tags), "tags"),
		gqlVar(NewAccessPolicyInput(res.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(res.AccessPolicy), "approvalMode"),
	}
}

// pamResourceVars are the mutation variables every PAM resource takes.
func pamResourceVars(res model.PAMResource) []gqlVarOption {
	return append(baseResourceVars(res.BaseResource), gqlID(res.GatewayID, "gatewayId"))
}

// pamVars builds the variables of a PAM resource mutation: the shared ones plus
// whatever the specific resource kind adds.
func pamVars(res model.PAMResource, extra ...gqlVarOption) map[string]any {
	return newVars(append(pamResourceVars(res), extra...)...)
}

// addPAMResourceAccess assigns the planned access to a just created PAM resource
// and records it on the created model.
func (client *Client) addPAMResourceAccess(ctx context.Context, created *model.PAMResource, input model.PAMResource) error {
	if err := client.AddResourceAccess(ctx, created.ID, NewAccessInputs(input.ServiceAccounts, input.GroupsAccess)); err != nil {
		return err
	}

	copyAccess(&created.BaseResource, input.BaseResource)

	return nil
}

// copyAccess copies the group access and service accounts of src to dst.
// It records the planned access on the model a mutation returns: the mutation
// payload does not include the access connection, or not the planned one.
func copyAccess(dst *model.BaseResource, src model.BaseResource) {
	dst.GroupsAccess = src.GroupsAccess
	dst.ServiceAccounts = src.ServiceAccounts
}
