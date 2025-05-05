package client

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hasura/go-graphql-client"
)

type TagInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func newTagInputs(tags map[string]string) []TagInput {
	tagInputs := make([]TagInput, 0, len(tags))

	if len(tags) == 0 {
		return tagInputs
	}

	for k, v := range tags {
		tagInputs = append(tagInputs, TagInput{
			Key:   k,
			Value: v,
		})
	}

	return tagInputs
}

type ProtocolsInput struct {
	UDP       *ProtocolInput `json:"udp"`
	TCP       *ProtocolInput `json:"tcp"`
	AllowIcmp bool           `json:"allowIcmp"`
}

type ProtocolInput struct {
	Ports  []*PortRangeInput `json:"ports"`
	Policy string            `json:"policy"`
}

type PortRangeInput struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func newProtocolsInput(protocols *model.Protocols) *ProtocolsInput {
	if protocols == nil {
		return nil
	}

	return &ProtocolsInput{
		UDP:       newProtocol(protocols.UDP),
		TCP:       newProtocol(protocols.TCP),
		AllowIcmp: protocols.AllowIcmp,
	}
}

func newProtocol(protocol *model.Protocol) *ProtocolInput {
	if protocol == nil {
		return nil
	}

	return &ProtocolInput{
		Ports:  newPorts(protocol.Ports),
		Policy: protocol.Policy,
	}
}

func newPorts(ports []*model.PortRange) []*PortRangeInput {
	return utils.Map[*model.PortRange, *PortRangeInput](ports, func(port *model.PortRange) *PortRangeInput {
		return &PortRangeInput{
			Start: port.Start,
			End:   port.End,
		}
	})
}

type AccessApprovalMode string

func NewAccessApprovalMode(approvalMode string) *AccessApprovalMode {
	if approvalMode == "" {
		return nil
	}

	mode := AccessApprovalMode(approvalMode)

	return &mode
}

