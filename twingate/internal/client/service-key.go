package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const serviceKeyResourceName = "service key"

type gqlServiceKey struct {
	IDName
	ExpiresAt      graphql.String
	Status         graphql.String
	ServiceAccount gqlServiceAccount
}

type createServiceAccountKeyQuery struct {
	ServiceAccountKeyCreate struct {
		Entity gqlServiceKey
		OkError
	} `graphql:"serviceAccountKeyCreate(expirationTime: $expirationTime, serviceAccountId: $serviceAccountId, name: $name)"`
}

func (client *Client) CreateServiceKey(ctx context.Context, serviceAccountKey *model.ServiceKey) (*model.ServiceKey, error) {
	if serviceAccountKey == nil || serviceAccountKey.Service == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "create", serviceKeyResourceName)
	}

	variables := newVars(
		gqlID(serviceAccountKey.Service, "serviceAccountId"),
		gqlField(serviceAccountKey.ExpirationTime, "expirationTime"),
		gqlNullableField(serviceAccountKey.Name, "name"),
	)

	response := createServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "createServiceAccountKey", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", serviceKeyResourceName)
	}

	if !response.ServiceAccountKeyCreate.Ok {
		message := response.ServiceAccountKeyCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", serviceKeyResourceName)
	}

	return response.ToModel()
}

type readServiceAccountKeyQuery struct {
	ServiceAccountKey *gqlServiceKey `graphql:"serviceAccountKey(id: $id)"`
}

func (client *Client) ReadServiceKey(ctx context.Context, serviceAccountKeyID string) (*model.ServiceKey, error) {
	if serviceAccountKeyID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", serviceKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := readServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readServiceAccountKey", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceKeyResourceName, serviceAccountKeyID)
	}

	if response.ServiceAccountKey == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceKeyResourceName, serviceAccountKeyID)
	}

	return response.ToModel()
}

type updateServiceAccountKeyQuery struct {
	ServiceAccountKeyUpdate struct {
		Entity *gqlServiceKey
		OkError
	} `graphql:"serviceAccountKeyUpdate(id: $id, name: $name)"`
}

func (client *Client) UpdateServiceKey(ctx context.Context, serviceAccountKey *model.ServiceKey) (*model.ServiceKey, error) {
	if serviceAccountKey == nil || serviceAccountKey.ID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "update", serviceKeyResourceName)
	}

	variables := newVars(
		gqlID(serviceAccountKey.ID),
		gqlField(serviceAccountKey.Name, "name"),
	)

	response := updateServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateServiceAccountKey", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", serviceKeyResourceName, serviceAccountKey.ID)
	}

	if !response.ServiceAccountKeyUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.ServiceAccountKeyUpdate.Error), "update", serviceKeyResourceName, serviceAccountKey.ID)
	}

	if response.ServiceAccountKeyUpdate.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", serviceKeyResourceName, serviceAccountKey.ID)
	}

	return response.ToModel()
}

type deleteServiceAccountKeyQuery struct {
	ServiceAccountKeyDelete *OkError `graphql:"serviceAccountKeyDelete(id: $id)"`
}

func (client *Client) DeleteServiceKey(ctx context.Context, serviceAccountKeyID string) error {
	if serviceAccountKeyID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", serviceKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := deleteServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "deleteServiceAccountKey", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", serviceKeyResourceName, serviceAccountKeyID)
	}

	if !response.ServiceAccountKeyDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ServiceAccountKeyDelete.Error), "delete", serviceKeyResourceName, serviceAccountKeyID)
	}

	return nil
}

type revokeServiceAccountKeyQuery struct {
	ServiceAccountKeyRevoke *OkError `graphql:"serviceAccountKeyRevoke(id: $id)"`
}

func (client *Client) RevokeServiceKey(ctx context.Context, serviceAccountKeyID string) error {
	if serviceAccountKeyID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "revoke", serviceKeyResourceName)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := revokeServiceAccountKeyQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "revokeServiceAccountKey", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "revoke", serviceKeyResourceName, serviceAccountKeyID)
	}

	if !response.ServiceAccountKeyRevoke.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ServiceAccountKeyRevoke.Error), "revoke", serviceKeyResourceName, serviceAccountKeyID)
	}

	return nil
}
