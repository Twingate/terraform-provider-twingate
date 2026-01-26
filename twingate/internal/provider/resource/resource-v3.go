package resource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/customplanmodifier"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/customvalidator"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceModelV3 struct {
	ID                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	Address                        types.String `tfsdk:"address"`
	RemoteNetworkID                types.String `tfsdk:"remote_network_id"`
	IsAuthoritative                types.Bool   `tfsdk:"is_authoritative"`
	Protocols                      types.Object `tfsdk:"protocols"`
	AccessPolicy                   types.Set    `tfsdk:"access_policy"`
	GroupAccess                    types.Set    `tfsdk:"access_group"`
	ServiceAccess                  types.Set    `tfsdk:"access_service"`
	IsActive                       types.Bool   `tfsdk:"is_active"`
	IsVisible                      types.Bool   `tfsdk:"is_visible"`
	IsBrowserShortcutEnabled       types.Bool   `tfsdk:"is_browser_shortcut_enabled"`
	Alias                          types.String `tfsdk:"alias"`
	SecurityPolicyID               types.String `tfsdk:"security_policy_id"`
	Tags                           types.Map    `tfsdk:"tags"`
	TagsAll                        types.Map    `tfsdk:"tags_all"`
	ApprovalMode                   types.String `tfsdk:"approval_mode"`                      // deprecated, kept for migration
	UsageBasedAutolockDurationDays types.Int64  `tfsdk:"usage_based_autolock_duration_days"` // deprecated, kept for migration
}

func accessPolicyBlockResourceStateV3() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Validators: []validator.Set{
			setvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attr.Mode: schema.StringAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.String{
						UseNullStringWhenValueOmitted(),
					},
					Validators: []validator.String{
						stringvalidator.OneOf(model.AccessPolicyModeManual, model.AccessPolicyModeAutoLock, model.AccessPolicyModeAccessRequest),
					},
				},

				attr.Duration: schema.StringAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.String{
						UseNullStringWhenValueOmitted(),
						customplanmodifier.Duration(),
					},
					Validators: []validator.String{
						customvalidator.Duration(),
					},
				},

				attr.ApprovalMode: schema.StringAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.String{
						UseNullStringWhenValueOmitted(),
					},
					Validators: []validator.String{
						stringvalidator.OneOf(model.ApprovalModeAutomatic, model.ApprovalModeManual),
					},
				},
			},
		},
	}
}

//nolint:funlen
func upgradeResourceStateV3() resource.StateUpgrader {
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
				attr.Protocols: schema.SingleNestedAttribute{
					Optional: true,
					Computed: true,
					Default:  objectdefault.StaticValue(defaultProtocolsObject()),
					Attributes: map[string]schema.Attribute{
						attr.AllowIcmp: schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						attr.UDP: schema.SingleNestedAttribute{
							Optional: true,
							Computed: true,
							Default:  objectdefault.StaticValue(defaultProtocolObject()),
							Attributes: map[string]schema.Attribute{
								attr.Policy: schema.StringAttribute{
									Optional: true,
									Computed: true,
									Validators: []validator.String{
										stringvalidator.OneOf(model.Policies...),
									},
									Default: stringdefault.StaticString(model.PolicyAllowAll),
								},
								attr.Ports: schema.SetAttribute{
									Optional:    true,
									Computed:    true,
									ElementType: types.StringType,
									PlanModifiers: []planmodifier.Set{
										PortsDiff(),
									},
									Default: setdefault.StaticValue(defaultEmptyPorts()),
								},
							},
						},
						attr.TCP: schema.SingleNestedAttribute{
							Optional: true,
							Computed: true,
							Default:  objectdefault.StaticValue(defaultProtocolObject()),
							Attributes: map[string]schema.Attribute{
								attr.Policy: schema.StringAttribute{
									Optional: true,
									Computed: true,
									Validators: []validator.String{
										stringvalidator.OneOf(model.Policies...),
									},
									Default: stringdefault.StaticString(model.PolicyAllowAll),
								},
								attr.Ports: schema.SetAttribute{
									Optional:    true,
									Computed:    true,
									ElementType: types.StringType,
									PlanModifiers: []planmodifier.Set{
										PortsDiff(),
									},
									Default: setdefault.StaticValue(defaultEmptyPorts()),
								},
							},
						},
					},
				},
				attr.Tags: schema.MapAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Computed:    true,
				},
				attr.TagsAll: schema.MapAttribute{
					ElementType: types.StringType,
					Computed:    true,
				},
				attr.ApprovalMode: schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				attr.UsageBasedAutolockDurationDays: schema.Int64Attribute{
					Optional: true,
					Computed: true,
				},
			},

			Blocks: map[string]schema.Block{
				attr.AccessGroup: schema.SetNestedBlock{
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							attr.GroupID: schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							attr.SecurityPolicyID: schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							attr.ApprovalMode: schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							attr.UsageBasedAutolockDurationDays: schema.Int64Attribute{
								Optional: true,
								Computed: true,
							},
						},
						Blocks: map[string]schema.Block{
							attr.AccessPolicy: accessPolicyBlockResourceStateV3(),
						},
					},
				},
				attr.AccessService: schema.SetNestedBlock{
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							attr.ServiceAccountID: schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
						},
					},
				},
				attr.AccessPolicy: accessPolicyBlockResourceStateV3(),
			},
		},

		StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
			var priorState resourceModelV3

			resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)

			if resp.Diagnostics.HasError() {
				return
			}

			accessGroup, diags := convertLegacyAccessGroupsToTerraformV3(ctx, priorState.GroupAccess)
			resp.Diagnostics.Append(diags...)

			var accessPolicy types.Set
			if !priorState.AccessPolicy.IsNull() && !priorState.AccessPolicy.IsUnknown() {
				accessPolicy = priorState.AccessPolicy
			} else {
				accessPolicy, diags = convertLegacyAccessPolicyToTerraform(ctx, priorState.ApprovalMode.ValueStringPointer(), priorState.UsageBasedAutolockDurationDays.ValueInt64Pointer())
				resp.Diagnostics.Append(diags...)
			}

			upgradedState := resourceModel{
				ID:                       priorState.ID,
				Name:                     priorState.Name,
				Address:                  priorState.Address,
				RemoteNetworkID:          priorState.RemoteNetworkID,
				Protocols:                priorState.Protocols,
				Alias:                    priorState.Alias,
				SecurityPolicyID:         priorState.SecurityPolicyID,
				AccessPolicy:             accessPolicy,
				GroupAccess:              accessGroup,
				ServiceAccess:            priorState.ServiceAccess,
				IsActive:                 priorState.IsActive,
				IsVisible:                priorState.IsVisible,
				IsAuthoritative:          priorState.IsAuthoritative,
				IsBrowserShortcutEnabled: priorState.IsBrowserShortcutEnabled,
				Tags:                     priorState.Tags,
				TagsAll:                  priorState.TagsAll,
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)

			resp.Diagnostics.AddWarning("Please use new access_policy block instead of approval_mode and usage_based_autolock_duration_days attributes.",
				"See the v3 to v4 migration guide in the Twingate Terraform Provider documentation https://registry.terraform.io/providers/Twingate/twingate/latest/docs/guides/migration-v3-to-v4-guide")
		},
	}
}

