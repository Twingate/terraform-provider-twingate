package resource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var resourceStateUpgradeV1 = resource.StateUpgrader{ //nolint
	PriorSchema: &schema.Schema{
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed: true,
			},
			attr.Name: schema.StringAttribute{
				Required: true,
			},
			attr.Address: schema.StringAttribute{
				Required: true,
			},
			attr.RemoteNetworkID: schema.StringAttribute{
				Required: true,
			},
			attr.IsAuthoritative: schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			attr.Alias: schema.StringAttribute{
				Optional: true,
			},
			attr.SecurityPolicyID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			attr.IsVisible: schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			attr.IsBrowserShortcutEnabled: schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			attr.Protocols: schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					attr.AllowIcmp: schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
					attr.UDP: schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Attributes: map[string]schema.Attribute{
							attr.Policy: schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							attr.Ports: schema.SetAttribute{
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
						},
					},
					attr.TCP: schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Attributes: map[string]schema.Attribute{
							attr.Policy: schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							attr.Ports: schema.SetAttribute{
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
		},

		Blocks: map[string]schema.Block{
			attr.Access: schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						attr.GroupIDs: schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
						attr.ServiceAccountIDs: schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
		},
	},
	StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
		var priorState resourceModelV1

		resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)

		if resp.Diagnostics.HasError() {
			return
		}

		protocols, err := convertProtocolsV1(priorState.Protocols)
		if err != nil {
			resp.Diagnostics.AddError(
				"failed to convert protocols for prior state version 0",
				err.Error(),
			)

			return
		}

		protocolsState, diags := convertProtocolsToTerraform(protocols, nil)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		upgradedState := resourceModelV2{
			ID:              priorState.ID,
			Name:            priorState.Name,
			Address:         priorState.Address,
			RemoteNetworkID: priorState.RemoteNetworkID,
			Protocols:       protocolsState,
		}

		if !priorState.IsAuthoritative.IsNull() {
			upgradedState.IsAuthoritative = priorState.IsAuthoritative
		}

		if !priorState.IsVisible.IsNull() {
			upgradedState.IsVisible = priorState.IsVisible
		}

		if !priorState.IsBrowserShortcutEnabled.IsNull() {
			upgradedState.IsBrowserShortcutEnabled = priorState.IsBrowserShortcutEnabled
		}

		if !priorState.Alias.IsNull() && priorState.Alias.ValueString() != "" {
			upgradedState.Alias = priorState.Alias
		}

		if !priorState.SecurityPolicyID.IsNull() && priorState.SecurityPolicyID.ValueString() != "" {
			upgradedState.SecurityPolicyID = priorState.SecurityPolicyID
		}

		securityPolicyID := upgradedState.SecurityPolicyID.ValueString()
		if securityPolicyID == "" {
			securityPolicyID = DefaultSecurityPolicyID
		}

		accessBlock := convertAccessBlocksV1(priorState.Access)
		var accessV2 []*ResourceAccessV2
		if accessBlock != nil {
			if len(accessBlock.ServiceAccountIDs) > 0 {
				accessV2 = append(accessV2, &ResourceAccessV2{
					ServiceAccountIDs: accessBlock.ServiceAccountIDs,
				})
			}

			for _, group := range accessBlock.GroupIDs {
				groupID := group
				accessV2 = append(accessV2, &ResourceAccessV2{
					GroupID:          &groupID,
					SecurityPolicyID: &securityPolicyID,
				})
			}
		}

		access, diags := convertAccessBlockV2ToTerraform(ctx, accessV2)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		upgradedState.Access = access

		resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)
	},
}

func convertAccessBlockV2ToTerraform(ctx context.Context, resources []*ResourceAccessV2) (types.Set, diag.Diagnostics) {
	accessObjects := make([]types.Object, 0, len(resources))

	for _, access := range resources {
		obj, diags := creatAccessV2Obj(access)
		if diags.HasError() {
			return makeNullSet(ctx, accessAttributeTypes()), diags
		}

		accessObjects = append(accessObjects, obj)
	}

	if len(accessObjects) == 0 {
		return makeNullSet(ctx, accessAttributeTypes()), nil
	}

	return makeSetList(ctx, accessObjects)
}

