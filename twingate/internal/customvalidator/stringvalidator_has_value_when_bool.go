package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = hasValueWhenBoolValidator{}

type hasValueWhenBoolValidator struct {
	sibling  path.Path
	expected bool
	value    string
}

func (v hasValueWhenBoolValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string must be omitted or set to %q when %q is %t", v.value, v.sibling, v.expected)
}

func (v hasValueWhenBoolValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v hasValueWhenBoolValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip on destroy, where the whole config is null.
	if req.Config.Raw.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Omitting the value is always allowed — the fixed value is filled in by the plan.
	if req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == v.value {
		return
	}

	// Current value would override the fixed one — check the sibling.
	var siblingVal types.Bool

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, v.sibling, &siblingVal)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// A null sibling falls back to its schema default and an unknown one
	// is only resolved at apply time, so neither can be checked here.
	if siblingVal.IsNull() || siblingVal.IsUnknown() || siblingVal.ValueBool() != v.expected {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid attribute combination",
		fmt.Sprintf("%q must be omitted or set to %q when %q is %t.", req.Path, v.value, v.sibling, v.expected),
	)
}

// HasValueWhenBoolEquals returns a validator that passes unless the bool attribute
// at siblingPath equals expected, in which case the current string must either be
// omitted or already match value.
func HasValueWhenBoolEquals(siblingPath path.Path, expected bool, value string) validator.String {
	return hasValueWhenBoolValidator{sibling: siblingPath, expected: expected, value: value}
}