//nolint:funlen
func convertLegacyAccessGroupsToTerraformV3(ctx context.Context, groupAccess types.Set) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if groupAccess.IsNull() {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	groups, err := getLegacyGroupAccessAttributeV3(groupAccess)
	if err != nil {
		diagnostics.AddError("failed to convert access groups", err.Error())

		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	if len(groups) == 0 {
		// no legacy groups to convert - return a null set
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	for _, group := range groups {
		if group.ApprovalMode != nil || group.UsageBasedDuration != nil {
			mode := model.AccessPolicyModeManual

			accessPolicy := &model.AccessPolicy{
				Mode:         &mode,
				ApprovalMode: group.ApprovalMode,
			}

			if group.UsageBasedDuration != nil {
				duration := fmt.Sprintf("%dh", *group.UsageBasedDuration*hoursInDay)
				accessPolicy.Duration = &duration

				if *group.UsageBasedDuration >= 1 {
					mode = model.AccessPolicyModeAutoLock
					accessPolicy.Mode = &mode
				}
			}

			group.AccessPolicy = accessPolicy

			// Deprecated
			group.ApprovalMode = nil
			group.UsageBasedDuration = nil
		}
	}

	objects := make([]types.Object, 0, len(groups))

	for _, access := range groups {
		attributes := map[string]tfattr.Value{
			attr.GroupID:          types.StringValue(access.GroupID),
			attr.SecurityPolicyID: types.StringPointerValue(access.SecurityPolicyID),
		}

		accessPolicy, diags := convertAccessPolicyToTerraformForImport(ctx, access.AccessPolicy)
		diagnostics.Append(diags...)

		attributes[attr.AccessPolicy] = accessPolicy

		obj, diags := types.ObjectValue(accessGroupAttributeTypes(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}

// getLegacyGroupAccessAttributeV3 reads the access_group attributes from v3 state.
func getLegacyGroupAccessAttributeV3(list types.Set) ([]*legacyAccessGroupV2, error) {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil, nil
	}

	access := make([]*legacyAccessGroupV2, 0, len(list.Elements()))

	for _, item := range list.Elements() {
		obj := item.(types.Object)
		if obj.IsNull() || obj.IsUnknown() {
			continue
		}

		groupVal := obj.Attributes()[attr.GroupID]
		accessGroup := &legacyAccessGroupV2{
			GroupID: groupVal.(types.String).ValueString(),
		}

		securityPolicyVal := obj.Attributes()[attr.SecurityPolicyID]
		if securityPolicyVal != nil && !securityPolicyVal.IsNull() && !securityPolicyVal.IsUnknown() {
			accessGroup.SecurityPolicyID = securityPolicyVal.(types.String).ValueStringPointer()
		}

		var (
			err          error
			accessPolicy *model.AccessPolicy
		)

		accessPolicyVal := obj.Attributes()[attr.AccessPolicy]
		if accessPolicyVal != nil && !accessPolicyVal.IsNull() && !accessPolicyVal.IsUnknown() {
			accessPolicyRaw, ok := accessPolicyVal.(types.Set)
			if ok {
				accessPolicy, err = getAccessPolicyAttribute(accessPolicyRaw)
				if err != nil {
					return nil, fmt.Errorf("error parsing access_policy: %w", err)
				}
			}

			accessGroup.AccessPolicy = accessPolicy

		} else {
			usageBasedDuration := obj.Attributes()[attr.UsageBasedAutolockDurationDays]
			if usageBasedDuration != nil && !usageBasedDuration.IsNull() && !usageBasedDuration.IsUnknown() {
				accessGroup.UsageBasedDuration = usageBasedDuration.(types.Int64).ValueInt64Pointer()
			}

			approvalModeVal := obj.Attributes()[attr.ApprovalMode]
			if approvalModeVal != nil && !approvalModeVal.IsNull() && !approvalModeVal.IsUnknown() {
				accessGroup.ApprovalMode = approvalModeVal.(types.String).ValueStringPointer()
			}
		}

		access = append(access, accessGroup)
	}

	return access, nil
}
