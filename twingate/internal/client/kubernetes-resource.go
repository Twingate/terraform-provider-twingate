package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
)

//nolint:dupl
func (client *Client) CreateKubernetesResource(ctx context.Context, k8sResource *model.KubernetesResource) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.create()

	variables := newVars(
		gqlVar(k8sResource.Name, "name"),
		gqlVar(k8sResource.Address, "address"),
		gqlID(k8sResource.GatewayID, "gatewayId"),
		gqlID(k8sResource.RemoteNetworkID, "remoteNetworkId"),
		gqlNullable(k8sResource.IsVisible, "isVisible"),
		gqlNullableEmpty(k8sResource.Alias, "alias"),
		gqlNullableID(k8sResource.SecurityPolicyID, "securityPolicyId"),
		gqlVar(newTagInputs(k8sResource.Tags), "tags"),
		gqlVar(newProtocolsInput(k8sResource.Protocols), "protocols"),
		gqlVar(NewAccessPolicyInput(k8sResource.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(k8sResource.AccessPolicy), "approvalMode"),
	)

	response := query.CreateKubernetesResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{name: k8sResource.Name}); err != nil {
		return nil, err
	}

	res := response.ToModel()
	if res == nil {
		return nil, nil //nolint:nilnil
	}

	if len(k8sResource.GroupsAccess) > 0 {
		if err := client.AddResourceAccess(ctx, res.ID, convertGroupsToAccessInput(k8sResource.GroupsAccess)); err != nil {
			return nil, err
		}
	}

	res.GroupsAccess = k8sResource.GroupsAccess

	return res, nil
}

func convertGroupsToAccessInput(groups []model.AccessGroup) []AccessInput {
	access := make([]AccessInput, 0, len(groups))

	for _, group := range groups {
		access = append(access, AccessInput{
			PrincipalID:      group.GroupID,
			SecurityPolicyID: group.SecurityPolicyID,
			ApprovalMode:     NewGroupAccessApprovalMode(group.AccessPolicy),
			AccessPolicy:     NewAccessPolicyInput(group.AccessPolicy),
		})
	}

	return access
}

func (client *Client) ReadKubernetesResource(ctx context.Context, resourceID string) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.read()

	if resourceID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(resourceID),
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)
	response := query.ReadKubernetesResource{}

	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	if err := response.Resource.Access.FetchPages(withOperationCtx(ctx, opr), client.readKubernetesResourceAccessAfter, newVars(gqlID(resourceID))); err != nil {
		return nil, err //nolint
	}

	return response.ToModel() //nolint:wrapcheck
}

func (client *Client) readKubernetesResourceAccessAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
	opr := resourceKubernetesResource.read().withCustomName("readKubernetesResourceAccessAfter")

	variables[query.CursorAccess] = cursor
	pageLimit(client.pageLimit)(variables)

	response := query.ReadKubernetesResource{}
	if err := client.query(ctx, &response, variables, opr, attr{}); err != nil {
		return nil, err
	}

	return &response.Resource.Access.PaginatedResource, nil
}

func (client *Client) UpdateKubernetesResource(ctx context.Context, k8sResource *model.KubernetesResource) (*model.KubernetesResource, error) {
	opr := resourceKubernetesResource.update()

	if k8sResource.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(k8sResource.ID),
		gqlVar(k8sResource.Name, "name"),
		gqlVar(k8sResource.Address, "address"),
		gqlID(k8sResource.GatewayID, "gatewayId"),
		gqlID(k8sResource.RemoteNetworkID, "remoteNetworkId"),
		gqlNullable(k8sResource.IsVisible, "isVisible"),
		gqlNullableEmpty(k8sResource.Alias, "alias"),
		gqlNullableID(k8sResource.SecurityPolicyID, "securityPolicyId"),
		gqlVar(newTagInputs(k8sResource.Tags), "tags"),
		gqlVar(newProtocolsInput(k8sResource.Protocols), "protocols"),
		gqlVar(NewAccessPolicyInput(k8sResource.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(k8sResource.AccessPolicy), "approvalMode"),
	)

	response := query.UpdateKubernetesResource{}

	if err := client.mutate(ctx, &response, variables, opr, attr{id: k8sResource.ID}); err != nil {
		return nil, err
	}

	res := response.ToModel()
	if res == nil {
		return nil, nil //nolint:nilnil
	}

	res.GroupsAccess = k8sResource.GroupsAccess

	return res, nil
}

func (client *Client) DeleteKubernetesResource(ctx context.Context, resourceID string) error {
	opr := resourceKubernetesResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}
