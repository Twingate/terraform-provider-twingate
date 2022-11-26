package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const serviceAccountResourceName = "service account"

type gqlServiceAccount struct {
	IDName
}

type createServiceAccountQuery struct {
	ServiceAccountCreate struct {
		Entity IDName
		OkError
	} `graphql:"serviceAccountCreate(name: $name)"`
}

func (client *Client) CreateServiceAccount(ctx context.Context, serviceAccountName string) (*model.ServiceAccount, error) {
	if serviceAccountName == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, "create", serviceAccountResourceName)
	}

	variables := newVars(gqlField(serviceAccountName, "name"))
	response := createServiceAccountQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "createServiceAccount", &response, variables)
	if err != nil {
		return nil, NewAPIError(err, "create", serviceAccountResourceName)
	}

	if !response.ServiceAccountCreate.Ok {
		message := response.ServiceAccountCreate.Error

		return nil, NewAPIError(NewMutationError(message), "create", serviceAccountResourceName)
	}

	return response.ToModel(), nil
}

type readServiceAccountQuery struct {
	ServiceAccount *gqlServiceAccount `graphql:"serviceAccount(id: $id)"`
}

func (client *Client) ReadServiceAccount(ctx context.Context, serviceAccountID string) (*model.ServiceAccount, error) {
	if serviceAccountID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", serviceAccountResourceName)
	}

	variables := newVars(gqlID(serviceAccountID))
	response := readServiceAccountQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readServiceAccount", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, serviceAccountID)
	}

	if response.ServiceAccount == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, serviceAccountID)
	}

	return response.ToModel(), nil
}

type updateServiceAccountQuery struct {
	ServiceAccountUpdate struct {
		Entity *gqlServiceAccount
		OkError
	} `graphql:"serviceAccountUpdate(id: $id, name: $name)"`
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
		gqlField(serviceAccount.Name, "name"),
	)

	response := updateServiceAccountQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateServiceAccount", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "update", serviceAccountResourceName, serviceAccount.ID)
	}

	if !response.ServiceAccountUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.ServiceAccountUpdate.Error), "update", serviceAccountResourceName, serviceAccount.ID)
	}

	if response.ServiceAccountUpdate.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "update", serviceAccountResourceName, serviceAccount.ID)
	}

	return response.ServiceAccountUpdate.Entity.ToModel(), nil
}

type deleteServiceAccountQuery struct {
	ServiceAccountDelete *OkError `graphql:"serviceAccountDelete(id: $id)"`
}

func (client *Client) DeleteServiceAccount(ctx context.Context, serviceAccountID string) error {
	if serviceAccountID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", serviceAccountResourceName)
	}

	variables := newVars(gqlID(serviceAccountID))
	response := deleteServiceAccountQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "deleteServiceAccount", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", serviceAccountResourceName, serviceAccountID)
	}

	if !response.ServiceAccountDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ServiceAccountDelete.Error), "delete", serviceAccountResourceName, serviceAccountID)
	}

	return nil
}

type ServiceAccountEdge struct {
	Node *gqlServiceAccount
}

type ServiceAccounts struct {
	PaginatedResource[*ServiceAccountEdge]
}

type readServiceAccountsQuery struct {
	ServiceAccounts ServiceAccounts
}

func (client *Client) ReadServiceAccounts(ctx context.Context) ([]*model.ServiceAccount, error) {
	response := readServiceAccountsQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readServiceAccounts", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, "All")
	}

	if len(response.ServiceAccounts.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, "All")
	}

	err = response.ServiceAccounts.fetchPages(ctx, client.readServiceAccountsAfter, nil)
	if err != nil {
		return nil, err
	}

	return response.ServiceAccounts.ToModel(), nil
}

type readServiceAccountsAfter struct {
	ServiceAccounts ServiceAccounts `graphql:"serviceAccounts(after: $serviceAccountsEndCursor)"`
}

