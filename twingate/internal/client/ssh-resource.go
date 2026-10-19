package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/utils"
)

// SSHDownstreamInput and SSHUpstreamInput must keep their names in sync with the
// GraphQL schema: the client derives the variable type from the Go type name.
type SSHDownstreamInput struct {
	Port int64 `json:"port"`
}

type SSHUpstreamInput struct {
	Port int64 `json:"port"`
}

func newSSHDownstreamInput(downstream *model.SSHDownstream) *SSHDownstreamInput {
	if downstream == nil {
		return nil
	}

	return &SSHDownstreamInput{Port: downstream.Port}
}

func newSSHUpstreamInput(upstream *model.SSHUpstream) *SSHUpstreamInput {
	if upstream == nil {
		return nil
	}

	return &SSHUpstreamInput{Port: upstream.Port}
}

func (client *Client) CreateSSHResource(ctx context.Context, sshResource *model.SSHResource) (*model.SSHResource, error) {
	opr := resourceSSHResource.create()

	variables := newVars(
		gqlVar(sshResource.Name, "name"),
		gqlVar(sshResource.Address, "address"),
		gqlID(sshResource.GatewayID, "gatewayId"),
		gqlID(sshResource.RemoteNetworkID, "remoteNetworkId"),
		gqlNullable(sshResource.IsVisible, "isVisible"),
		gqlNullable(sshResource.Alias, "alias"),
		gqlNullableID(sshResource.SecurityPolicyID, "securityPolicyId"),
		gqlVar(newTagInputs(sshResource.Tags), "tags"),
		gqlNullable(newSSHDownstreamInput(sshResource.Downstream), "downstream"),
		gqlNullable(newSSHUpstreamInput(sshResource.Upstream), "upstream"),
		gqlVar(NewAccessPolicyInput(sshResource.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(sshResource.AccessPolicy), "approvalMode"),
	)

	response := query.CreateSSHResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: sshResource.Name}); err != nil {
		return nil, err
	}

	res := response.ToModel()

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
		gqlNullable(sshResource.Alias, "alias"),
		gqlNullableID(sshResource.SecurityPolicyID, "securityPolicyId"),
		gqlVar(newTagInputs(sshResource.Tags), "tags"),
		gqlNullable(newSSHDownstreamInput(sshResource.Downstream), "downstream"),
		gqlNullable(newSSHUpstreamInput(sshResource.Upstream), "upstream"),
		gqlVar(NewAccessPolicyInput(sshResource.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(sshResource.AccessPolicy), "approvalMode"),
	)

	response := query.UpdateSSHResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: sshResource.ID}); err != nil {
		return nil, err
	}

	res := response.ToModel()
	res.GroupsAccess = sshResource.GroupsAccess

	return res, nil
}

func (client *Client) ReadSSHResources(ctx context.Context) ([]*model.SSHResource, error) {
	opr := resourceSSHResource.read().withCustomName("readSSHResources")

	variables := newVars(
		cursor(query.CursorResources),
		pageLimit(client.pageLimit),
	)

	response := query.ReadShallowResourcesWithType{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil && !errors.Is(err, ErrGraphqlResultIsEmpty) {
		return nil, err
	}

	if err := response.FetchPages(ctx, client.readShallowResourcesWithTypeAfter, variables); err != nil {
		return nil, err //nolint
	}

	return utils.FilterMap(response.Edges,
		func(edge *query.ShallowResourceEdge) bool {
			return edge.Node.Type == "SSHResource"
		},
		func(edge *query.ShallowResourceEdge) *model.SSHResource {
			return &model.SSHResource{
				ID:   string(edge.Node.ID),
				Name: edge.Node.Name,
			}
		}), nil
}

func (client *Client) readShallowResourcesWithTypeAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.ShallowResourceEdge], error) {
	opr := resourceSSHResource.read().withCustomName("readShallowResourcesWithTypeAfter")

	variables[query.CursorResources] = cursor

	response := query.ReadShallowResourcesWithType{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	//nolint:staticcheck
	return &response.ShallowResourcesWithType.PaginatedResource, nil
}

func (client *Client) DeleteSSHResource(ctx context.Context, resourceID string) error {
	opr := resourceSSHResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}
