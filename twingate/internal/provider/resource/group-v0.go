package resource

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type groupModelV0 struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	IsAuthoritative  types.Bool   `tfsdk:"is_authoritative"`
	UserIDs          types.Set    `tfsdk:"user_ids"`
	SecurityPolicyID types.String `tfsdk:"security_policy_id"`
}

func upgradeGroupStateV0() resource.StateUpgrader {
	return resource.StateUpgrader{
		PriorSchema: &schema.Schema{
			Attributes: map[string]schema.Attribute{
				attr.ID: schema.StringAttribute{
					Computed: true,
				},
				attr.Name: schema.StringAttribute{
					Required: true,
				},
				attr.IsAuthoritative: schema.BoolAttribute{
					Optional: true,
					Computed: true,
				},
				attr.UserIDs: schema.SetAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
				attr.SecurityPolicyID: schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
			},
		},

		StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
			var priorState groupModelV0

			resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)

			if resp.Diagnostics.HasError() {
				return
			}

			upgradedState := groupModel{
				ID:              priorState.ID,
				Name:            priorState.Name,
				IsAuthoritative: priorState.IsAuthoritative,
				UserIDs:         priorState.UserIDs,
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)
		},
	}
}
