package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

const queryReadDLPPolicies = "readDLPPolicies"

func (client *Client) ReadDLPPolicy(ctx context.Context, dlpPolicyID, dlpPolicyName string) (*model.DLPPolicy, error) {
	opr := resourceDLPPolicy.read()

	if dlpPolicyID == "" && dlpPolicyName == "" {
		return nil, opr.apiError(ErrGraphqlEmptyBothNameAndID)
	}

	variables := newVars(
		gqlID(dlpPolicyID),
		gqlNullable(dlpPolicyName, "name"),
	)

	response := query.ReadDLPPolicy{}
	if err := client.query(ctx, &response, variables, opr, attr{id: dlpPolicyID}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadDLPPolicies(ctx context.Context, name, filter string) ([]*model.DLPPolicy, error) {
	opr := resourceDLPPolicy.read()

	variables := newVars(
		gqlNullable(query.NewDLPPolicyFilterField(name, filter), "filter"),
		cursor(query.CursorDLPPolicies),
		pageLimit(client.pageLimit),
	)

	response := query.ReadDLPPolicies{}

	err := client.query(ctx, &response, variables, opr.withCustomName(queryReadDLPPolicies))
	if err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			return nil, nil
		}

		return nil, err
	}

	err = response.FetchPages(ctx, client.readDLPPoliciesAfter, variables)
	if err != nil {
		return nil, opr.apiError(err)
	}

	return response.ToModel(), nil
}

func (client *Client) readDLPPoliciesAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.DLPPolicyEdge], error) {
	opr := resourceDLPPolicy.read()

	variables[query.CursorDLPPolicies] = cursor

	response := query.ReadDLPPolicies{}

	err := client.query(ctx, &response, variables, opr.withCustomName(queryReadDLPPolicies))
	if err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}