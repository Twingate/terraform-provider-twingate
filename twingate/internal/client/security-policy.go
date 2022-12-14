package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const (
	securityPolicyResourceName = "security policy"

	queryReadSecurityPolicy   = "readSecurityPolicy"
	queryReadSecurityPolicies = "readSecurityPolicies"
)

func (client *Client) ReadSecurityPolicy(ctx context.Context, securityPolicyID, securityPolicyName string) (*model.SecurityPolicy, error) {
	if securityPolicyID == "" && securityPolicyName == "" {
		return nil, NewAPIError(ErrGraphqlEmptyBothNameAndID, operationRead, securityPolicyResourceName)
	}

	variables := newVars(
		gqlID(securityPolicyID),
		gqlNullable(securityPolicyName, "name"),
	)
	response := query.ReadSecurityPolicy{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadSecurityPolicy, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, securityPolicyResourceName, securityPolicyID)
	}

	if response.SecurityPolicy == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, securityPolicyResourceName, securityPolicyID)
	}

	return response.ToModel(), nil
}

func (client *Client) ReadSecurityPolicies(ctx context.Context) ([]*model.SecurityPolicy, error) {
	variables := newVars(
		gqlNullable("", query.CursorPolicies),
	)
	response := query.ReadSecurityPolicies{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadSecurityPolicies, &response, variables)
	if err != nil {
		return nil, NewAPIError(err, operationRead, securityPolicyResourceName)
	}

	if len(response.Edges) == 0 {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, operationRead, securityPolicyResourceName)
	}

	err = response.FetchPages(ctx, client.readSecurityPoliciesAfter, variables)
	if err != nil {
		return nil, NewAPIError(err, operationRead, securityPolicyResourceName)
	}

	return response.ToModel(), nil
}

func (client *Client) readSecurityPoliciesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*query.PaginatedResource[*query.SecurityPolicyEdge], error) {
	variables[query.CursorPolicies] = cursor
	response := query.ReadSecurityPolicies{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadSecurityPolicies, &response, variables)
	if err != nil {
		return nil, err //nolint
	}

	if len(response.Edges) == 0 {
		return nil, ErrGraphqlResultIsEmpty
	}

	return &response.PaginatedResource, nil
}
