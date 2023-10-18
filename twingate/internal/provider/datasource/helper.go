package datasource

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
