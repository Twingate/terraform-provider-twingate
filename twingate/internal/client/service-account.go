package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

const (
	queryReadServiceAccounts = "readServiceAccounts"
	queryReadServices        = "readServices"
)

func (client *Client) CreateServiceAccount(ctx context.Context, serviceAccountName string) (*model.ServiceAccount, error) {
	opr := resourceServiceAccount.create()

	if serviceAccountName == "" {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	variables := newVars(gqlVar(serviceAccountName, "name"))

	response := query.CreateServiceAccount{}
	if err := client.mutate(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadShallowServiceAccount(ctx context.Context, serviceAccountID string) (*model.ServiceAccount, error) {
	opr := resourceServiceAccount.read()

	if serviceAccountID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.ReadShallowServiceAccount{}
	if err := client.query(ctx, &response, newVars(gqlID(serviceAccountID)), opr, attr{id: serviceAccountID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) UpdateServiceAccount(ctx context.Context, serviceAccount *model.ServiceAccount) (*model.ServiceAccount, error) {
	opr := resourceServiceAccount.update()

	if serviceAccount == nil || serviceAccount.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if serviceAccount.Name == "" && len(serviceAccount.Resources) == 0 {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	variables := newVars(
		gqlID(serviceAccount.ID),
		gqlNullable(serviceAccount.Name, "name"),
		gqlIDs(serviceAccount.Resources, "addedResourceIds"),
	)

	response := query.UpdateServiceAccount{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: serviceAccount.ID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteServiceAccount(ctx context.Context, serviceAccountID string) error {
	opr := resourceServiceAccount.delete()

	if serviceAccountID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteServiceAccount{}

	return client.mutate(ctx, &response, newVars(gqlID(serviceAccountID)), opr, attr{id: serviceAccountID})
}

func (client *Client) ReadShallowServiceAccounts(ctx context.Context) ([]*model.ServiceAccount, error) {
	opr := resourceServiceAccount.read()

	variables := newVars(
		cursor(query.CursorServices),
		pageLimit(client.pageLimit),
	)

	response := query.ReadShallowServiceAccounts{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	if err := response.FetchPages(ctx, client.readServiceAccountsAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readServiceAccountsAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.ServiceAccountEdge], error) {
	opr := resourceServiceAccount.read()

	variables[query.CursorServices] = cursor

	response := query.ReadShallowServiceAccounts{}
	if err := client.query(ctx, &response, variables, opr.withCustomName(queryReadServiceAccounts), attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) ReadServiceAccounts(ctx context.Context, input ...string) ([]*model.ServiceAccount, error) {
	opr := resourceServiceAccount.read()

	var name, filter string
	if len(input) > 0 {
		name = input[0]
	}

	if len(input) > 1 {
		filter = input[1]
	}

	variables := newVars(
		gqlNullable(query.NewServiceAccountFilterInput(name, filter), "filter"),
		cursor(query.CursorServices),
		cursor(query.CursorResources),
		cursor(query.CursorServiceKeys),
		pageLimit(client.pageLimit),
	)

	response := query.ReadServiceAccounts{}
	if err := client.query(ctx, &response, variables, opr.withCustomName(queryReadServiceAccounts), attr{id: "All"}); err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			return nil, nil
		}

		return nil, err
	}

	if err := response.FetchPages(ctx, client.readServicesAfter, variables); err != nil {
		return nil, err //nolint
	}

	for i := range response.Edges {
		err := client.fetchServiceInternalResources(ctx, response.Edges[i].Node)
		if err != nil { // && !errors.Is(err, ErrGraphqlResultIsEmpty) {
			return nil, err
		}
	}

	return response.ToModel(), nil
}

func (client *Client) readServicesAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.ServiceEdge], error) {
	opr := resourceServiceAccount.read()

	variables[query.CursorServices] = cursor

	response := query.ReadServiceAccounts{}
	if err := client.query(ctx, &response, variables, opr.withCustomName(queryReadServices), attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) readServiceResourcesAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.GqlResourceIDEdge], error) {
	opr := resourceServiceAccount.read()

	gqlNullable("", query.CursorServiceKeys)(variables)
	variables[query.CursorResources] = cursor

	response := query.ReadServiceAccount{}
	if err := client.query(ctx, &response, variables, opr.withCustomName(queryReadServices), attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.Service.Resources.PaginatedResource, nil
}

func (client *Client) readServiceKeysAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.GqlKeyIDEdge], error) {
	opr := resourceServiceAccount.read()

	gqlNullable("", query.CursorResources)(variables)
	variables[query.CursorServiceKeys] = cursor

	response := query.ReadServiceAccount{}
	if err := client.query(ctx, &response, variables, opr.withCustomName(queryReadServices), attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.Service.Keys.PaginatedResource, nil
}

func (client *Client) ReadServiceAccount(ctx context.Context, serviceAccountID string) (*model.ServiceAccount, error) {
	opr := resourceServiceAccount.read()

	if serviceAccountID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(serviceAccountID),
		cursor(query.CursorResources),
		cursor(query.CursorServiceKeys),
		pageLimit(client.pageLimit),
	)

	response := query.ReadServiceAccount{}
	if err := client.query(ctx, &response, variables, opr, attr{id: serviceAccountID}); err != nil {
		return nil, err
	}

	if err := client.fetchServiceInternalResources(ctx, response.Service); err != nil {
		return nil, err
	}

	return response.Service.ToModel(), nil
}

func (client *Client) fetchServiceInternalResources(ctx context.Context, serviceAccount *query.GqlService) error {
	vars := newVars(gqlID(serviceAccount.ID), pageLimit(client.pageLimit))

	err := serviceAccount.Resources.FetchPages(ctx, client.readServiceResourcesAfter, vars)
	if err != nil {
		return err //nolint
	}

	serviceAccount.Resources.Edges = utils.Filter[*query.GqlResourceIDEdge](serviceAccount.Resources.Edges, query.IsGqlResourceActive)

	err = serviceAccount.Keys.FetchPages(ctx, client.readServiceKeysAfter, vars)
	if err != nil {
		return err //nolint
	}

	serviceAccount.Keys.Edges = utils.Filter[*query.GqlKeyIDEdge](serviceAccount.Keys.Edges, query.IsGqlKeyActive)

	return nil
}

func (client *Client) UpdateServiceAccountRemoveResources(ctx context.Context, serviceAccountID string, resourceIDsToRemove []string) error {
	opr := resourceServiceAccount.update()

	if len(resourceIDsToRemove) == 0 {
		return nil
	}

	if serviceAccountID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	_, err := client.ReadShallowServiceAccount(ctx, serviceAccountID)
	if errors.Is(err, ErrGraphqlResultIsEmpty) {
		// no-op - service does not exist
		return nil
	}

	variables := newVars(
		gqlID(serviceAccountID),
		gqlIDs(resourceIDsToRemove, "removedResourceIds"),
	)

	response := query.UpdateServiceAccountRemoveResources{}

	return client.mutate(ctx, &response, variables, opr, attr{id: serviceAccountID})
}
