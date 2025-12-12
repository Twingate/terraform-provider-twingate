package resource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
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

const hoursInDay = 24

type resourceModelV2 struct {
	ID                             types.String `tfsdk:"id"`
	Name                           types.String `tfsdk:"name"`
	Address                        types.String `tfsdk:"address"`
	RemoteNetworkID                types.String `tfsdk:"remote_network_id"`
	IsAuthoritative                types.Bool   `tfsdk:"is_authoritative"`
	Protocols                      types.Object `tfsdk:"protocols"`
	GroupAccess                    types.Set    `tfsdk:"access_group"`
	ServiceAccess                  types.Set    `tfsdk:"access_service"`
	IsActive                       types.Bool   `tfsdk:"is_active"`
	IsVisible                      types.Bool   `tfsdk:"is_visible"`
	IsBrowserShortcutEnabled       types.Bool   `tfsdk:"is_browser_shortcut_enabled"`
	Alias                          types.String `tfsdk:"alias"`
	SecurityPolicyID               types.String `tfsdk:"security_policy_id"`
	ApprovalMode                   types.String `tfsdk:"approval_mode"`
	Tags                           types.Map    `tfsdk:"tags"`
	TagsAll                        types.Map    `tfsdk:"tags_all"`
	UsageBasedAutolockDurationDays types.Int64  `tfsdk:"usage_based_autolock_duration_days"`
}

//nolint:funlen
func upgradeResourceStateV2() resource.StateUpgrader {
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
			},
		},

		StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
			var priorState resourceModelV2

			resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)

			if resp.Diagnostics.HasError() {
				return
			}

			accessGroup, diags := convertLegacyAccessGroupsToTerraform(ctx, priorState.GroupAccess)
			resp.Diagnostics.Append(diags...)

			accessPolicy, diags := convertLegacyAccessPolicyToTerraform(ctx, priorState.ApprovalMode.ValueStringPointer(), priorState.UsageBasedAutolockDurationDays.ValueInt64Pointer())
			resp.Diagnostics.Append(diags...)

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

				// Deprecated
				ApprovalMode:                   types.StringNull(),
				UsageBasedAutolockDurationDays: types.Int64Null(),
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)

			resp.Diagnostics.AddWarning("Please use new access_policy block instead of approval_mode and usage_based_autolock_duration_days attributes.",
				"See the v3 to v4 migration guide in the Twingate Terraform Provider documentation https://registry.terraform.io/providers/Twingate/twingate/latest/docs/guides/migration-v3-to-v4-guide")
		},
	}
}

func convertLegacyAccessPolicyToTerraform(ctx context.Context, approvalMode *string, usageBasedAutolockDurationDays *int64) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	attributes := map[string]tfattr.Value{
		attr.Mode:         types.StringValue(model.AccessPolicyModeManual),
		attr.Duration:     types.StringNull(),
		attr.ApprovalMode: types.StringPointerValue(approvalMode),
	}

	if approvalMode == nil && usageBasedAutolockDurationDays == nil {
		return makeObjectsSetNull(ctx, accessPolicyAttributeTypes()), diagnostics
	}

	if usageBasedAutolockDurationDays != nil {
		duration := fmt.Sprintf("%dh", *usageBasedAutolockDurationDays*hoursInDay)
		attributes[attr.Duration] = types.StringValue(duration)

		if *usageBasedAutolockDurationDays >= 1 {
			attributes[attr.Mode] = types.StringValue(model.AccessPolicyModeAutoLock)
		}
	}

	obj, diags := types.ObjectValue(accessPolicyAttributeTypes(), attributes)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessPolicyAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, obj)
}

//nolint:funlen
func convertLegacyAccessGroupsToTerraform(ctx context.Context, groupAccess types.Set) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if groupAccess.IsNull() {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	groups, err := getLegacyGroupAccessAttribute(groupAccess)
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
			attr.GroupID:                        types.StringValue(access.GroupID),
			attr.SecurityPolicyID:               types.StringPointerValue(access.SecurityPolicyID),
			attr.UsageBasedAutolockDurationDays: types.Int64Null(),
			attr.ApprovalMode:                   types.StringNull(),
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
