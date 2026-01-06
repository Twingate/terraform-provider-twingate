package customplanmodifier

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Duration() planmodifier.String {
	return duration{}
}

type duration struct{}

func (m duration) Description(_ context.Context) string {
	return ""
}

func (m duration) MarkdownDescription(_ context.Context) string {
	return ""
}

func (m duration) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	if equalDuration(req.StateValue, req.PlanValue) {
		resp.PlanValue = req.StateValue
	}
}

func equalDuration(stateDuration, planDuraion types.String) bool {
	stateVal, err := utils.ParseDurationWithDays(stateDuration.ValueString())
	if err != nil {
		return false
	}

	planVal, err := utils.ParseDurationWithDays(planDuraion.ValueString())
	if err != nil {
		return false
	}

	return stateVal == planVal
}
