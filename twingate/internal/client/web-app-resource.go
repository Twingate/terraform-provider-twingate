package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/utils"
)

// WebAppUpstreamInput and the input types below must keep their names in sync
// with the GraphQL schema: the client derives the variable type from the Go type
// name, so a rename silently produces an invalid query.
type WebAppUpstreamInput struct {
	Port int64 `json:"port"`
}

type WebAppDownstreamInput struct {
	Port int64 `json:"port"`
}

type KeyValueInputObject struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// newKeyValueInputs always returns a non-nil slice so that an empty map is sent
// as `[]` rather than omitted. The API clears the stored header rewrites on an
// empty list, which keeps `null` and `{}` in config converging on the same state.
func newKeyValueInputs(pairs map[string]string) []KeyValueInputObject {
	inputs := make([]KeyValueInputObject, 0, len(pairs))

	for key, value := range pairs {
		inputs = append(inputs, KeyValueInputObject{
			Key:   key,
			Value: value,
		})
	}

	return inputs
}

func newWebAppResourceVars(webAppResource *model.WebAppResource) []gqlVarOption {
	return []gqlVarOption{
		gqlVar(webAppResource.Name, "name"),
		gqlVar(webAppResource.Address, "address"),
		gqlID(webAppResource.GatewayID, "gatewayId"),
		gqlID(webAppResource.RemoteNetworkID, "remoteNetworkId"),
		gqlNullable(webAppResource.IsVisible, "isVisible"),
		gqlNullable(webAppResource.Alias, "alias"),
		gqlNullableID(webAppResource.SecurityPolicyID, "securityPolicyId"),
		gqlVar(newTagInputs(webAppResource.Tags), "tags"),
		gqlVar(WebAppUpstreamInput{Port: webAppResource.UpstreamPort}, "upstream"),
		gqlVar(WebAppDownstreamInput{Port: webAppResource.DownstreamPort}, "downstream"),
		gqlVar(newKeyValueInputs(webAppResource.RequestHeaderRewrites), "requestHeaderRewrites"),
		gqlVar(NewAccessPolicyInput(webAppResource.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(webAppResource.AccessPolicy), "approvalMode"),
	}
}

func (client *Client) CreateWebAppResource(ctx context.Context, webAppResource *model.WebAppResource) (*model.WebAppResource, error) {
	opr := resourceWebAppResource.create()

	variables := newVars(newWebAppResourceVars(webAppResource)...)

	response := query.CreateWebAppResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: webAppResource.Name}); err != nil {
		return nil, err
	}

	// client.mutate already fails with ErrGraphqlResultIsEmpty when the payload
	// entity is missing, so ToModel never returns nil here.
	res := response.ToModel()

	if len(webAppResource.GroupsAccess) > 0 {
		if err := client.AddResourceAccess(ctx, res.ID, convertGroupsToAccessInput(webAppResource.GroupsAccess)); err != nil {
			return nil, err
		}
	}

	res.GroupsAccess = webAppResource.GroupsAccess

	return res, nil
}

func (client *Client) ReadWebAppResource(ctx context.Context, resourceID string) (*model.WebAppResource, error) {
	opr := resourceWebAppResource.read()

	if resourceID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(resourceID),
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)
	response := query.ReadWebAppResource{}

	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	if err := response.Resource.Access.FetchPages(withOperationCtx(ctx, opr), client.readWebAppResourceAccessAfter, newVars(gqlID(resourceID))); err != nil {
		return nil, err //nolint
	}

	return response.ToModel() //nolint:wrapcheck
}

func (client *Client) readWebAppResourceAccessAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
	opr := resourceWebAppResource.read().withCustomName("readWebAppResourceAccessAfter")

	variables[query.CursorAccess] = cursor
	pageLimit(client.pageLimit)(variables)

	response := query.ReadWebAppResource{}
	if err := client.query(ctx, &response, variables, opr, attr{}); err != nil {
		return nil, err
	}

	return &response.Resource.Access.PaginatedResource, nil
}

func (client *Client) UpdateWebAppResource(ctx context.Context, webAppResource *model.WebAppResource) (*model.WebAppResource, error) {
	opr := resourceWebAppResource.update()

	if webAppResource.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(append(
		[]gqlVarOption{gqlID(webAppResource.ID)},
		newWebAppResourceVars(webAppResource)...,
	)...)

	response := query.UpdateWebAppResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: webAppResource.ID}); err != nil {
		return nil, err
	}

	res := response.ToModel()

	res.GroupsAccess = webAppResource.GroupsAccess

	return res, nil
}

func (client *Client) ReadWebAppResources(ctx context.Context) ([]*model.WebAppResource, error) {
	opr := resourceWebAppResource.read().withCustomName("readWebAppResources")

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
			return edge.Node.Type == "WebAppResource"
		},
		func(edge *query.ShallowResourceEdge) *model.WebAppResource {
			return &model.WebAppResource{
				ID:   string(edge.Node.ID),
				Name: edge.Node.Name,
			}
		}), nil
}

func (client *Client) DeleteWebAppResource(ctx context.Context, resourceID string) error {
	opr := resourceWebAppResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}
