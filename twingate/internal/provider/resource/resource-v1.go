package resource

import (
	"context"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

type resourceModelV1 struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Address                  types.String `tfsdk:"address"`
	RemoteNetworkID          types.String `tfsdk:"remote_network_id"`
	IsAuthoritative          types.Bool   `tfsdk:"is_authoritative"`
	Protocols                types.Object `tfsdk:"protocols"`
	Access                   types.List   `tfsdk:"access"`
	IsActive                 types.Bool   `tfsdk:"is_active"`
	IsVisible                types.Bool   `tfsdk:"is_visible"`
	IsBrowserShortcutEnabled types.Bool   `tfsdk:"is_browser_shortcut_enabled"`
	Alias                    types.String `tfsdk:"alias"`
	SecurityPolicyID         types.String `tfsdk:"security_policy_id"`
}

//nolint:funlen
func upgradeResourceStateV1() resource.StateUpgrader {
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

			groupIDs := getAccessAttribute(priorState.Access, attr.GroupIDs)
			serviceAccountIDs := getAccessAttribute(priorState.Access, attr.ServiceAccountIDs)

			accessGroup, diags := convertAccessGroupsToTerraform(ctx, groupIDs)
			resp.Diagnostics.Append(diags...)

			accessServiceAccount, diags := convertAccessServiceAccountsToTerraform(ctx, serviceAccountIDs)
			resp.Diagnostics.Append(diags...)

			upgradedState := resourceModel{
				ID:                             priorState.ID,
				Name:                           priorState.Name,
				Address:                        priorState.Address,
				RemoteNetworkID:                priorState.RemoteNetworkID,
				Protocols:                      priorState.Protocols,
				GroupAccess:                    accessGroup,
				ServiceAccess:                  accessServiceAccount,
				IsActive:                       priorState.IsActive,
				Tags:                           types.MapNull(types.StringType),
				TagsAll:                        types.MapNull(types.StringType),
				ApprovalMode:                   types.StringNull(),
				UsageBasedAutolockDurationDays: types.Int64Null(),
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

			resp.Diagnostics.AddWarning("Please update the access blocks.",
				"See the v2 to v3 migration guide in the Twingate Terraform Provider documentation https://registry.terraform.io/providers/Twingate/twingate/latest/docs/guides/migration-v2-to-v3-guide")
		},
	}
}

func convertAccessGroupsToTerraform(ctx context.Context, groups []string) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(groups) == 0 {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	objects := make([]types.Object, 0, len(groups))

	for _, g := range groups {
		attributes := map[string]tfattr.Value{
			attr.GroupID:                        types.StringValue(g),
			attr.SecurityPolicyID:               types.StringNull(),
			attr.UsageBasedAutolockDurationDays: types.Int64Null(),
			attr.ApprovalMode:                   types.StringNull(),
		}

		obj, diags := types.ObjectValue(accessGroupAttributeTypes(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}

func convertAccessServiceAccountsToTerraform(ctx context.Context, serviceAccounts []string) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(serviceAccounts) == 0 {
		return makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()), diagnostics
	}

	objects := make([]types.Object, 0, len(serviceAccounts))

	for _, account := range serviceAccounts {
		attributes := map[string]tfattr.Value{
			attr.ServiceAccountID: types.StringValue(account),
		}

		obj, diags := types.ObjectValue(accessServiceAccountAttributeTypes(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}

func accessBlockAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.GroupIDs:          types.SetType{ElemType: types.StringType},
		attr.ServiceAccountIDs: types.SetType{ElemType: types.StringType},
	}
}

func convertAccessBlockToTerraform(ctx context.Context, groups, serviceAccounts []string) (types.List, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(groups)+len(serviceAccounts) == 0 {
		return makeObjectsListNull(ctx, accessBlockAttributeTypes()), diagnostics
	}

	serviceAccountsSet, diags := makeStringsSet(serviceAccounts)
	diagnostics.Append(diags...)

	groupsSet, diags := makeStringsSet(groups)
	diagnostics.Append(diags...)

	attributes := map[string]tfattr.Value{
		attr.ServiceAccountIDs: serviceAccountsSet,
		attr.GroupIDs:          groupsSet,
	}

	obj, diags := types.ObjectValue(accessBlockAttributeTypes(), attributes)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeObjectsListNull(ctx, accessServiceAccountAttributeTypes()), diagnostics
	}

	return makeObjectsList(ctx, obj)
}
