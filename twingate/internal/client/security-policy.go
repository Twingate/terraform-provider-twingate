package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

const queryReadSecurityPolicies = "readSecurityPolicies"

func (client *Client) ReadSecurityPolicy(ctx context.Context, securityPolicyID, securityPolicyName string) (*model.SecurityPolicy, error) {
	opr := resourceSecurityPolicy.read()

	if securityPolicyID == "" && securityPolicyName == "" {
		return nil, opr.apiError(ErrGraphqlEmptyBothNameAndID)
	}

	variables := newVars(
		gqlID(securityPolicyID),
		gqlNullable(securityPolicyName, "name"),
	)

	response := query.ReadSecurityPolicy{}
	if err := client.query(ctx, &response, variables, opr, attr{id: securityPolicyID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadSecurityPolicies(ctx context.Context, name, filter string) ([]*model.SecurityPolicy, error) {
	opr := resourceSecurityPolicy.read()

	variables := newVars(
		gqlNullable(query.NewSecurityPolicyFilterField(name, filter), "filter"),
		cursor(query.CursorPolicies),
		pageLimit(client.pageLimit),
	)

	response := query.ReadSecurityPolicies{}

	err := client.query(ctx, &response, variables, opr.withCustomName(queryReadSecurityPolicies))
	if err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			return nil, nil
		}

		return nil, err
	}

	err = response.FetchPages(ctx, client.readSecurityPoliciesAfter, variables)
	if err != nil {
		return nil, opr.apiError(err)
	}

	return response.ToModel(), nil
}

func (client *Client) readSecurityPoliciesAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.SecurityPolicyEdge], error) {
	opr := resourceSecurityPolicy.read()

	variables[query.CursorPolicies] = cursor

	response := query.ReadSecurityPolicies{}

	err := client.query(ctx, &response, variables, opr.withCustomName(queryReadSecurityPolicies))
	if err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}
