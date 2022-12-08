package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const (
	securityPolicyResourceName = "security policy"

	queryReadSecurityPolicy   = "readSecurityPolicy"
	queryReadSecurityPolicies = "readSecurityPolicies"

	cursorGroups   = "groupsEndCursor"
	cursorPolicies = "policiesEndCursor"
)

type gqlSecurityPolicy struct {
	IDName
	PolicyType graphql.String
	Groups     Groups `graphql:"groups(after: $groupsEndCursor)"`
}

type readSecurityPolicyQuery struct {
	SecurityPolicy *gqlSecurityPolicy `graphql:"securityPolicy(id: $id, name: $name)"`
}

func (client *Client) ReadSecurityPolicy(ctx context.Context, securityPolicyID, securityPolicyName string) (*model.SecurityPolicy, error) {
	if securityPolicyID == "" && securityPolicyName == "" {
		return nil, NewAPIError(ErrGraphqlEmptyBothNameAndID, operationRead, securityPolicyResourceName)
	}

	variables := newVars(
		gqlID(securityPolicyID),
		gqlNullableField(securityPolicyName, "name"),
		gqlNullableField("", cursorGroups),
	)
	response := readSecurityPolicyQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadSecurityPolicy, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, securityPolicyResourceName, securityPolicyID)
	}

	if response.SecurityPolicy == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, securityPolicyResourceName, securityPolicyID)
	}

	err = response.SecurityPolicy.Groups.fetchPages(ctx, client.readSecurityPolicyGroupsAfter, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, securityPolicyResourceName, securityPolicyID)
	}

	return response.ToModel(), nil
}

func (client *Client) readSecurityPolicyGroupsAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*GroupEdge], error) {
	variables[cursorGroups] = cursor
	response := readSecurityPolicyQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadSecurityPolicy, &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, operationRead, groupResourceName, "All")
	}

	if response.SecurityPolicy == nil || len(response.SecurityPolicy.Groups.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, operationRead, groupResourceName, "All")
	}

	return &response.SecurityPolicy.Groups.PaginatedResource, nil
}

type SecurityPolicyEdge struct {
	Node *gqlSecurityPolicy
}

type SecurityPolicies struct {
	PaginatedResource[*SecurityPolicyEdge]
}

type readSecurityPoliciesQuery struct {
	SecurityPolicies SecurityPolicies `graphql:"securityPolicies(after: $policiesEndCursor)"`
}

func (client *Client) ReadSecurityPolicies(ctx context.Context) ([]*model.SecurityPolicy, error) {
	variables := newVars(
		gqlNullableField("", cursorPolicies),
		gqlNullableField("", cursorGroups),
	)
	response := readSecurityPoliciesQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadSecurityPolicies, &response, variables)
	if err != nil {
		return nil, NewAPIError(err, operationRead, securityPolicyResourceName)
	}

	if len(response.SecurityPolicies.Edges) == 0 {
		return nil, NewAPIError(ErrGraphqlResultIsEmpty, operationRead, securityPolicyResourceName)
	}

	err = response.SecurityPolicies.fetchPages(ctx, client.readSecurityPoliciesAfter, variables)
	if err != nil {
		return nil, NewAPIError(err, operationRead, securityPolicyResourceName)
	}

	for i, edge := range response.SecurityPolicies.Edges {
		securityPolicyID := edge.Node.StringID()

		err = response.SecurityPolicies.Edges[i].Node.Groups.fetchPages(ctx,
			client.readSecurityPolicyGroupsAfter,
			newVars(
				gqlID(securityPolicyID),
				gqlNullableField("", "name"),
			),
		)
		if err != nil {
			return nil, NewAPIErrorWithID(err, operationRead, securityPolicyResourceName, securityPolicyID)
		}
	}

	return response.ToModel(), nil
}

func (client *Client) readSecurityPoliciesAfter(ctx context.Context, variables map[string]interface{}, cursor graphql.String) (*PaginatedResource[*SecurityPolicyEdge], error) {
	variables[cursorPolicies] = cursor
	response := readSecurityPoliciesQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, queryReadSecurityPolicies, &response, variables)
	if err != nil {
		return nil, err //nolint
	}

	if len(response.SecurityPolicies.Edges) == 0 {
		return nil, ErrGraphqlResultIsEmpty
	}

	return &response.SecurityPolicies.PaginatedResource, nil
}
