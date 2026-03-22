package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
)

//nolint:dupl
func (client *Client) CreateSSHResource(ctx context.Context, sshResource *model.SSHResource) (*model.SSHResource, error) {
	opr := resourceSSHResource.create()

	variables := newVars(
		gqlVar(sshResource.Name, "name"),
		gqlVar(sshResource.Address, "address"),
		gqlID(sshResource.GatewayID, "gatewayId"),
		gqlID(sshResource.RemoteNetworkID, "remoteNetworkId"),
		gqlNullable(sshResource.IsVisible, "isVisible"),
		gqlNullableEmpty(sshResource.Alias, "alias"),
		gqlNullableID(sshResource.SecurityPolicyID, "securityPolicyId"),
		gqlVar(newTagInputs(sshResource.Tags), "tags"),
		gqlVar(newProtocolsInput(sshResource.Protocols), "protocols"),
		gqlVar(NewAccessPolicyInput(sshResource.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(sshResource.AccessPolicy), "approvalMode"),
	)

	response := query.CreateSSHResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: sshResource.Name}); err != nil {
		return nil, err
	}

	res := response.ToModel()
	if res == nil {
		return nil, nil //nolint:nilnil
	}

	if len(sshResource.GroupsAccess) > 0 {
		if err := client.AddResourceAccess(ctx, res.ID, convertGroupsToAccessInput(sshResource.GroupsAccess)); err != nil {
			return nil, err
		}
	}

	res.GroupsAccess = sshResource.GroupsAccess

	return res, nil
}

func (client *Client) ReadSSHResource(ctx context.Context, resourceID string) (*model.SSHResource, error) {
	opr := resourceSSHResource.read()

	if resourceID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(resourceID),
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)
	response := query.ReadSSHResource{}

	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	if err := response.Resource.Access.FetchPages(withOperationCtx(ctx, opr), client.readSSHResourceAccessAfter, newVars(gqlID(resourceID))); err != nil {
		return nil, err //nolint
	}

	return response.ToModel() //nolint:wrapcheck
}

func (client *Client) readSSHResourceAccessAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
	opr := resourceSSHResource.read().withCustomName("readSSHResourceAccessAfter")

	variables[query.CursorAccess] = cursor
	pageLimit(client.pageLimit)(variables)

	response := query.ReadSSHResource{}
	if err := client.query(ctx, &response, variables, opr, attr{}); err != nil {
		return nil, err
	}

	return &response.Resource.Access.PaginatedResource, nil
}

func (client *Client) UpdateSSHResource(ctx context.Context, sshResource *model.SSHResource) (*model.SSHResource, error) {
	opr := resourceSSHResource.update()

	if sshResource.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(sshResource.ID),
		gqlVar(sshResource.Name, "name"),
		gqlVar(sshResource.Address, "address"),
		gqlID(sshResource.GatewayID, "gatewayId"),
		gqlID(sshResource.RemoteNetworkID, "remoteNetworkId"),
		gqlNullable(sshResource.IsVisible, "isVisible"),
		gqlNullableEmpty(sshResource.Alias, "alias"),
		gqlNullableID(sshResource.SecurityPolicyID, "securityPolicyId"),
		gqlVar(newTagInputs(sshResource.Tags), "tags"),
		gqlVar(newProtocolsInput(sshResource.Protocols), "protocols"),
		gqlVar(NewAccessPolicyInput(sshResource.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(sshResource.AccessPolicy), "approvalMode"),
	)

	response := query.UpdateSSHResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: sshResource.ID}); err != nil {
		return nil, err
	}

	res := response.ToModel()
	if res == nil {
		return nil, nil //nolint:nilnil
	}

	res.GroupsAccess = sshResource.GroupsAccess

	return res, nil
}

func (client *Client) DeleteSSHResource(ctx context.Context, resourceID string) error {
	opr := resourceSSHResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}
