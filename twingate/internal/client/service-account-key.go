package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const serviceAccountKeyResourceName = "service account key"

type gqlServiceAccountKey struct {
	IDName
	ExpiresAt      graphql.String
	Status         graphql.String
	ServiceAccount gqlServiceAccount
}

type createServiceAccountKeyQuery struct {
	ServiceAccountKeyCreate struct {
		Entity gqlServiceAccountKey
		OkError
	} `graphql:"serviceAccountKeyCreate(expirationTime: $expirationTime, serviceAccountId: $serviceAccountId, name: $name)"`
}

func (client *Client) CreateServiceAccountKey(ctx context.Context, serviceAccountKey *model.ServiceAccountKey) (*model.ServiceAccountKey, error) {
	if serviceAccountKey == nil || serviceAccountKey.ServiceAccountID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "create", serviceAccountKeyResourceName)
	}

	variables := newVars(
		gqlID(serviceAccountKey.ServiceAccountID, "serviceAccountId"),
		gqlField(serviceAccountKey.ExpirationTime, "expirationTime"),
		gqlNullableField(serviceAccountKey.Name, "name"),
	)

	response := createServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "createServiceAccountKey", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", serviceAccountKeyResourceName)
	}

	if !response.ServiceAccountKeyCreate.Ok {
		message := response.ServiceAccountKeyCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", serviceAccountKeyResourceName)
	}

	return response.ToModel()
}

type readServiceAccountKeyQuery struct {
	ServiceAccountKey *gqlServiceAccountKey `graphql:"serviceAccountKey(id: $id)"`
}

func (client *Client) ReadServiceAccountKey(ctx context.Context, serviceAccountKeyID string) (*model.ServiceAccountKey, error) {
	if serviceAccountKeyID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", serviceAccountKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := readServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readServiceAccountKey", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountKeyResourceName, serviceAccountKeyID)
	}

	if response.ServiceAccountKey == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountKeyResourceName, serviceAccountKeyID)
	}

	return response.ToModel()
}

type updateServiceAccountKeyQuery struct {
	ServiceAccountKeyUpdate struct {
		Entity *gqlServiceAccountKey
		OkError
	} `graphql:"serviceAccountKeyUpdate(id: $id, name: $name)"`
}

func (client *Client) UpdateServiceAccountKey(ctx context.Context, serviceAccountKey *model.ServiceAccountKey) (*model.ServiceAccountKey, error) {
	if serviceAccountKey == nil || serviceAccountKey.ID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "update", serviceAccountKeyResourceName)
	}

	variables := newVars(
		gqlID(serviceAccountKey.ID),
		gqlField(serviceAccountKey.Name, "name"),
	)

	response := updateServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateServiceAccountKey", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", serviceAccountKeyResourceName, serviceAccountKey.ID)
	}

	if !response.ServiceAccountKeyUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.ServiceAccountKeyUpdate.Error), "update", serviceAccountKeyResourceName, serviceAccountKey.ID)
	}

	if response.ServiceAccountKeyUpdate.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", serviceAccountKeyResourceName, serviceAccountKey.ID)
	}

	return response.ToModel()
}

type deleteServiceAccountKeyQuery struct {
	ServiceAccountKeyDelete *OkError `graphql:"serviceAccountKeyDelete(id: $id)"`
}

func (client *Client) DeleteServiceAccountKey(ctx context.Context, serviceAccountKeyID string) error {
	if serviceAccountKeyID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", serviceAccountKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := deleteServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "deleteServiceAccountKey", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", serviceAccountKeyResourceName, serviceAccountKeyID)
	}

	if !response.ServiceAccountKeyDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ServiceAccountKeyDelete.Error), "delete", serviceAccountKeyResourceName, serviceAccountKeyID)
	}

	return nil
}

type revokeServiceAccountKeyQuery struct {
	ServiceAccountKeyRevoke *OkError `graphql:"serviceAccountKeyRevoke(id: $id)"`
}

func (client *Client) RevokeServiceAccountKey(ctx context.Context, serviceAccountKeyID string) error {
	if serviceAccountKeyID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "revoke", serviceAccountKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := revokeServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "revokeServiceAccountKey", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "revoke", serviceAccountKeyResourceName, serviceAccountKeyID)
	}

	if !response.ServiceAccountKeyRevoke.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ServiceAccountKeyRevoke.Error), "revoke", serviceAccountKeyResourceName, serviceAccountKeyID)
	}

	return nil
}
