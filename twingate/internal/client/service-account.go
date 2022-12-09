package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const serviceAccountResourceName = "service account"

func (client *Client) CreateServiceAccount(ctx context.Context, serviceAccountName string) (*model.ServiceAccount, error) {
	if serviceAccountName == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", serviceAccountResourceName)
	}

	variables := newVars(gqlVar(serviceAccountName, "name"))
	response := query.CreateServiceAccount{}

	err := client.GraphqlClient.NamedMutate(ctx, "createServiceAccount", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", serviceAccountResourceName)
	}

	if !response.Ok {
		return nil, NewAPIError(NewMutationError(response.Error), "create", serviceAccountResourceName)
	}

	return response.ToModel(), nil
}

func (client *Client) ReadServiceAccount(ctx context.Context, serviceAccountID string) (*model.ServiceAccount, error) {
	if serviceAccountID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", serviceAccountResourceName)
	}

	variables := newVars(gqlID(serviceAccountID))
	response := query.ReadServiceAccount{}

	err := client.GraphqlClient.NamedQuery(ctx, "readServiceAccount", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, serviceAccountID)
	}

	if response.ServiceAccount == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, serviceAccountID)
	}

	return response.ToModel(), nil
}

func (client *Client) UpdateServiceAccount(ctx context.Context, serviceAccount *model.ServiceAccount) (*model.ServiceAccount, error) {
	if serviceAccount == nil || serviceAccount.ID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "update", serviceAccountResourceName)
	}

	if serviceAccount.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "update", serviceAccountResourceName)
	}

	variables := newVars(
		gqlID(serviceAccount.ID),
		gqlVar(serviceAccount.Name, "name"),
	)

	response := query.UpdateServiceAccount{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateServiceAccount", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", serviceAccountResourceName, serviceAccount.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "update", serviceAccountResourceName, serviceAccount.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", serviceAccountResourceName, serviceAccount.ID)
	}

	return response.ToModel(), nil
}

func (client *Client) DeleteServiceAccount(ctx context.Context, serviceAccountID string) error {
	if serviceAccountID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", serviceAccountResourceName)
	}

	variables := newVars(gqlID(serviceAccountID))
	response := query.DeleteServiceAccount{}

	err := client.GraphqlClient.NamedMutate(ctx, "deleteServiceAccount", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", serviceAccountResourceName, serviceAccountID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), "delete", serviceAccountResourceName, serviceAccountID)
	}

	return nil
}

func (client *Client) ReadServiceAccounts(ctx context.Context) ([]*model.ServiceAccount, error) {
	response := query.ReadServiceAccounts{}

	variables := newVars(gqlNullable("", query.CursorServiceAccounts))

	err := client.GraphqlClient.NamedQuery(ctx, "readServiceAccounts", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, "All")
	}

	err = response.FetchPages(ctx, client.readServiceAccountsAfter, variables)
	if err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readServiceAccountsAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.ServiceAccountEdge], error) {
	variables[query.CursorServiceAccounts] = cursor

	response := query.ReadServiceAccounts{}

	err := client.GraphqlClient.NamedQuery(ctx, "readServiceAccounts", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, "All")
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, "All")
	}

	return &response.PaginatedResource, nil
}
