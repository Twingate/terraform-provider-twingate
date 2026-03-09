package client //nolint:dupl

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
)

func (client *Client) CreateKubernetesResource(ctx context.Context, k8sResource *model.KubernetesResource) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.create()

	variables := newVars(
		gqlVar(k8sResource.Name, "name"),
		gqlVar(k8sResource.Address, "address"),
		gqlID(k8sResource.GatewayID, "gatewayId"),
		gqlID(k8sResource.RemoteNetworkID, "remoteNetworkId"),
	)

	response := query.CreateKubernetesResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: k8sResource.Name}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadKubernetesResource(ctx context.Context, resourceID string) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.read()

	if resourceID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(gqlID(resourceID))
	response := query.ReadKubernetesResource{}

	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) UpdateKubernetesResource(ctx context.Context, k8sResource *model.KubernetesResource) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.update()

	if k8sResource.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(k8sResource.ID),
		gqlVar(k8sResource.Name, "name"),
		gqlVar(k8sResource.Address, "address"),
		gqlID(k8sResource.GatewayID, "gatewayId"),
		gqlID(k8sResource.RemoteNetworkID, "remoteNetworkId"),
	)

	response := query.UpdateKubernetesResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: k8sResource.ID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteKubernetesResource(ctx context.Context, resourceID string) error {
	opr := resourceKubernetesResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}
