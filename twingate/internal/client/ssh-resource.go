package client //nolint:dupl

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
)

func (client *Client) CreateSSHResource(ctx context.Context, sshResource *model.SSHResource) (*model.SSHResource, error) {
	opr := resourceSSHResource.create()

	variables := newVars(
		gqlVar(sshResource.Name, "name"),
		gqlVar(sshResource.Address, "address"),
		gqlID(sshResource.GatewayID, "gatewayId"),
		gqlID(sshResource.RemoteNetworkID, "remoteNetworkId"),
	)

	response := query.CreateSSHResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: sshResource.Name}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadSSHResource(ctx context.Context, resourceID string) (*model.SSHResource, error) {
	opr := resourceSSHResource.read()

	if resourceID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(gqlID(resourceID))
	response := query.ReadSSHResource{}

	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
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
	)

	response := query.UpdateSSHResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: sshResource.ID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteSSHResource(ctx context.Context, resourceID string) error {
	opr := resourceSSHResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}
