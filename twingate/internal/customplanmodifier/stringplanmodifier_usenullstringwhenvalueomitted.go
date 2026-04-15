package customplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func UseNullStringWhenValueOmitted() planmodifier.String {
	return useNullStringWhenValueOmitted{}
}

type useNullStringWhenValueOmitted struct{}

func (m useNullStringWhenValueOmitted) Description(_ context.Context) string {
	return ""
}

func (m useNullStringWhenValueOmitted) MarkdownDescription(_ context.Context) string {
	return ""
}

func (m useNullStringWhenValueOmitted) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() && req.ConfigValue.IsNull() {
		resp.PlanValue = types.StringNull()

		return
	}

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.IsNull() && !req.PlanValue.IsNull() {
		resp.PlanValue = types.StringNull()
	}
}
