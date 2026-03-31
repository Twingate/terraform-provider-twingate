package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = alsoRequiresWhenValueValidator{}

type alsoRequiresWhenValueValidator struct {
	expression path.Expression
	value      string
}

func (v alsoRequiresWhenValueValidator) Description(_ context.Context) string {
	return fmt.Sprintf("when this attribute equals %q, %q must also be set", v.value, v.expression)
}

func (v alsoRequiresWhenValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v alsoRequiresWhenValueValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.ValueString() != v.value {
		return
	}

	paths, diags := req.Config.PathMatches(ctx, req.PathExpression.Merge(v.expression))
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, attributePath := range paths {
		var attributeVal types.String

		diags = req.Config.GetAttribute(ctx, attributePath, &attributeVal)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if attributeVal.IsNull() || attributeVal.IsUnknown() || attributeVal.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid attribute combination",
				fmt.Sprintf("%q is %q, so %q must also be set.", req.Path, v.value, attributePath),
			)
		}
	}
}

// AlsoRequiresWhenValueIs returns a validator that, when the current
// attribute equals triggerValue, requires the expression attribute at siblingPath
// to be set (non-null, non-empty). Attach this to the attribute whose value
// controls whether the expression is required (e.g. type = "iam" requires
// service_account_email).
func AlsoRequiresWhenValueIs(expression path.Expression, triggerValue string) validator.String {
	return alsoRequiresWhenValueValidator{expression: expression, value: triggerValue}
}
