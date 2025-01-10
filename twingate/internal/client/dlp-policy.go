package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

func (client *Client) ReadDLPPolicy(ctx context.Context, policy *model.DLPPolicy) (*model.DLPPolicy, error) {
	opr := resourceDLPPolicy.read()

	if policy == nil || policy.ID == "" && policy.Name == "" {
		return nil, opr.apiError(ErrGraphqlEmptyBothNameAndID)
	}

	variables := newVars(
		gqlNullableID(policy.ID, "id"),
		gqlNullable(policy.Name, "name"),
	)

	response := query.ReadDLPPolicy{}
	if err := client.query(ctx, &response, variables, opr, attr{id: policy.ID, name: policy.Name}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadDLPPolicies(ctx context.Context, name, filter string) ([]*model.DLPPolicy, error) {
	opr := resourceDLPPolicy.read().withCustomName("readDLPPolicies")

	variables := newVars(
		gqlNullable(query.NewDLPPoliciesFilterInput(name, filter), "filter"),
		cursor(query.CursorDLPPolicies),
		pageLimit(client.pageLimit),
	)

	response := query.ReadDLPPolicies{}
	if err := client.query(ctx, &response, variables, opr,
		attr{id: "All", name: name}); err != nil {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.FetchPages(oprCtx, client.readDLPPoliciesAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readDLPPoliciesAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.DLPPolicyEdge], error) {
	opr := resourceDLPPolicy.read().withCustomName("readDLPPoliciesAfter")

	variables[query.CursorDLPPolicies] = cursor

	response := query.ReadDLPPolicies{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}
