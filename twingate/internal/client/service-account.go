package client

import (
	"context"

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

	cursorServiceAccounts  = "serviceAccountsEndCursor"
	cursorServices         = "servicesEndCursor"
	cursorServiceResources = "resourcesEndCursor"
	cursorServiceKeys      = "keysEndCursor"
)

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
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, operationCreate, serviceAccountResourceName)
	}

	variables := newVars(gqlField(serviceAccountName, "name"))
	response := createServiceAccountQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, mutationCreateServiceAccount, &response, variables)
	if err != nil {
		return nil, NewAPIError(err, operationCreate, serviceAccountResourceName)
	}

	if !response.ServiceAccountCreate.Ok {
		message := response.ServiceAccountCreate.Error

		return nil, NewAPIError(NewMutationError(message), operationCreate, serviceAccountResourceName)
	}

	return response.ToModel(), nil
}

type readServiceAccountQuery struct {
	ServiceAccount *gqlServiceAccount `graphql:"serviceAccount(id: $id)"`
}

func (client *Client) ReadServiceAccount(ctx context.Context, serviceAccountID string) (*model.ServiceAccount, error) {
	if serviceAccountID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, operationRead, serviceAccountResourceName)
	}

	variables := newVars(gqlID(serviceAccountID))
	response := readServiceAccountQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServiceAccount, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, serviceAccountID)
	}

	if response.ServiceAccount == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, serviceAccountID)
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
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, operationUpdate, serviceAccountResourceName)
	}

	if serviceAccount.Name == "" {
		return nil, NewAPIError(ErrGraphqlNameIsEmpty, operationUpdate, serviceAccountResourceName)
	}

	variables := newVars(
		gqlID(serviceAccount.ID),
		gqlField(serviceAccount.Name, "name"),
	)

	response := updateServiceAccountQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, mutationUpdateServiceAccount, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationUpdate, serviceAccountResourceName, serviceAccount.ID)
	}

	if !response.ServiceAccountUpdate.Ok {
		return nil, NewAPIErrorWithID(NewMutationError(response.ServiceAccountUpdate.Error), operationUpdate, serviceAccountResourceName, serviceAccount.ID)
	}

	if response.ServiceAccountUpdate.Entity == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationUpdate, serviceAccountResourceName, serviceAccount.ID)
	}

	return response.ServiceAccountUpdate.Entity.ToModel(), nil
}

type deleteServiceAccountQuery struct {
	ServiceAccountDelete *OkError `graphql:"serviceAccountDelete(id: $id)"`
}

