package customplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.String = useStringWhenBoolEquals{}

// UseStringWhenBoolEquals returns a plan modifier that pins an attribute to value
// whenever the bool attribute at siblingPath equals expected: an omitted attribute
// is planned as value, and one configured to anything else is reported as an error.
func UseStringWhenBoolEquals(siblingPath path.Path, expected bool, value string) planmodifier.String {
	return useStringWhenBoolEquals{sibling: siblingPath, expected: expected, value: value}
}

type useStringWhenBoolEquals struct {
	sibling  path.Path
	expected bool
	value    string
}

func (m useStringWhenBoolEquals) Description(_ context.Context) string {
	return fmt.Sprintf("value defaults to %q when %q is %t", m.value, m.sibling, m.expected)
}

func (m useStringWhenBoolEquals) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStringWhenBoolEquals) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Skip on destroy, where the whole plan is null.
	if req.Plan.Raw.IsNull() {
		return
	}

	var siblingVal types.Bool

	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, m.sibling, &siblingVal)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Reading the sibling from the plan rather than the config means schema
	// defaults are already applied, so an omitted sibling is seen as its default.
	// An unknown one is only resolved at apply time and cannot be checked here.
	if siblingVal.IsNull() || siblingVal.IsUnknown() || siblingVal.ValueBool() != m.expected {
		return
	}

	// Terraform requires the planned value of a configured attribute to match the
	// config, so a conflicting value can only be reported, never corrected.
	if !req.ConfigValue.IsNull() {
		if !req.ConfigValue.IsUnknown() && req.ConfigValue.ValueString() != m.value {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid attribute combination",
				fmt.Sprintf("%q must be omitted or set to %q when %q is %t.", req.Path, m.value, m.sibling, m.expected),
			)
		}

		return
	}

	resp.PlanValue = types.StringValue(m.value)
}
