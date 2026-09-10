package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = nonEmptyWhenBoolValidator{}

type nonEmptyWhenBoolValidator struct {
	sibling  path.Path
	expected bool
}

func (v nonEmptyWhenBoolValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string must be set to a non-empty value when %q is %t", v.sibling, v.expected)
}

func (v nonEmptyWhenBoolValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v nonEmptyWhenBoolValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip on destroy, where the whole config is null.
	if req.Config.Raw.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.IsNull() && req.ConfigValue.ValueString() != "" {
		return
	}

	// Current value is missing or empty — check the sibling.
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
		"Missing required attribute",
		fmt.Sprintf("%q must be set to a non-empty value when %q is %t.", req.Path, v.sibling, v.expected),
	)
}

// NonEmptyWhenBoolEquals returns a validator that passes unless the bool attribute
// at siblingPath equals expected, in which case the current string must be set to
// a non-empty value.
func NonEmptyWhenBoolEquals(siblingPath path.Path, expected bool) validator.String {
	return nonEmptyWhenBoolValidator{sibling: siblingPath, expected: expected}
}
