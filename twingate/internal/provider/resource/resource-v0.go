package resource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceModelV0 struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Address                  types.String `tfsdk:"address"`
	RemoteNetworkID          types.String `tfsdk:"remote_network_id"`
	IsAuthoritative          types.Bool   `tfsdk:"is_authoritative"`
	Protocols                types.List   `tfsdk:"protocols"`
	Access                   types.List   `tfsdk:"access"`
	IsActive                 types.Bool   `tfsdk:"is_active"`
	IsVisible                types.Bool   `tfsdk:"is_visible"`
	IsBrowserShortcutEnabled types.Bool   `tfsdk:"is_browser_shortcut_enabled"`
	Alias                    types.String `tfsdk:"alias"`
	SecurityPolicyID         types.String `tfsdk:"security_policy_id"`
}

//nolint:funlen,cyclop
func upgradeResourceStateV0() resource.StateUpgrader {
	return resource.StateUpgrader{
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
				attr.IsActive: schema.BoolAttribute{
					Optional: true,
					Computed: true,
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
				attr.Protocols: schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							attr.AllowIcmp: schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
						},
						Blocks: map[string]schema.Block{
							attr.UDP: schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
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
							attr.TCP: schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
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
				},
			},
		},

		StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
			var priorState resourceModelV0

			resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)

			if resp.Diagnostics.HasError() {
				return
			}

			protocols, err := convertProtocolsV0(priorState.Protocols)
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

			upgradedState := resourceModelV1{
				ID:              priorState.ID,
				Name:            priorState.Name,
				Address:         priorState.Address,
				RemoteNetworkID: priorState.RemoteNetworkID,
				Protocols:       protocolsState,
				Access:          priorState.Access,
				IsActive:        priorState.IsActive,
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

			resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)

			resp.Diagnostics.AddWarning("Please update the protocols sections format from a block to an object",
				"See the v1 to v2 migration guide in the Twingate Terraform Provider documentation https://registry.terraform.io/providers/Twingate/twingate/latest/docs/guides/migration-v1-to-v2-guide")
		},
	}
}

func convertProtocolsV0(protocols types.List) (*model.Protocols, error) {
	if protocols.IsNull() || protocols.IsUnknown() || len(protocols.Elements()) == 0 {
		return model.DefaultProtocols(), nil
	}

	obj := protocols.Elements()[0].(types.Object)
	if obj.IsNull() || obj.IsUnknown() {
		return model.DefaultProtocols(), nil
	}

	udp, err := convertProtocolV0(obj.Attributes()[attr.UDP])
	if err != nil {
		return nil, err
	}

	tcp, err := convertProtocolV0(obj.Attributes()[attr.TCP])
	if err != nil {
		return nil, err
	}

	return &model.Protocols{
		AllowIcmp: obj.Attributes()[attr.AllowIcmp].(types.Bool).ValueBool(),
		UDP:       udp,
		TCP:       tcp,
	}, nil
}

func convertProtocolV0(protocol tfattr.Value) (*model.Protocol, error) {
	obj := convertProtocolObjV0(protocol)
	if obj.IsNull() {
		return nil, nil //nolint:nilnil
	}

	ports, err := decodePortsV0(obj)
	if err != nil {
		return nil, err
	}

	policy := obj.Attributes()[attr.Policy].(types.String).ValueString()
	if err := isValidPolicyV0(policy, ports); err != nil {
		return nil, err
	}

	if policy == model.PolicyDenyAll {
		policy = model.PolicyRestricted
	}

	return model.NewProtocol(policy, ports), nil
}

func convertProtocolObjV0(protocol tfattr.Value) types.Object {
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

func decodePortsV0(obj types.Object) ([]*model.PortRange, error) {
	portsVal := obj.Attributes()[attr.Ports]
	if portsVal == nil || portsVal.IsNull() {
		return nil, nil
	}

	portsList, ok := portsVal.(types.Set)
	if !ok {
		return nil, nil
	}

	return convertPortsV0(portsList)
}

func convertPortsV0(list types.Set) ([]*model.PortRange, error) {
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

func isValidPolicyV0(policy string, ports []*model.PortRange) error {
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