func creatAccessV2Obj(access *ResourceAccessV2) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := map[string]tfattr.Value{
		attr.GroupID:           types.StringNull(),
		attr.SecurityPolicyID:  types.StringNull(),
		attr.ServiceAccountIDs: types.SetNull(types.StringType),
	}

	if access.GroupID != nil {
		attributes[attr.GroupID] = types.StringPointerValue(access.GroupID)
		attributes[attr.SecurityPolicyID] = types.StringPointerValue(access.SecurityPolicyID)
	} else {
		var serviceAccounts types.Set
		serviceAccounts, diags = makeSet(access.ServiceAccountIDs)
		attributes[attr.ServiceAccountIDs] = serviceAccounts
	}

	if diags.HasError() {
		return types.ObjectNull(accessAttributeTypes()), diags
	}

	return types.ObjectValue(accessAttributeTypes(), attributes)
}

func convertAccessBlocksV1(blocks types.List) *ResourceAccessV1 {
	if blocks.IsNull() || blocks.IsUnknown() || len(blocks.Elements()) == 0 {
		return nil
	}

	obj := blocks.Elements()[0]

	return convertAccessBlockV1(obj.(types.Object))
}

type ResourceAccessV1 struct {
	GroupIDs          []string
	ServiceAccountIDs []string
}

type ResourceAccessV2 struct {
	SecurityPolicyID  *string
	GroupID           *string
	ServiceAccountIDs []string
}

func convertAccessBlockV1(obj types.Object) *ResourceAccessV1 {
	var access ResourceAccessV1

	attributes := obj.Attributes()

	serviceAccounts := attributes[attr.ServiceAccountIDs]
	if serviceAccounts != nil && !serviceAccounts.IsNull() && !serviceAccounts.IsUnknown() {
		access.ServiceAccountIDs = convertIDs(serviceAccounts.(types.Set))
	}

	groupIDs := attributes[attr.GroupIDs]
	if groupIDs != nil && !groupIDs.IsNull() && !groupIDs.IsUnknown() {
		access.GroupIDs = convertIDs(groupIDs.(types.Set))
	}

	return &access
}

func convertProtocolsV1(obj types.Object) (*model.Protocols, error) {
	if obj.IsNull() || obj.IsUnknown() {
		return model.DefaultProtocols(), nil
	}

	udp, err := convertProtocolV1(obj.Attributes()[attr.UDP])
	if err != nil {
		return nil, err
	}

	tcp, err := convertProtocolV1(obj.Attributes()[attr.TCP])
	if err != nil {
		return nil, err
	}

	return &model.Protocols{
		AllowIcmp: obj.Attributes()[attr.AllowIcmp].(types.Bool).ValueBool(),
		UDP:       udp,
		TCP:       tcp,
	}, nil
}

func convertProtocolV1(protocol tfattr.Value) (*model.Protocol, error) {
	obj := convertProtocolObjV1(protocol)
	if obj.IsNull() {
		return nil, nil //nolint:nilnil
	}

	ports, err := decodePortsV1(obj)
	if err != nil {
		return nil, err
	}

	policy := obj.Attributes()[attr.Policy].(types.String).ValueString()
	if err := isValidPolicyV1(policy, ports); err != nil {
		return nil, err
	}

	if policy == model.PolicyDenyAll {
		policy = model.PolicyRestricted
	}

	return model.NewProtocol(policy, ports), nil
}

func convertProtocolObjV1(protocol tfattr.Value) types.Object {
	if protocol == nil || protocol.IsNull() {
		return types.ObjectNull(nil)
	}

	list, ok := protocol.(types.List)
	if !ok || list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return types.ObjectNull(nil)
	}

	obj := list.Elements()[0].(types.Object)
	if obj.IsNull() || obj.IsUnknown() {
		return types.ObjectNull(nil)
	}

	return obj
}

func decodePortsV1(obj types.Object) ([]*model.PortRange, error) {
	portsVal := obj.Attributes()[attr.Ports]
	if portsVal == nil || portsVal.IsNull() {
		return nil, nil
	}

	portsList, ok := portsVal.(types.Set)
	if !ok {
		return nil, nil
	}

	return convertPortsV1(portsList)
}

func convertPortsV1(list types.Set) ([]*model.PortRange, error) {
	items := list.Elements()

	var ports = make([]*model.PortRange, 0, len(items))

	for _, port := range items {
		portRange, err := model.NewPortRange(port.(types.String).ValueString())
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		ports = append(ports, portRange)
	}

	return ports, nil
}

func isValidPolicyV1(policy string, ports []*model.PortRange) error {
	switch policy {
	case model.PolicyAllowAll:
		if len(ports) > 0 {
			return ErrPortsWithPolicyAllowAll
		}

	case model.PolicyDenyAll:
		if len(ports) > 0 {
			return ErrPortsWithPolicyDenyAll
		}

	case model.PolicyRestricted:
		if len(ports) == 0 {
			return ErrPolicyRestrictedWithoutPorts
		}
	}

	return nil
}
