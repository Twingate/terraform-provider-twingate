package resource

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStateUpgraderV3(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		priorState      func() resourceModelV3
		expectedState   func() resourceModel
		expectedWarning bool
	}{
		{
			name: "bare case",
			priorState: func() resourceModelV3 {
				return resourceModelV3{
					ID:                             types.StringValue("test-id"),
					Name:                           types.StringValue("test-name"),
					Address:                        types.StringValue("test-address"),
					RemoteNetworkID:                types.StringValue("test-remote-network-id"),
					Protocols:                      defaultProtocolsObject(),
					IsAuthoritative:                types.BoolNull(),
					IsActive:                       types.BoolNull(),
					IsVisible:                      types.BoolNull(),
					IsBrowserShortcutEnabled:       types.BoolNull(),
					Alias:                          types.StringNull(),
					SecurityPolicyID:               types.StringNull(),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypesV3()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				return resourceModel{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                defaultProtocolsObject(),
					IsAuthoritative:          types.BoolNull(),
					IsActive:                 types.BoolNull(),
					IsVisible:                types.BoolNull(),
					IsBrowserShortcutEnabled: types.BoolNull(),
					Alias:                    types.StringNull(),
					SecurityPolicyID:         types.StringNull(),
					AccessPolicy:             makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:              makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:            makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                     types.MapNull(types.StringType),
					TagsAll:                  types.MapNull(types.StringType),
				}
			},
			expectedWarning: true,
		},

		{
			name: "minimum case",
			priorState: func() resourceModelV3 {
				return resourceModelV3{
					ID:                             types.StringValue("test-id"),
					Name:                           types.StringValue("test-name"),
					Address:                        types.StringValue("test-address"),
					RemoteNetworkID:                types.StringValue("test-remote-network-id"),
					Protocols:                      defaultProtocolsObject(),
					IsAuthoritative:                types.BoolValue(true),
					IsActive:                       types.BoolValue(true),
					IsVisible:                      types.BoolValue(false),
					IsBrowserShortcutEnabled:       types.BoolValue(false),
					Alias:                          types.StringValue("alias.com"),
					SecurityPolicyID:               types.StringValue("security-policy-id"),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypesV3()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				return resourceModel{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                defaultProtocolsObject(),
					IsAuthoritative:          types.BoolValue(true),
					IsActive:                 types.BoolValue(true),
					IsVisible:                types.BoolValue(false),
					IsBrowserShortcutEnabled: types.BoolValue(false),
					Alias:                    types.StringValue("alias.com"),
					SecurityPolicyID:         types.StringValue("security-policy-id"),
					AccessPolicy:             makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:              makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:            makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                     types.MapNull(types.StringType),
					TagsAll:                  types.MapNull(types.StringType),
				}
			},
			expectedWarning: true,
		},

		{
			name: "base case",
			priorState: func() resourceModelV3 {
				return resourceModelV3{
					ID:                             types.StringValue("test-id"),
					Name:                           types.StringValue("test-name"),
					Address:                        types.StringValue("test-address"),
					RemoteNetworkID:                types.StringValue("test-remote-network-id"),
					Protocols:                      defaultProtocolsObject(),
					IsAuthoritative:                types.BoolValue(true),
					IsActive:                       types.BoolValue(true),
					IsVisible:                      types.BoolValue(false),
					IsBrowserShortcutEnabled:       types.BoolValue(false),
					Alias:                          types.StringValue("alias.com"),
					SecurityPolicyID:               types.StringValue("security-policy-id"),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypesV3()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					ApprovalMode:                   types.StringValue(model.ApprovalModeAutomatic),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				mode := model.AccessPolicyModeManual
				approvalMode := model.ApprovalModeAutomatic

				accessPolicy, diags := convertAccessPolicyToTerraformForImport(context.TODO(), &model.AccessPolicy{
					Mode:         &mode,
					ApprovalMode: &approvalMode,
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModel{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                defaultProtocolsObject(),
					IsAuthoritative:          types.BoolValue(true),
					IsActive:                 types.BoolValue(true),
					IsVisible:                types.BoolValue(false),
					IsBrowserShortcutEnabled: types.BoolValue(false),
					Alias:                    types.StringValue("alias.com"),
					SecurityPolicyID:         types.StringValue("security-policy-id"),
					AccessPolicy:             accessPolicy,
					GroupAccess:              makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:            makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                     types.MapNull(types.StringType),
					TagsAll:                  types.MapNull(types.StringType),
				}
			},
			expectedWarning: true,
		},

		{
			name: "base case with usage_based_autolock_duration_days",
			priorState: func() resourceModelV3 {
				return resourceModelV3{
					ID:                             types.StringValue("test-id"),
					Name:                           types.StringValue("test-name"),
					Address:                        types.StringValue("test-address"),
					RemoteNetworkID:                types.StringValue("test-remote-network-id"),
					Protocols:                      defaultProtocolsObject(),
					IsAuthoritative:                types.BoolValue(true),
					IsActive:                       types.BoolValue(true),
					IsVisible:                      types.BoolValue(false),
					IsBrowserShortcutEnabled:       types.BoolValue(false),
					Alias:                          types.StringValue("alias.com"),
					SecurityPolicyID:               types.StringValue("security-policy-id"),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypesV3()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					ApprovalMode:                   types.StringValue(model.ApprovalModeManual),
					UsageBasedAutolockDurationDays: types.Int64Value(2),
				}
			},
			expectedState: func() resourceModel {
				mode := model.AccessPolicyModeAutoLock
				approvalMode := model.ApprovalModeManual
				duration := "48h"

				accessPolicy, diags := convertAccessPolicyToTerraformForImport(context.TODO(), &model.AccessPolicy{
					Mode:         &mode,
					ApprovalMode: &approvalMode,
					Duration:     &duration,
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModel{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                defaultProtocolsObject(),
					IsAuthoritative:          types.BoolValue(true),
					IsActive:                 types.BoolValue(true),
					IsVisible:                types.BoolValue(false),
					IsBrowserShortcutEnabled: types.BoolValue(false),
					Alias:                    types.StringValue("alias.com"),
					SecurityPolicyID:         types.StringValue("security-policy-id"),
					AccessPolicy:             accessPolicy,
					GroupAccess:              makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:            makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                     types.MapNull(types.StringType),
					TagsAll:                  types.MapNull(types.StringType),
				}
			},
			expectedWarning: true,
		},

		{
			name: "base case with group_access",
			priorState: func() resourceModelV3 {
				approvalMode := model.ApprovalModeManual

				accessGroup, diags := convertAccessGroupsToTerraformV3(ctx, []*legacyAccessGroupV2{
					{
						GroupID:      "test-group-id",
						ApprovalMode: &approvalMode,
					},
				})
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV3{
					ID:                             types.StringValue("test-id"),
					Name:                           types.StringValue("test-name"),
					Address:                        types.StringValue("test-address"),
					RemoteNetworkID:                types.StringValue("test-remote-network-id"),
					Protocols:                      defaultProtocolsObject(),
					IsAuthoritative:                types.BoolValue(true),
					IsActive:                       types.BoolValue(true),
					IsVisible:                      types.BoolValue(false),
					IsBrowserShortcutEnabled:       types.BoolValue(false),
					Alias:                          types.StringValue("alias.com"),
					SecurityPolicyID:               types.StringValue("security-policy-id"),
					ApprovalMode:                   types.StringValue(model.ApprovalModeAutomatic),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    accessGroup,
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				mode := model.AccessPolicyModeManual
				automaticApprovalMode := model.ApprovalModeAutomatic
				manualApprovalMode := model.ApprovalModeManual

				accessPolicy, diags := convertAccessPolicyToTerraformForImport(context.TODO(), &model.AccessPolicy{
					Mode:         &mode,
					ApprovalMode: &automaticApprovalMode,
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				groupAccess, diags := convertGroupsAccessToTerraformForImport(context.TODO(), []model.AccessGroup{
					{
						GroupID: "test-group-id",
						AccessPolicy: &model.AccessPolicy{
							Mode:         &mode,
							ApprovalMode: &manualApprovalMode,
						},
					},
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModel{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                defaultProtocolsObject(),
					IsAuthoritative:          types.BoolValue(true),
					IsActive:                 types.BoolValue(true),
					IsVisible:                types.BoolValue(false),
					IsBrowserShortcutEnabled: types.BoolValue(false),
					Alias:                    types.StringValue("alias.com"),
					SecurityPolicyID:         types.StringValue("security-policy-id"),
					AccessPolicy:             accessPolicy,
					GroupAccess:              groupAccess,
					ServiceAccess:            makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                     types.MapNull(types.StringType),
					TagsAll:                  types.MapNull(types.StringType),
				}
			},
			expectedWarning: true,
		},

		{
			name: "base case with group_access and usage_based_autolock_duration_days",
			priorState: func() resourceModelV3 {
				approvalMode := model.ApprovalModeAutomatic
				usageBaseDuration := int64(3)

				accessGroup, diags := convertAccessGroupsToTerraformV3(ctx, []*legacyAccessGroupV2{
					{
						GroupID:            "test-group-id",
						ApprovalMode:       &approvalMode,
						UsageBasedDuration: &usageBaseDuration,
					},
				})
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV3{
					ID:                             types.StringValue("test-id"),
					Name:                           types.StringValue("test-name"),
					Address:                        types.StringValue("test-address"),
					RemoteNetworkID:                types.StringValue("test-remote-network-id"),
					Protocols:                      defaultProtocolsObject(),
					IsAuthoritative:                types.BoolValue(true),
					IsActive:                       types.BoolValue(true),
					IsVisible:                      types.BoolValue(false),
					IsBrowserShortcutEnabled:       types.BoolValue(false),
					Alias:                          types.StringValue("alias.com"),
					SecurityPolicyID:               types.StringValue("security-policy-id"),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    accessGroup,
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				mode := model.AccessPolicyModeAutoLock
				approvalMode := model.ApprovalModeAutomatic
				duration := "72h"

				groupAccess, diags := convertGroupsAccessToTerraformForImport(context.TODO(), []model.AccessGroup{
					{
						GroupID: "test-group-id",
						AccessPolicy: &model.AccessPolicy{
							Mode:         &mode,
							ApprovalMode: &approvalMode,
							Duration:     &duration,
						},
					},
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModel{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                defaultProtocolsObject(),
					IsAuthoritative:          types.BoolValue(true),
					IsActive:                 types.BoolValue(true),
					IsVisible:                types.BoolValue(false),
					IsBrowserShortcutEnabled: types.BoolValue(false),
					Alias:                    types.StringValue("alias.com"),
					SecurityPolicyID:         types.StringValue("security-policy-id"),
					AccessPolicy:             makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:              groupAccess,
					ServiceAccess:            makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                     types.MapNull(types.StringType),
					TagsAll:                  types.MapNull(types.StringType),
				}
			},
			expectedWarning: true,
		},

		{
			name: "base case with access_policy and group_access with access_policy",
			priorState: func() resourceModelV3 {
				mode := model.AccessPolicyModeAutoLock
				approvalMode := model.ApprovalModeAutomatic
				duration := "2d"

				accessPolicy := &model.AccessPolicy{
					Mode:         &mode,
					ApprovalMode: &approvalMode,
					Duration:     &duration,
				}

				accessPolicySet, diags := convertAccessPolicyToTerraformForImport(ctx, accessPolicy)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				accessGroup, diags := convertAccessGroupsToTerraformV3(ctx, []*legacyAccessGroupV2{
					{
						GroupID:      "test-group-id",
						AccessPolicy: accessPolicy,
					},
				})
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV3{
					ID:                             types.StringValue("test-id"),
					Name:                           types.StringValue("test-name"),
					Address:                        types.StringValue("test-address"),
					RemoteNetworkID:                types.StringValue("test-remote-network-id"),
					Protocols:                      defaultProtocolsObject(),
					IsAuthoritative:                types.BoolValue(true),
					IsActive:                       types.BoolValue(true),
					IsVisible:                      types.BoolValue(false),
					IsBrowserShortcutEnabled:       types.BoolValue(false),
					Alias:                          types.StringValue("alias.com"),
					SecurityPolicyID:               types.StringValue("security-policy-id"),
					AccessPolicy:                   accessPolicySet,
					GroupAccess:                    accessGroup,
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				mode := model.AccessPolicyModeAutoLock
				approvalMode := model.ApprovalModeAutomatic
				duration := "2d"

				accessPolicy := &model.AccessPolicy{
					Mode:         &mode,
					ApprovalMode: &approvalMode,
					Duration:     &duration,
				}

				accessPolicySet, diags := convertAccessPolicyToTerraformForImport(ctx, accessPolicy)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				groupAccess, diags := convertGroupsAccessToTerraformForImport(context.TODO(), []model.AccessGroup{
					{
						GroupID:      "test-group-id",
						AccessPolicy: accessPolicy,
					},
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModel{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                defaultProtocolsObject(),
					IsAuthoritative:          types.BoolValue(true),
					IsActive:                 types.BoolValue(true),
					IsVisible:                types.BoolValue(false),
					IsBrowserShortcutEnabled: types.BoolValue(false),
					Alias:                    types.StringValue("alias.com"),
					SecurityPolicyID:         types.StringValue("security-policy-id"),
					AccessPolicy:             accessPolicySet,
					GroupAccess:              groupAccess,
					ServiceAccess:            makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                     types.MapNull(types.StringType),
					TagsAll:                  types.MapNull(types.StringType),
				}
			},
			expectedWarning: false,
		},

		{
			name: "case with access_policy only in group_access",
			priorState: func() resourceModelV3 {
				mode := model.AccessPolicyModeAutoLock
				approvalMode := model.ApprovalModeAutomatic
				duration := "2d"

				accessPolicy := &model.AccessPolicy{
					Mode:         &mode,
					ApprovalMode: &approvalMode,
					Duration:     &duration,
				}

				accessGroup, diags := convertAccessGroupsToTerraformV3(ctx, []*legacyAccessGroupV2{
					{
						GroupID:      "test-group-id",
						AccessPolicy: accessPolicy,
					},
				})
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV3{
					ID:                             types.StringValue("test-id"),
					Name:                           types.StringValue("test-name"),
					Address:                        types.StringValue("test-address"),
					RemoteNetworkID:                types.StringValue("test-remote-network-id"),
					Protocols:                      defaultProtocolsObject(),
					IsAuthoritative:                types.BoolValue(true),
					IsActive:                       types.BoolValue(true),
					IsVisible:                      types.BoolValue(false),
					IsBrowserShortcutEnabled:       types.BoolValue(false),
					Alias:                          types.StringValue("alias.com"),
					SecurityPolicyID:               types.StringValue("security-policy-id"),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    accessGroup,
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				mode := model.AccessPolicyModeAutoLock
				approvalMode := model.ApprovalModeAutomatic
				duration := "2d"

				groupAccess, diags := convertGroupsAccessToTerraformForImport(context.TODO(), []model.AccessGroup{
					{
						GroupID: "test-group-id",
						AccessPolicy: &model.AccessPolicy{
							Mode:         &mode,
							ApprovalMode: &approvalMode,
							Duration:     &duration,
						},
					},
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModel{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                defaultProtocolsObject(),
					IsAuthoritative:          types.BoolValue(true),
					IsActive:                 types.BoolValue(true),
					IsVisible:                types.BoolValue(false),
					IsBrowserShortcutEnabled: types.BoolValue(false),
					Alias:                    types.StringValue("alias.com"),
					SecurityPolicyID:         types.StringValue("security-policy-id"),
					AccessPolicy:             makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:              groupAccess,
					ServiceAccess:            makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                     types.MapNull(types.StringType),
					TagsAll:                  types.MapNull(types.StringType),
				}
			},
			expectedWarning: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given: Prior state (resourceModelV3)
			state := tfsdk.State{
				Schema: upgradeResourceStateV3().PriorSchema,
			}

			diags := state.Set(ctx, test.priorState())
			if diags.HasError() {
				t.Fatalf("unexpected errors during upgrade: %v", diags)
			}

			// Mock the request and response
			req := resource.UpgradeStateRequest{
				State: &state,
			}

			newResource := NewResourceResource()
			newSchema := resource.SchemaResponse{}
			newResource.Schema(nil, resource.SchemaRequest{}, &newSchema)

			resp := &resource.UpgradeStateResponse{
				State: tfsdk.State{
					Schema: newSchema.Schema,
				},
			}

			// Call the StateUpgrader function under test
			upgradeResourceStateV3().StateUpgrader(ctx, req, resp)

			// Then: Verify the upgraded state
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected errors during upgrade: %v", resp.Diagnostics)
			}

			// Validate the warning message
			if test.expectedWarning {
				assert.Len(t, resp.Diagnostics, 1)
				assert.Equal(t, "Please use new access_policy block instead of approval_mode and usage_based_autolock_duration_days attributes.", resp.Diagnostics[0].Summary())
			} else {
				assert.Len(t, resp.Diagnostics, 0)
			}

			// Retrieve the upgraded state
			var upgradedState resourceModel
			digs := resp.State.Get(ctx, &upgradedState)
			assert.False(t, digs.HasError())

			assert.Equal(t, test.expectedState(), upgradedState)
		})
	}

}

func accessGroupAttributeTypesV3() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.GroupID:                        types.StringType,
		attr.SecurityPolicyID:               types.StringType,
		attr.UsageBasedAutolockDurationDays: types.Int64Type,
		attr.ApprovalMode:                   types.StringType,
		attr.AccessPolicy: types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: accessPolicyAttributeTypes(),
			},
		},
	}
}

func convertAccessGroupsToTerraformV3(ctx context.Context, groups []*legacyAccessGroupV2) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(groups) == 0 {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypesV3()), diagnostics
	}

	objects := make([]types.Object, 0, len(groups))

	for _, g := range groups {
		accessPolicy, diags := convertAccessPolicyToTerraformForImport(ctx, g.AccessPolicy)
		diagnostics.Append(diags...)

		attributes := map[string]tfattr.Value{
			attr.GroupID:                        types.StringValue(g.GroupID),
			attr.SecurityPolicyID:               types.StringPointerValue(g.SecurityPolicyID),
			attr.UsageBasedAutolockDurationDays: types.Int64PointerValue(g.UsageBasedDuration),
			attr.ApprovalMode:                   types.StringPointerValue(g.ApprovalMode),
			attr.AccessPolicy:                   accessPolicy,
		}

		obj, diags := types.ObjectValue(accessGroupAttributeTypesV3(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypesV3()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}
