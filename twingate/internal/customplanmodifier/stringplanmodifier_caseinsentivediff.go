package customplanmodifier

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func CaseInsensitiveDiff() planmodifier.String {
	return caseInsensitiveDiffModifier{
		description: "Handles case insensitive strings",
	}
}

type caseInsensitiveDiffModifier struct {
	description string
}

func (m caseInsensitiveDiffModifier) Description(_ context.Context) string {
	return m.description
}

func (m caseInsensitiveDiffModifier) MarkdownDescription(_ context.Context) string {
	return m.description
}

func (m caseInsensitiveDiffModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do not replace on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do not replace on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() && req.StateValue.IsNull() {
		return
	}

	if strings.EqualFold(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