func (client *Client) DeleteServiceAccount(ctx context.Context, serviceAccountID string) error {
	if serviceAccountID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, operationDelete, serviceAccountResourceName)
	}

	variables := newVars(gqlID(serviceAccountID))
	response := deleteServiceAccountQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, mutationDeleteServiceAccount, &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, operationDelete, serviceAccountResourceName, serviceAccountID)
	}

	if !response.ServiceAccountDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ServiceAccountDelete.Error), operationDelete, serviceAccountResourceName, serviceAccountID)
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

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServiceAccounts, &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if len(response.ServiceAccounts.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
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

	variables[cursorServiceAccounts] = cursor
	response := readServiceAccountsAfter{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServiceAccounts, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if len(response.ServiceAccounts.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	return &response.ServiceAccounts.PaginatedResource, nil
}

type gqlResourceID struct {
	ID       graphql.ID
	IsActive graphql.Boolean
}

func (r gqlResourceID) isActive() bool {
	return bool(r.IsActive)
}

type gqlResourceIDEdge struct {
	Node *gqlResourceID
}

func isGqlResourceActive(item *gqlResourceIDEdge) bool {
	return item.Node.isActive()
}

type gqlResourceIDs struct {
	PaginatedResource[*gqlResourceIDEdge]
}

type gqlKeyID struct {
	ID     graphql.ID
	Status graphql.String
}

func (k gqlKeyID) isActive() bool {
	return string(k.Status) == model.StatusActive
}

type gqlKeyIDEdge struct {
	Node *gqlKeyID
}

func isGqlKeyActive(item *gqlKeyIDEdge) bool {
	return item.Node.isActive()
}

type gqlKeyIDs struct {
	PaginatedResource[*gqlKeyIDEdge]
}

type gqlService struct {
	IDName
	Resources gqlResourceIDs `graphql:"resources(after: $resourcesEndCursor)"`
	Keys      gqlKeyIDs      `graphql:"keys(after: $keysEndCursor)"`
}

type ServiceEdge struct {
	Node *gqlService
}

type Services struct {
	PaginatedResource[*ServiceEdge]
}

type readServicesByNameQuery struct {
	Services Services `graphql:"serviceAccounts(filter: $filter, after: $servicesEndCursor)"`
}

type ServiceAccountFilterInput struct {
	Name StringFilter `json:"name"`
}

type StringFilter struct {
	Eq graphql.String `json:"eq"`
}

func newServiceAccountFilterInput(name string) *ServiceAccountFilterInput {
	if name == "" {
		return nil
	}

	return &ServiceAccountFilterInput{
		Name: StringFilter{
			Eq: graphql.String(name),
		},
	}
}

func (client *Client) ReadServices(ctx context.Context, input ...string) ([]*model.Service, error) {
	var name string
	if len(input) > 0 {
		name = input[0]
	}

	response := readServicesByNameQuery{}
	variables := newVars(
		gqlNullableField(newServiceAccountFilterInput(name), "filter"),
		gqlNullableField("", cursorServices),
		gqlNullableField("", cursorServiceResources),
		gqlNullableField("", cursorServiceKeys),
	)

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServices, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithName(err, operationRead, serviceAccountResourceName, name)
	}

	if len(response.Services.Edges) == 0 {
		return nil, nil
	}

	err = response.Services.fetchPages(ctx, client.readServicesAfter, variables)
	if err != nil {
		return nil, err
	}

	for i := range response.Services.Edges {
		err = response.Services.Edges[i].Node.fetchInternalResources(ctx, client)
		if err != nil {
			return nil, err
		}
	}

	return response.Services.ToModel(), nil
}

func (client *Client) readServicesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*ServiceEdge], error) {
	response := readServicesByNameQuery{}
	variables[cursorServices] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServices, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if len(response.Services.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	return &response.Services.PaginatedResource, nil
}

type readServiceQuery struct {
	Service *gqlService `graphql:"serviceAccount(id: $id)"`
}

func (client *Client) readServiceResourcesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*gqlResourceIDEdge], error) {
	response := readServiceQuery{}

	gqlNullableField("", cursorServiceKeys)(variables)
	variables[cursorServiceResources] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServices, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if len(response.Service.Resources.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	return &response.Service.Resources.PaginatedResource, nil
}

func (client *Client) readServiceKeysAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*gqlKeyIDEdge], error) {
	response := readServiceQuery{}

	gqlNullableField("", cursorServiceResources)(variables)
	variables[cursorServiceKeys] = cursor

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServices, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, "All")
	}

	if len(response.Service.Keys.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, "All")
	}

	return &response.Service.Keys.PaginatedResource, nil
}

type readServiceWithResourceAndKeysQuery struct {
	ServiceAccount *gqlService `graphql:"serviceAccount(id: $id)"`
}

func (client *Client) ReadService(ctx context.Context, serviceAccountID string) (*model.Service, error) {
	if serviceAccountID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, operationRead, serviceAccountResourceName)
	}

	variables := newVars(
		gqlID(serviceAccountID),
		gqlNullableField("", cursorServiceResources),
		gqlNullableField("", cursorServiceKeys),
	)
	response := readServiceWithResourceAndKeysQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadServiceAccount, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, serviceAccountResourceName, serviceAccountID)
	}

	if response.ServiceAccount == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, serviceAccountResourceName, serviceAccountID)
	}

	err = response.ServiceAccount.fetchInternalResources(ctx, client)
	if err != nil {
		return nil, err
	}

	return response.ServiceAccount.ToModel(), nil
}

func (s *gqlService) fetchInternalResources(ctx context.Context, client *Client) error {
	vars := newVars(gqlID(s.ID))

	err := s.Resources.fetchPages(ctx, client.readServiceResourcesAfter, vars)
	if err != nil {
		return err
	}

	s.Resources.Edges = utils.Filter[*gqlResourceIDEdge](s.Resources.Edges, isGqlResourceActive)

	err = s.Keys.fetchPages(ctx, client.readServiceKeysAfter, vars)
	if err != nil {
		return err
	}

	s.Keys.Edges = utils.Filter[*gqlKeyIDEdge](s.Keys.Edges, isGqlKeyActive)

	return nil
}
