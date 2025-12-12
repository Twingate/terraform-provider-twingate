package client

import (
	"context"
	"errors"
	"log"

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

type AccessMode string
type AccessPolicyInput struct {
	Mode            AccessMode `json:"mode"`
	DurationSeconds *int64     `json:"durationSeconds"`
}

func NewAccessPolicyInput(accessPolicy *model.AccessPolicy) *AccessPolicyInput {
	if accessPolicy == nil {
		return &AccessPolicyInput{
			Mode: model.AccessPolicyModeManual,
		}
	}

	var durationSeconds *int64

	if accessPolicy.Duration != nil {
		duration, _ := accessPolicy.ParseDuration()
		seconds := int64(duration.Seconds())
		durationSeconds = &seconds
	}

	mode := AccessMode(model.AccessPolicyModeManual)
	if durationSeconds != nil {
		mode = model.AccessPolicyModeAutoLock
	}

	if accessPolicy.Mode != nil {
		mode = AccessMode(*accessPolicy.Mode)
	}

	return &AccessPolicyInput{
		Mode:            mode,
		DurationSeconds: durationSeconds,
	}
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

func NewAccessApprovalMode(accessPolicy *model.AccessPolicy) *AccessApprovalMode {
	var approvalMode string

	if accessPolicy != nil && accessPolicy.ApprovalMode != nil {
		approvalMode = *accessPolicy.ApprovalMode
	}

	if approvalMode == "" {
		approvalMode = model.ApprovalModeManual
	}

	val := AccessApprovalMode(approvalMode)

	return &val
}

func NewGroupAccessApprovalMode(accessPolicy *model.AccessPolicy) *AccessApprovalMode {
	if accessPolicy == nil || accessPolicy.ApprovalMode == nil || *accessPolicy.ApprovalMode == "" {
		return nil
	}

	val := AccessApprovalMode(*accessPolicy.ApprovalMode)

	return &val
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
		gqlVar(NewAccessPolicyInput(input.AccessPolicy), "accessPolicy"),
		gqlVar(NewAccessApprovalMode(input.AccessPolicy), "approvalMode"),
		gqlVar(newTagInputs(input.Tags), "tags"),

		cursor(query.CursorAccess),
		pageLimit(client.pageLimit),
	)

	response := query.CreateResource{}
	if err := client.mutate(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	resource, err := response.Entity.ToModel()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

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

	res, err := response.Resource.ToModel()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	setResource(res)

	return res, nil
}

func (client *Client) readResourceAccessAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
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

func (client *Client) readResourcesAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.ResourceEdge], error) {
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

	return response.ToModel() //nolint:wrapcheck
}

func (client *Client) ReadFullResourcesByName(ctx context.Context, filter *model.ResourcesFilter) ([]*model.Resource, error) {
	opr := resourceResource.read().withCustomName("readFullResourcesByName")

	variables := newVars(
		gqlNullable(query.NewResourceFilterInput(filter.GetName(), filter.GetFilterBy(), filter.GetTags()), "filter"),
		cursor(query.CursorAccess),
		cursor(query.CursorResources),
		pageLimit(extendedPageLimit),
	)

	response := query.ReadFullResourcesByName{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil && !errors.Is(err, ErrGraphqlResultIsEmpty) {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.FetchPages(oprCtx, client.readFullResourcesByNameAfter, variables); err != nil {
		return nil, err //nolint
	}

	for i := range response.Edges {
		if err := response.Edges[i].Node.Access.FetchPages(oprCtx, client.readExtendedResourceAccessAfter, newVars(gqlID(response.Edges[i].Node.ID))); err != nil {
			return nil, err //nolint:wrapcheck
		}
	}

	return response.ToModel() //nolint:wrapcheck
}

func (client *Client) readFullResourcesByNameAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.FullResourceEdge], error) {
	opr := resourceResource.read().withCustomName("readFullResourcesByNameAfter")

	variables[query.CursorResources] = cursor

	response := query.ReadFullResourcesByName{}
	if err := client.query(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) readFullResourcesAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.FullResourceEdge], error) {
	opr := resourceResource.read().withCustomName("readFullResourcesAfter")

	variables[query.CursorResources] = cursor

	response := query.ReadFullResources{}
	if err := client.query(ctx, &response, variables, opr); err != nil {
		return nil, err
	}

	return &response.PaginatedResource, nil
}

func (client *Client) readExtendedResourceAccessAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.AccessEdge], error) {
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
		gqlVar(NewAccessApprovalMode(input.AccessPolicy), "approvalMode"),
		gqlVar(NewAccessPolicyInput(input.AccessPolicy), "accessPolicy"),
		gqlVar(newTagInputs(input.Tags), "tags"),
		// gqlNullable(input.UsageBasedAutolockDurationDays, "usageBasedAutolockDurationDays"),
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

	resource, err := response.Entity.ToModel()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

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

	// if input.AccessPolicy == nil {
	//	resource.AccessPolicy = nil
	//}

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

func (client *Client) ReadResourcesByName(ctx context.Context, filter *model.ResourcesFilter) ([]*model.Resource, error) {
	opr := resourceResource.read().withCustomName("readResourcesByName")

	// cache is not used when cache filter config set or cache disabled
	if isCacheReady[*model.Resource]() {
		if matched := matchResources[*model.Resource](filter); len(matched) > 0 {
			log.Printf(
				"[DEBUG] ReadResourcesByName: matched #%d resources from cache: %v",
				len(matched), utils.Map(matched, func(item *model.Resource) string {
					return item.Name
				}))

			return matched, nil
		}

		log.Println("[DEBUG] ReadResourcesByName: no matched resource in cache: fallback to query API")
	}

	variables := newVars(
		gqlNullable(query.NewResourceFilterInput(filter.GetName(), filter.GetFilterBy(), filter.GetTags()), "filter"),
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

func (client *Client) readResourcesByNameAfter(ctx context.Context, variables map[string]any, cursor string) (*query.PaginatedResource[*query.ResourceEdge], error) {
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
	PrincipalID      string              `json:"principalId"`
	SecurityPolicyID *string             `json:"securityPolicyId"`
	ApprovalMode     *AccessApprovalMode `json:"approvalMode"`
	AccessPolicy     *AccessPolicyInput  `json:"accessPolicy"`
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
