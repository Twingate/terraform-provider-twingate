package resource

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

var defaultProtocols = &model.Protocols{
	AllowIcmp: true,
	TCP: &model.Protocol{
		Policy: model.PolicyAllowAll,
	},
	UDP: &model.Protocol{
		Policy: model.PolicyAllowAll,
	},
}

func TestStateUpgraderV0(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		priorState    func() resourceModelV0
		expectedState func() resourceModel
	}{
		{
			name: "base case",
			priorState: func() resourceModelV0 {
				protocolsV0, diags := convertProtocolsToTerraformV0(ctx, defaultProtocols)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV0{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                protocolsV0,
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

					AccessPolicy:  makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:   makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess: makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:          types.MapNull(types.StringType),
					TagsAll:       types.MapNull(types.StringType),

					// Deprecated
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "minimum case",
			priorState: func() resourceModelV0 {
				protocolsV0, diags := convertProtocolsToTerraformV0(ctx, defaultProtocols)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV0{
					ID:              types.StringValue("test-id"),
					Name:            types.StringValue("test-name"),
					Address:         types.StringValue("test-address"),
					RemoteNetworkID: types.StringValue("test-remote-network-id"),
					Protocols:       protocolsV0,
					Access:          makeObjectsListNull(ctx, accessBlockAttributeTypes()),
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

					AccessPolicy:  makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:   makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess: makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:          types.MapNull(types.StringType),
					TagsAll:       types.MapNull(types.StringType),

					// Deprecated
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "with empty alias and security policy ID",
			priorState: func() resourceModelV0 {
				protocolsV0, diags := convertProtocolsToTerraformV0(ctx, defaultProtocols)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV0{
					ID:               types.StringValue("test-id"),
					Name:             types.StringValue("test-name"),
					Address:          types.StringValue("test-address"),
					RemoteNetworkID:  types.StringValue("test-remote-network-id"),
					Protocols:        protocolsV0,
					Access:           makeObjectsListNull(ctx, accessBlockAttributeTypes()),
					Alias:            types.StringValue(""),
					SecurityPolicyID: types.StringValue(""),
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

					AccessPolicy:  makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					GroupAccess:   makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
					ServiceAccess: makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()),
					Tags:          types.MapNull(types.StringType),
					TagsAll:       types.MapNull(types.StringType),

					// Deprecated
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},

		{
			name: "with access block",
			priorState: func() resourceModelV0 {
				protocolsV0, diags := convertProtocolsToTerraformV0(ctx, defaultProtocols)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				groupIDs := []string{"test-group-id-1", "test-group-id-2"}
				serviceAccountIDs := []string{"test-service-account-id-1", "test-service-account-id-2"}
				access, diags := convertAccessBlockToTerraform(ctx, groupIDs, serviceAccountIDs)
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				return resourceModelV0{
					ID:                       types.StringValue("test-id"),
					Name:                     types.StringValue("test-name"),
					Address:                  types.StringValue("test-address"),
					RemoteNetworkID:          types.StringValue("test-remote-network-id"),
					Protocols:                protocolsV0,
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
				groupAccess, diags := convertGroupsAccessToTerraformForImport(context.TODO(), []model.AccessGroup{
					{GroupID: "test-group-id-1"}, {GroupID: "test-group-id-2"},
				})

				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags)
				}

				serviceAccess, diags := convertServiceAccessToTerraform(ctx, []string{"test-service-account-id-1", "test-service-account-id-2"})
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

					GroupAccess:   groupAccess,
					ServiceAccess: serviceAccess,

					AccessPolicy: makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
					Tags:         types.MapNull(types.StringType),
					TagsAll:      types.MapNull(types.StringType),

					// Deprecated
					ApprovalMode:                   types.StringNull(),
					UsageBasedAutolockDurationDays: types.Int64Null(),
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given: Prior state resourceModelV0
			state := tfsdk.State{
				Schema: upgradeResourceStateV0().PriorSchema,
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
			upgradeResourceStateV0().StateUpgrader(ctx, req, resp)

			// Then: Verify the upgraded state
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected errors during upgrade: %v", resp.Diagnostics)
			}

			// Validate the warning message
			assert.Len(t, resp.Diagnostics, 1)
			assert.Equal(t, "Please update the protocols sections format from a block to an object", resp.Diagnostics[0].Summary())

			// Retrieve the upgraded state
			var upgradedState resourceModel
			digs := resp.State.Get(ctx, &upgradedState)
			assert.False(t, digs.HasError())

			assert.Equal(t, test.expectedState(), upgradedState)
		})
	}

}

func convertProtocolsToTerraformV0(ctx context.Context, protocols *model.Protocols) (types.List, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	tcp, diags := convertProtocolModelToTerraformV0(ctx, protocols.TCP)
	diagnostics.Append(diags...)

	udp, diags := convertProtocolModelToTerraformV0(ctx, protocols.UDP)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeObjectsListNull(ctx, protocolsAttributeTypesV0()), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.AllowIcmp: types.BoolValue(protocols.AllowIcmp),
		attr.TCP:       tcp,
		attr.UDP:       udp,
	}

	obj := types.ObjectValueMust(protocolsAttributeTypesV0(), attributes)

	return makeObjectsList(ctx, obj)
}

func convertProtocolModelToTerraformV0(ctx context.Context, protocol *model.Protocol) (types.List, diag.Diagnostics) {
	if protocol == nil {
		return makeObjectsListNull(ctx, protocolAttributeTypesV0()), nil
	}

	ports := convertPortsToTerraform(protocol.Ports)

	policy := protocol.Policy
	if policy == model.PolicyRestricted && len(ports.Elements()) == 0 {
		policy = model.PolicyDenyAll
	}

	attributes := map[string]tfattr.Value{
		attr.Policy: types.StringValue(policy),
		attr.Ports:  ports,
	}

	obj, diags := types.ObjectValue(protocolAttributeTypesV0(), attributes)
	if diags.HasError() {
		return makeObjectsListNull(ctx, protocolAttributeTypesV0()), diags
	}

	return makeObjectsList(ctx, obj)
}

func protocolAttributeTypesV0() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.Policy: types.StringType,
		attr.Ports: types.SetType{
			ElemType: types.StringType,
		},
	}
}

func protocolsAttributeTypesV0() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.AllowIcmp: types.BoolType,
		attr.TCP: types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: protocolAttributeTypesV0(),
			},
		},
		attr.UDP: types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: protocolAttributeTypesV0(),
			},
		},
	}
}