func (client *Client) readServiceAccountsAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*ServiceAccountEdge], error) {
	if variables == nil {
		variables = make(map[string]interface{})
	}

	variables["serviceAccountsEndCursor"] = cursor
	response := readServiceAccountsAfter{}

	err := client.GraphqlClient.NamedQuery(ctx, "readServiceAccounts", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, "All")
	}

	if len(response.ServiceAccounts.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, "All")
	}

	return &response.ServiceAccounts.PaginatedResource, nil
}

type gqlResourceID struct {
	ID graphql.ID
}

type gqlResourceIDEdge struct {
	Node *gqlResourceID
}

type gqlResourceIDs struct {
	PaginatedResource[*gqlResourceIDEdge]
}

type gqlService struct {
	IDName
	Resources gqlResourceIDs `graphql:"resources(after: $resourcesEndCursor)"`
	Keys      gqlResourceIDs
}

type ServiceEdge struct {
	Node *gqlService
}

type Services struct {
	PaginatedResource[*ServiceEdge]
}

type readServicesByNameQuery struct {
	Services Services `graphql:"serviceAccounts(filter: {name: {eq: $name}})"`
}

func (client *Client) ReadServicesByName(ctx context.Context, name string) ([]*model.Service, error) {
	if name == "" {
		return nil, NewAPIError(ErrGraphqlGroupNameIsEmpty, "read", serviceAccountResourceName)
	}

	response := readServicesByNameQuery{}
	variables := newVars(
		gqlField(name, "name"),
		gqlNullableField("", "resourcesEndCursor"),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readServices", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, "read", serviceAccountResourceName, name)
	}

	if len(response.Services.Edges) == 0 {
		return nil, nil
	}

	err = response.Services.fetchPages(ctx, client.readServicesByNameAfter, variables)
	if err != nil {
		return nil, err
	}

	for i := range response.Services.Edges {
		response.Services.Edges[i].Node.Resources.fetchPages(ctx, client.readServiceResourcesByNameAfter, variables)

	}

	return response.Services.ToModel(), nil
}

type readServicesByNameAfter struct {
	Services Services `graphql:"serviceAccounts(filter: {name: {eq: $name}}, after: $servicesEndCursor)"`
}

func (client *Client) readServicesByNameAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*ServiceEdge], error) {
	response := readServicesByNameAfter{}
	variables["servicesEndCursor"] = cursor
	variables["resourcesEndCursor"] = nil

	err := client.GraphqlClient.NamedQuery(ctx, "readServices", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, "All")
	}

	if len(response.Services.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, "All")
	}

	return &response.Services.PaginatedResource, nil
}

func (client *Client) readServiceResourcesByNameAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*gqlResourceIDEdge], error) {
	response := readServicesByNameAfter{}
	variables["servicesEndCursor"] = nil
	variables["resourcesEndCursor"] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, "readServices", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, "All")
	}

	if len(response.Services.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, "All")
	}

	return &response.Services.PaginatedResource.Edges[0].Node.Resources.PaginatedResource, nil
}

//type gqlServiceWithResources struct {
//	IDName
//	Resources gqlResourceIDs `graphql:"resources(after: $resourcesEndCursor)"`
//	Keys      gqlResourceIDs
//}
//
//type ServiceEdgeWithResources struct {
//	Node *gqlServiceWithResources
//}
//
//type ServicesWithResources struct {
//	PaginatedResource[*ServiceEdge]
//}
//
//type readServicesByNameQuery struct {
//	Services Services `graphql:"serviceAccounts(filter: {name: {eq: $name}})"`
//}
//
//// after: $servicesEndCursor
//type readResourceServicesByNameAfter struct {
//	Services Services `graphql:"serviceAccounts(filter: {name: {eq: $name}})"`
//}
//
//func (client *Client) readResourceServicesByNameAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*ServiceEdge], error) {
//	response := readServicesByNameAfter{}
//	variables["servicesEndCursor"] = cursor
//
//	err := client.GraphqlClient.NamedQuery(ctx, "readServices", &response, variables)
//	if err != nil {
//		return nil, NewAPIErrorWithID(err, "read", serviceAccountResourceName, "All")
//	}
//
//	if len(response.Services.Edges) == 0 {
//		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", serviceAccountResourceName, "All")
//	}
//
//	return &response.Services.PaginatedResource, nil
//}
