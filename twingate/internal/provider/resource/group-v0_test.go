package resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestGroupStateUpgraderV0(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		priorState    func() groupModelV0
		expectedState func() groupModel
	}{
		{
			name: "base case",
			priorState: func() groupModelV0 {
				userIDs, diags := types.SetValueFrom(ctx, types.StringType, []string{"test-user-id-1", "test-user-id-2"})
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags.Errors())
				}

				return groupModelV0{
					ID:               types.StringValue("test-id"),
					Name:             types.StringValue("test-name"),
					IsAuthoritative:  types.BoolValue(true),
					SecurityPolicyID: types.StringValue("security-policy-id"),
					UserIDs:          userIDs,
				}
			},
			expectedState: func() groupModel {
				userIDs, diags := types.SetValueFrom(ctx, types.StringType, []string{"test-user-id-1", "test-user-id-2"})
				if diags.HasError() {
					t.Fatalf("unexpected errors during upgrade: %v", diags.Errors())
				}

				return groupModel{
					ID:              types.StringValue("test-id"),
					Name:            types.StringValue("test-name"),
					IsAuthoritative: types.BoolValue(true),
					UserIDs:         userIDs,
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given: Prior state
			state := tfsdk.State{
				Schema: upgradeGroupStateV0().PriorSchema,
			}
			diags := state.Set(ctx, test.priorState())
			if diags.HasError() {
				t.Fatalf("unexpected errors during upgrade: %v", diags)
			}

			// Mock the request and response
			req := resource.UpgradeStateRequest{
				State: &state,
			}

			newGroup := NewGroupResource()
			newSchema := resource.SchemaResponse{}
			newGroup.Schema(nil, resource.SchemaRequest{}, &newSchema)

			resp := &resource.UpgradeStateResponse{
				State: tfsdk.State{
					Schema: newSchema.Schema,
				},
			}

			// Call the StateUpgrader function under test
			upgradeGroupStateV0().StateUpgrader(ctx, req, resp)

			// Then: Verify the upgraded state
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected errors during upgrade: %v", resp.Diagnostics)
			}

			// Retrieve the upgraded state
			var upgradedState groupModel
			digs := resp.State.Get(ctx, &upgradedState)
			assert.False(t, digs.HasError())

			assert.Equal(t, test.expectedState(), upgradedState)
		})
	}

}
