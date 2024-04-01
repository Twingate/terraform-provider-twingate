package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

func (client *Client) CreateServiceKey(ctx context.Context, serviceAccountKey *model.ServiceKey) (*model.ServiceKey, error) {
	opr := resourceServiceKey.create()

	if serviceAccountKey == nil || serviceAccountKey.Service == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(serviceAccountKey.Service, "serviceAccountId"),
		gqlVar(serviceAccountKey.ExpirationTime, "expirationTime"),
		gqlNullable(serviceAccountKey.Name, "name"),
	)

	response := query.CreateServiceAccountKey{}
	if err := client.mutate(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	return response.ToModel() //nolint
}

func (client *Client) ReadServiceKey(ctx context.Context, serviceAccountKeyID string) (*model.ServiceKey, error) {
	opr := resourceServiceKey.read()

	if serviceAccountKeyID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(gqlID(serviceAccountKeyID))

	response := query.ReadServiceAccountKey{}
	if err := client.query(ctx, &response, variables, opr, attr{id: serviceAccountKeyID}); err != nil {
		return nil, err
	}

	return response.ToModel() //nolint
}

func (client *Client) UpdateServiceKey(ctx context.Context, serviceAccountKey *model.ServiceKey) (*model.ServiceKey, error) {
	opr := resourceServiceKey.update()

	if serviceAccountKey == nil || serviceAccountKey.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(serviceAccountKey.ID),
		gqlVar(serviceAccountKey.Name, "name"),
	)

	response := query.UpdateServiceAccountKey{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: serviceAccountKey.ID}); err != nil {
		return nil, err
	}

	return response.ToModel() //nolint
}

func (client *Client) DeleteServiceKey(ctx context.Context, serviceAccountKeyID string) error {
	opr := resourceServiceKey.delete()

	if serviceAccountKeyID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(gqlID(serviceAccountKeyID))
	response := query.DeleteServiceAccountKey{}

	return client.mutate(ctx, &response, variables, opr, attr{id: serviceAccountKeyID})
}

func (client *Client) RevokeServiceKey(ctx context.Context, serviceAccountKeyID string) error {
	opr := resourceServiceKey.revoke()

	if serviceAccountKeyID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.RevokeServiceAccountKey{}

	return client.mutate(ctx, &response, newVars(gqlID(serviceAccountKeyID)), opr, attr{id: serviceAccountKeyID})
}
