package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.List = atLeastOneNonEmptyValidator{}

type atLeastOneNonEmptyValidator struct {
	sibling path.Expression
}

func (v atLeastOneNonEmptyValidator) Description(_ context.Context) string {
	return fmt.Sprintf("at least one of this list or %q must be non-empty", v.sibling)
}

func (v atLeastOneNonEmptyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v atLeastOneNonEmptyValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.IsNull() && len(req.ConfigValue.Elements()) > 0 {
		return
	}

	// Current list is empty — check the expression.
	paths, diags := req.Config.PathMatches(ctx, v.sibling)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, siblingPath := range paths {
		var siblingVal types.List

		diags = req.Config.GetAttribute(ctx, siblingPath, &siblingVal)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !siblingVal.IsNull() && !siblingVal.IsUnknown() && len(siblingVal.Elements()) > 0 {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Empty configuration",
		fmt.Sprintf("At least one of %q or %q must contain one or more items.", req.Path, v.sibling),
	)
}

// AtLeastOneNonEmptyWith returns a validator that passes when either the current
// list or the list at siblingPath has at least one element.
// Attach it to both lists so the check is symmetric.
func AtLeastOneNonEmptyWith(siblingPath path.Expression) validator.List {
	return atLeastOneNonEmptyValidator{sibling: siblingPath}
}
