package resource

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStateUpgraderV2(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		priorState    func() resourceModelV2
		expectedState func() resourceModel
	}{
		{
			name: "minimum case",
			priorState: func() resourceModelV2 {
				return resourceModelV2{
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
					ApprovalMode:                   types.StringNull(),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypesV2()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				return resourceModel{
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
					ApprovalMode:                   types.StringNull(),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "base case",
			priorState: func() resourceModelV2 {
				return resourceModelV2{
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
					ApprovalMode:                   types.StringValue(model.ApprovalModeManual),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypesV2()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				mode := model.AccessPolicyModeManual
				approvalMode := model.ApprovalModeManual

				accessPolicy, diags := convertAccessPolicyToTerraformForImport(context.TODO(), &model.AccessPolicy{
					Mode:         &mode,
					ApprovalMode: &approvalMode,
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModel{
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
					ApprovalMode:                   types.StringNull(),
					AccessPolicy:                   accessPolicy,
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "base case with usage_based_autolock_duration_days",
			priorState: func() resourceModelV2 {
				return resourceModelV2{
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
					ApprovalMode:                   types.StringValue(model.ApprovalModeManual),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypesV2()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
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
					ApprovalMode:                   types.StringNull(),
					AccessPolicy:                   accessPolicy,
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "base case with group_access",
			priorState: func() resourceModelV2 {
				approvalMode := model.ApprovalModeManual

				accessGroup, diags := convertAccessGroupsToTerraformV2(ctx, []*model.LegacyAccessGroup{
					{
						GroupID:      "test-group-id",
						ApprovalMode: &approvalMode,
					},
				})
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV2{
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
					ApprovalMode:                   types.StringValue(model.ApprovalModeManual),
					GroupAccess:                    accessGroup,
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
			expectedState: func() resourceModel {
				mode := model.AccessPolicyModeManual
				approvalMode := model.ApprovalModeManual

				accessPolicy, diags := convertAccessPolicyToTerraformForImport(context.TODO(), &model.AccessPolicy{
					Mode:         &mode,
					ApprovalMode: &approvalMode,
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				groupAccess, diags := convertGroupsAccessToTerraformForImport(context.TODO(), []model.AccessGroup{
					{
						GroupID: "test-group-id",
						AccessPolicy: &model.AccessPolicy{
							Mode:         &mode,
							ApprovalMode: &approvalMode,
						},
					},
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModel{
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
					ApprovalMode:                   types.StringNull(),
					AccessPolicy:                   accessPolicy,
					GroupAccess:                    groupAccess,
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "base case with group_access and usage_based_autolock_duration_days",
			priorState: func() resourceModelV2 {
				approvalMode := model.ApprovalModeAutomatic
				usageBaseDuration := int64(3)

				accessGroup, diags := convertAccessGroupsToTerraformV2(ctx, []*model.LegacyAccessGroup{
					{
						GroupID:            "test-group-id",
						ApprovalMode:       &approvalMode,
						UsageBasedDuration: &usageBaseDuration,
					},
				})
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV2{
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
					ApprovalMode:                   types.StringNull(),
					GroupAccess:                    accessGroup,
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
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
					ApprovalMode:                   types.StringNull(),
					AccessPolicy:                   makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:                    groupAccess,
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given: Prior state (resourceModelV2)
			state := tfsdk.State{
				Schema: upgradeResourceStateV2().PriorSchema,
			}
			state.Set(ctx, test.priorState())

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
			upgradeResourceStateV2().StateUpgrader(ctx, req, resp)

			// Then: Verify the upgraded state
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected errors during upgrade: %v", resp.Diagnostics)
			}

			// Validate the warning message
			assert.Len(t, resp.Diagnostics, 1)
			assert.Equal(t, "Please use new access_policy block instead of approval_mode and usage_based_autolock_duration_days attributes.", resp.Diagnostics[0].Summary())

			// Retrieve the upgraded state
			var upgradedState resourceModel
			digs := resp.State.Get(ctx, &upgradedState)
			assert.False(t, digs.HasError())

			assert.Equal(t, test.expectedState(), upgradedState)
		})
	}

}