func (client *Client) CreateResource(ctx context.Context, input *model.Resource) (*model.Resource, error) {
	opr := resourceResource.create()

	variables := newVars(
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlVar(input.Name, "name"),
		gqlVar(input.Address, "address"),
		gqlVar(newProtocolsInput(input.Protocols), "protocols"),
		gqlNullable(input.IsVisible, "isVisible"),
		gqlNullable(input.IsBrowserShortcutEnabled, "isBrowserShortcutEnabled"),
		gqlNullable(input.Alias, "alias"),
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
		gqlVar(NewAccessApprovalMode(input.ApprovalMode), "approvalMode"),
		gqlVar(newTagInputs(input.Tags), "tags"),
		gqlNullable(input.UsageBasedAutolockDurationDays, "usageBasedAutolockDurationDays"),
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)

	response := query.CreateResource{}
	if err := client.mutate(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	resource := response.Entity.ToModel()
	resource.GroupsAccess = input.GroupsAccess
	resource.ServiceAccounts = input.ServiceAccounts
	resource.IsAuthoritative = input.IsAuthoritative

	if input.IsVisible == nil {
		resource.IsVisible = nil
	}

	if input.IsBrowserShortcutEnabled == nil {
		resource.IsBrowserShortcutEnabled = nil
	}

	if input.SecurityPolicyID != nil && *input.SecurityPolicyID == "" {
		resource.SecurityPolicyID = input.SecurityPolicyID
	}

	return resource, nil
}

func (client *Client) ReadResource(ctx context.Context, resourceID string) (*model.Resource, error) {
	opr := resourceResource.read()

	if resourceID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	if res, ok := getResource[*model.Resource](resourceID); ok {
		return res, nil
	}

	variables := newVars(
		gqlID(resourceID),
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)

	response := query.ReadResource{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	if err := response.Resource.Access.FetchPages(withOperationCtx(ctx, opr), client.readResourceAccessAfter, newVars(gqlID(resourceID))); err != nil {
		return nil, err //nolint
	}

	res := response.Resource.ToModel()

	setResource(res)

	return res, nil
}

func (client *Client) readResourceAccessAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
	opr := resourceResource.read().withCustomName("readResourceAccessAfter")

	resourceID := string(variables["id"].(graphql.ID))
	variables[query.CursorAccess] = cursor
	pageLimit(client.pageLimit)(variables)

	response := query.ReadResourceAccess{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return &response.Resource.Access.PaginatedResource, nil
}

func (client *Client) ReadResources(ctx context.Context) ([]*model.Resource, error) {
	opr := resourceResource.read().withCustomName("readResources")

	variables := newVars(
		cursor(query.CursorResources),
		pageLimit(client.pageLimit),
	)

	response := query.ReadResources{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil && !errors.Is(err, ErrGraphqlResultIsEmpty) {
		return nil, err
	}

	if err := response.FetchPages(withOperationCtx(ctx, opr), client.readResourcesAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readResourcesAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.ResourceEdge], error) {
	opr := resourceResource.read().withCustomName("readResourcesAfter")

	variables[query.CursorResources] = cursor

	response := query.ReadResources{}
	if err := client.query(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) ReadFullResources(ctx context.Context) ([]*model.Resource, error) {
	opr := resourceResource.read().withCustomName("readFullResources")

	variables := newVars(
		cursor(query.CursorAccess),
		cursor(query.CursorResources),
		pageLimit(extendedPageLimit),
	)

	response := query.ReadFullResources{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil && !errors.Is(err, ErrGraphqlResultIsEmpty) {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.FetchPages(oprCtx, client.readFullResourcesAfter, variables); err != nil {
		return nil, err //nolint
	}

	for i := range response.Edges {
		if err := response.Edges[i].Node.Access.FetchPages(oprCtx, client.readExtendedResourceAccessAfter, newVars(gqlID(response.Edges[i].Node.ID))); err != nil {
			return nil, err //nolint:wrapcheck
		}
	}

	return response.ToModel(), nil
}

func (client *Client) readFullResourcesAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.FullResourceEdge], error) {
	opr := resourceResource.read().withCustomName("readFullResourcesAfter")

	variables[query.CursorResources] = cursor

	response := query.ReadFullResources{}
	if err := client.query(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) readExtendedResourceAccessAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
	opr := resourceResource.read().withCustomName("readExtendedResourceAccessAfter")

	resourceID := string(variables["id"].(graphql.ID))
	variables[query.CursorAccess] = cursor
	pageLimit(extendedPageLimit)(variables)

	response := query.ReadResourceAccess{}
	if err := client.query(ctx, &response, variables, opr, attr{id: resourceID}); err != nil {
		return nil, err
	}

	return &response.Resource.Access.PaginatedResource, nil
}

func (client *Client) UpdateResource(ctx context.Context, input *model.Resource) (*model.Resource, error) {
	opr := resourceResource.update()

	invalidateResource[*model.Resource](input.ID)

	variables := newVars(
		gqlID(input.ID),
		gqlID(input.RemoteNetworkID, "remoteNetworkId"),
		gqlVar(input.Name, "name"),
		gqlVar(input.Address, "address"),
		gqlVar(newProtocolsInput(input.Protocols), "protocols"),
		gqlVar(input.IsActive, "isActive"),
		gqlNullable(input.IsVisible, "isVisible"),
		gqlNullable(input.IsBrowserShortcutEnabled, "isBrowserShortcutEnabled"),
		gqlNullable(input.Alias, "alias"),
		gqlNullableID(input.SecurityPolicyID, "securityPolicyId"),
		gqlVar(NewAccessApprovalMode(input.ApprovalMode), "approvalMode"),
		gqlVar(newTagInputs(input.Tags), "tags"),
		gqlNullable(input.UsageBasedAutolockDurationDays, "usageBasedAutolockDurationDays"),
		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)

	response := query.UpdateResource{}
	if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
		return nil, err
	}

	if err := response.Entity.Access.FetchPages(withOperationCtx(ctx, opr), client.readResourceAccessAfter, newVars(gqlID(input.ID))); err != nil {
		return nil, err //nolint
	}

	resource := response.Entity.ToModel()
	resource.IsAuthoritative = input.IsAuthoritative

	if input.IsVisible == nil {
		resource.IsVisible = nil
	}

	if input.IsBrowserShortcutEnabled == nil {
		resource.IsBrowserShortcutEnabled = nil
	}

	if input.SecurityPolicyID == nil {
		resource.SecurityPolicyID = nil
	}

	setResource(resource)

	return resource, nil
}

func (client *Client) DeleteResource(ctx context.Context, resourceID string) error {
	opr := resourceResource.delete()

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	invalidateResource[*model.Resource](resourceID)

	response := query.DeleteResource{}

	return client.mutate(ctx, &response, newVars(gqlID(resourceID)), opr, attr{id: resourceID})
}

func (client *Client) UpdateResourceActiveState(ctx context.Context, resource *model.Resource) error {
	opr := resourceResource.update()

	invalidateResource[*model.Resource](resource.ID)

	variables := newVars(
		gqlID(resource.ID),
		gqlVar(resource.IsActive, "isActive"),
	)

	response := query.UpdateResourceActiveState{}

	return client.mutate(ctx, &response, variables, opr, attr{id: resource.ID})
}

func (client *Client) ReadResourcesByName(ctx context.Context, name, filter string, tags map[string]string) ([]*model.Resource, error) {
	opr := resourceResource.read().withCustomName("readResourcesByName")

	variables := newVars(
		gqlNullable(query.NewResourceFilterInput(name, filter, tags), "filter"),
		cursor(query.CursorResources),
		pageLimit(client.pageLimit),
	)

	response := query.ReadResourcesByName{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	if err := response.FetchPages(withOperationCtx(ctx, opr), client.readResourcesByNameAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) readResourcesByNameAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.ResourceEdge], error) {
	opr := resourceResource.read().withCustomName("readResourcesByName")

	variables[query.CursorResources] = cursor

	response := query.ReadResourcesByName{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) RemoveResourceAccess(ctx context.Context, resourceID string, principalIDs []string) error {
	opr := resourceResourceAccess.delete()

	if len(principalIDs) == 0 {
		return nil
	}

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	invalidateResource[*model.Resource](resourceID)

	variables := newVars(
		gqlID(resourceID),
		gqlIDs(principalIDs, "principalIds"),
	)

	response := query.RemoveResourceAccess{}

	return client.mutate(ctx, &response, variables, opr, attr{id: resourceID})
}

type AccessInput struct {
	PrincipalID                    string              `json:"principalId"`
	SecurityPolicyID               *string             `json:"securityPolicyId"`
	UsageBasedAutolockDurationDays *int64              `json:"usageBasedAutolockDurationDays"`
	ApprovalMode                   *AccessApprovalMode `json:"approvalMode"`
}

func (client *Client) AddResourceAccess(ctx context.Context, resourceID string, access []AccessInput) error {
	opr := resourceResourceAccess.update()

	if len(access) == 0 {
		return nil
	}

	if resourceID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	invalidateResource[*model.Resource](resourceID)

	variables := newVars(
		gqlID(resourceID),
		gqlNullable(access, "access"),
	)

	response := query.AddResourceAccess{}

	return client.mutate(ctx, &response, variables, opr, attr{id: resourceID})
}
