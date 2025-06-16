package resource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStateUpgraderV1(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		priorState    func() resourceModelV1
		expectedState func() resourceModel
	}{
		{
			name: "base case",
			priorState: func() resourceModelV1 {
				return resourceModelV1{
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
					Access:                   makeObjectsListNull(ctx, accessBlockAttributeTypes()),
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
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "minimum case",
			priorState: func() resourceModelV1 {
				return resourceModelV1{
					ID:              types.StringValue("test-id"),
					Name:            types.StringValue("test-name"),
					Address:         types.StringValue("test-address"),
					RemoteNetworkID: types.StringValue("test-remote-network-id"),
					Protocols:       defaultProtocolsObject(),
					Access:          makeObjectsListNull(ctx, accessBlockAttributeTypes()),
				}
			},
			expectedState: func() resourceModel {
				return resourceModel{
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
					ApprovalMode:                   types.StringNull(),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "with empty alias and security policy ID",
			priorState: func() resourceModelV1 {
				return resourceModelV1{
					ID:               types.StringValue("test-id"),
					Name:             types.StringValue("test-name"),
					Address:          types.StringValue("test-address"),
					RemoteNetworkID:  types.StringValue("test-remote-network-id"),
					Protocols:        defaultProtocolsObject(),
					Access:           makeObjectsListNull(ctx, accessBlockAttributeTypes()),
					Alias:            types.StringValue(""),
					SecurityPolicyID: types.StringValue(""),
				}
			},
			expectedState: func() resourceModel {
				return resourceModel{
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
					ApprovalMode:                   types.StringNull(),
					GroupAccess:                    makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess:                  makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "with access block",
			priorState: func() resourceModelV1 {
				groupIDs := []string{"test-group-id-1", "test-group-id-2"}
				serviceAccountIDs := []string{"test-service-account-id-1", "test-service-account-id-2"}
				access, diags := convertAccessBlockToTerraform(ctx, groupIDs, serviceAccountIDs)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV1{
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
					Access:                   access,
				}
			},
			expectedState: func() resourceModel {
				groupIDs := []string{"test-group-id-1", "test-group-id-2"}
				serviceAccountIDs := []string{"test-service-account-id-1", "test-service-account-id-2"}
				accessGroup, diags := convertAccessGroupsToTerraform(ctx, groupIDs)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				accessServiceAccount, diags := convertAccessServiceAccountsToTerraform(ctx, serviceAccountIDs)
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
					GroupAccess:                    accessGroup,
					ServiceAccess:                  accessServiceAccount,
					Tags:                           types.MapNull(types.StringType),
					TagsAll:                        types.MapNull(types.StringType),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given: Prior state (resourceModelV1)
			state := tfsdk.State{
				Schema: upgradeResourceStateV1().PriorSchema,
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
			upgradeResourceStateV1().StateUpgrader(ctx, req, resp)

			// Then: Verify the upgraded state
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected errors during upgrade: %v", resp.Diagnostics)
			}

			// Validate the warning message
			assert.Len(t, resp.Diagnostics, 1)
			assert.Equal(t, "Please update the access blocks.", resp.Diagnostics[0].Summary())

			// Retrieve the upgraded state
			var upgradedState resourceModel
			digs := resp.State.Get(ctx, &upgradedState)
			assert.False(t, digs.HasError())

			assert.Equal(t, test.expectedState(), upgradedState)
		})
	}

}
