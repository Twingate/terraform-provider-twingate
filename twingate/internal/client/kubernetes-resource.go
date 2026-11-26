package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/utils"
)

func (client *Client) CreateKubernetesResource(ctx context.Context, k8sResource *model.KubernetesResource) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.create()

	variables := pamVars(k8sResource.PAMResource)

	response := query.CreateKubernetesResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: k8sResource.Name}); err != nil {
		return nil, err
	}

	res := response.ToModel()

	if err := client.addPAMResourceAccess(ctx, &res.PAMResource, k8sResource.PAMResource); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) ReadKubernetesResource(ctx context.Context, resourceID string) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.read()

	if resourceID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(resourceID),
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)
	response := query.ReadKubernetesResource{}

	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	if err := response.Resource.Access.FetchPages(withOperationCtx(ctx, opr), client.readKubernetesResourceAccessAfter, newVars(gqlID(resourceID))); err != nil {
		return nil, err //nolint
	}

	return response.ToModel() //nolint:wrapcheck
}

func (client *Client) readKubernetesResourceAccessAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
	opr := resourceKubernetesResource.read().withCustomName("readKubernetesResourceAccessAfter")

	variables[query.CursorAccess] = cursor
	pageLimit(client.pageLimit)(variables)

	response := query.ReadKubernetesResource{}
	if err := client.query(ctx, &response, variables, opr, attr{}); err != nil {
		return nil, err
	}

	return &response.Resource.Access.PaginatedResource, nil
}

func (client *Client) UpdateKubernetesResource(ctx context.Context, k8sResource *model.KubernetesResource) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.update()

	if k8sResource.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := pamVars(k8sResource.PAMResource, gqlID(k8sResource.ID))

	response := query.UpdateKubernetesResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: k8sResource.ID}); err != nil {
		return nil, err
	}

	res := response.ToModel()

	copyAccess(&res.BaseResource, k8sResource.BaseResource)

	return res, nil
}

func (client *Client) ReadKubernetesResources(ctx context.Context) ([]*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.read().withCustomName("readKubernetesResources")

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
			return edge.Node.Type == "KubernetesResource"
		},
		func(edge *query.ShallowResourceEdge) *model.KubernetesResource {
			return &model.KubernetesResource{
				PAMResource: model.PAMResource{
					BaseResource: model.BaseResource{
						ID:   string(edge.Node.ID),
						Name: edge.Node.Name,
					},
				},
			}
		}), nil
}

func (client *Client) DeleteKubernetesResource(ctx context.Context, resourceID string) error {
	opr := resourceKubernetesResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}
