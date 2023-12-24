package datasource

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func addErr(diagnostics *diag.Diagnostics, err error, resource string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operationRead, resource),
		err.Error(),
	)
}

func countOptionalAttributes(attributes ...types.String) int {
	var count int

	for _, attr := range attributes {
		if attr.ValueString() != "" {
			count++
		}
	}

	return count
}
