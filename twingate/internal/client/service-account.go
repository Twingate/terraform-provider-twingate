package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/twingate/go-graphql-client"
)

const (
	serviceAccountResourceName = "service account"

	mutationCreateServiceAccount = "createServiceAccount"
	mutationUpdateServiceAccount = "updateServiceAccount"
	mutationDeleteServiceAccount = "deleteServiceAccount"

	queryReadServiceAccount  = "readServiceAccount"
	queryReadServiceAccounts = "readServiceAccounts"
	queryReadServices        = "readServices"
)

func (client *Client) CreateServiceAccount(ctx context.Context, serviceAccountName string) (*model.ServiceAccount, error) {
	if serviceAccountName == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, operationCreate, serviceAccountResourceName)
	}

	variables := newVars(gqlVar(serviceAccountName, "name"))
	response := query.CreateServiceAccount{}

	err := client.GraphqlClient.NamedMutate(ctx, mutationCreateServiceAccount, &response, variables)
	if err != nil {
		return nil, NewAPIError(err, operationCreate, serviceAccountResourceName)
	}

	if !response.Ok {
		return nil, NewAPIError(NewMutationError(response.Error), "create", serviceAccountResourceName)
	}

	return response.ToModel(), nil
}

func (client *Client) ReadShallowServiceAccount(ctx context.Context, serviceAccountID string) (*model.ServiceAccount, error) {
	if serviceAccountID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, operationRead, serviceAccountResourceName)
	}

	variables := newVars(gqlID(serviceAccountID))
	response := query.ReadShallowServiceAccount{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServiceAccount, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, serviceAccountID)
	}

	if response.ServiceAccount == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, serviceAccountID)
	}

	return response.ToModel(), nil
}

func (client *Client) UpdateServiceAccount(ctx context.Context, serviceAccount *model.ServiceAccount) (*model.ServiceAccount, error) {
	if serviceAccount == nil || serviceAccount.ID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, operationUpdate, serviceAccountResourceName)
	}

	if serviceAccount.Name == "" && len(serviceAccount.Resources) == 0 {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, operationUpdate, serviceAccountResourceName)
	}

	variables := newVars(
		gqlID(serviceAccount.ID),
		gqlNullable(serviceAccount.Name, "name"),
		gqlIDs(serviceAccount.Resources, "addedResourceIds"),
	)

	response := query.UpdateServiceAccount{}

	err := client.GraphqlClient.NamedMutate(ctx, mutationUpdateServiceAccount, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationUpdate, serviceAccountResourceName, serviceAccount.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), operationUpdate, serviceAccountResourceName, serviceAccount.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationUpdate, serviceAccountResourceName, serviceAccount.ID)
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteServiceAccount(ctx context.Context, serviceAccountID string) error {
	if serviceAccountID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, operationDelete, serviceAccountResourceName)
	}

	variables := newVars(gqlID(serviceAccountID))
	response := query.DeleteServiceAccount{}

	err := client.GraphqlClient.NamedMutate(ctx, mutationDeleteServiceAccount, &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, operationDelete, serviceAccountResourceName, serviceAccountID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), operationDelete, serviceAccountResourceName, serviceAccountID)
	}

	return nil
}

func (client *Client) ReadShallowServiceAccounts(ctx context.Context) ([]*model.ServiceAccount, error) {
	response := query.ReadShallowServiceAccounts{}

	variables := newVars(gqlNullable("", query.CursorServiceAccounts))

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServiceAccounts, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	err = response.FetchPages(ctx, client.readServiceAccountsAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readServiceAccountsAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.ServiceAccountEdge], error) {
	variables[query.CursorServiceAccounts] = cursor

	response := query.ReadShallowServiceAccounts{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServiceAccounts, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	return &response.PaginatedResource, nil
}

func (client *Client) ReadServiceAccounts(ctx context.Context, input ...string) ([]*model.ServiceAccount, error) {
	var name string
	if len(input) > 0 {
		name = input[0]
	}

	response := query.ReadServiceAccounts{}
	variables := newVars(
		gqlNullable(query.NewServiceAccountFilterInput(name), "filter"),
		gqlNullable("", query.CursorServices),
		gqlNullable("", query.CursorServiceResources),
		gqlNullable("", query.CursorServiceKeys),
	)

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServices, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, operationRead, serviceAccountResourceName, name)
	}

	if len(response.Edges) == 0 {
		return nil, nil
	}

	err = response.FetchPages(ctx, client.readServicesAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	for i := range response.Edges {
		err = client.fetchServiceInternalResources(ctx, response.Edges[i].Node)
		if err != nil {
			return nil, err
		}
	}

	return response.Services.ToModel(), nil
}

func (client *Client) readServicesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.ServiceEdge], error) {
	response := query.ReadServiceAccounts{}
	variables[query.CursorServices] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServices, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	return &response.PaginatedResource, nil
}

func (client *Client) readServiceResourcesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.GqlResourceIDEdge], error) {
	response := query.ReadServiceAccount{}

	gqlNullable("", query.CursorServiceKeys)(variables)
	variables[query.CursorServiceResources] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServices, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if response.Service == nil || len(response.Service.Resources.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	return &response.Service.Resources.PaginatedResource, nil
}

func (client *Client) readServiceKeysAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.GqlKeyIDEdge], error) {
	response := query.ReadServiceAccount{}

	gqlNullable("", query.CursorServiceResources)(variables)
	variables[query.CursorServiceKeys] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServices, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if response.Service == nil || len(response.Service.Keys.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	return &response.Service.Keys.PaginatedResource, nil
}

func (client *Client) ReadServiceAccount(ctx context.Context, serviceAccountID string) (*model.ServiceAccount, error) {
	if serviceAccountID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, operationRead, serviceAccountResourceName)
	}

	variables := newVars(
		gqlID(serviceAccountID),
		gqlNullable("", query.CursorServiceResources),
		gqlNullable("", query.CursorServiceKeys),
	)
	response := query.ReadServiceAccount{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServiceAccount, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, serviceAccountID)
	}

	if response.Service == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, serviceAccountID)
	}

	err = client.fetchServiceInternalResources(ctx, response.Service)
	if err != nil {
		return nil, err
	}

	return response.Service.ToModel(), nil
}

func (client *Client) fetchServiceInternalResources(ctx context.Context, serviceAccount *query.GqlService) error {
	vars := newVars(gqlID(serviceAccount.ID))

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
