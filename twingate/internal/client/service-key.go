package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

const serviceKeyResourceName = "service key"

func (client *Client) CreateServiceKey(ctx context.Context, serviceAccountKey *model.ServiceKey) (*model.ServiceKey, error) {
	if serviceAccountKey == nil || serviceAccountKey.Service == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "create", serviceKeyResourceName)
	}

	variables := newVars(
		gqlID(serviceAccountKey.Service, "serviceAccountId"),
		gqlVar(serviceAccountKey.ExpirationTime, "expirationTime"),
		gqlNullable(serviceAccountKey.Name, "name"),
	)

	response := query.CreateServiceAccountKey{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("createServiceAccountKey"))
	if err != nil {
		return nil, NewAPIError(err, "create", serviceKeyResourceName)
	}

	if !response.Ok {
		return nil, NewAPIError(NewMutationError(response.Error), "create", serviceKeyResourceName)
	}

	return response.ToModel() //nolint
}

func (client *Client) ReadServiceKey(ctx context.Context, serviceAccountKeyID string) (*model.ServiceKey, error) {
	if serviceAccountKeyID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", serviceKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := query.ReadServiceAccountKey{}

	err := client.GraphqlClient.Query(ctx, &response, variables, graphql.OperationName("readServiceAccountKey"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceKeyResourceName, serviceAccountKeyID)
	}

	if response.ServiceAccountKey == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceKeyResourceName, serviceAccountKeyID)
	}

	return response.ToModel() //nolint
}

func (client *Client) UpdateServiceKey(ctx context.Context, serviceAccountKey *model.ServiceKey) (*model.ServiceKey, error) {
	if serviceAccountKey == nil || serviceAccountKey.ID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "update", serviceKeyResourceName)
	}

	variables := newVars(
		gqlID(serviceAccountKey.ID),
		gqlVar(serviceAccountKey.Name, "name"),
	)

	response := query.UpdateServiceAccountKey{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("updateServiceAccountKey"))
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", serviceKeyResourceName, serviceAccountKey.ID)
	}

	if !response.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.Error), "update", serviceKeyResourceName, serviceAccountKey.ID)
	}

	if response.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", serviceKeyResourceName, serviceAccountKey.ID)
	}

	return response.ToModel() //nolint
}

func (client *Client) DeleteServiceKey(ctx context.Context, serviceAccountKeyID string) error {
	if serviceAccountKeyID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", serviceKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := query.DeleteServiceAccountKey{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("deleteServiceAccountKey"))
	if err != nil {
		return NewAPIErrorWithID(err, "delete", serviceKeyResourceName, serviceAccountKeyID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), "delete", serviceKeyResourceName, serviceAccountKeyID)
	}

	return nil
}

func (client *Client) RevokeServiceKey(ctx context.Context, serviceAccountKeyID string) error {
	if serviceAccountKeyID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "revoke", serviceKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := query.RevokeServiceAccountKey{}

	err := client.GraphqlClient.Mutate(ctx, &response, variables, graphql.OperationName("revokeServiceAccountKey"))
	if err != nil {
		return NewAPIErrorWithID(err, "revoke", serviceKeyResourceName, serviceAccountKeyID)
	}

	if !response.Ok {
		return NewAPIErrorWithID(NewMutationError(response.Error), "revoke", serviceKeyResourceName, serviceAccountKeyID)
	}

	return nil
}
