package datasource

import (
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
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

func getNameFilter(name, nameRegexp, nameContains, nameExclude, namePrefix, nameSuffix types.String) (string, string) {
	var value, filter string

	if name.ValueString() != "" {
		value = name.ValueString()
	}

	if nameRegexp.ValueString() != "" {
		value = nameRegexp.ValueString()
		filter = attr.FilterByRegexp
	}

	if nameContains.ValueString() != "" {
		value = nameContains.ValueString()
		filter = attr.FilterByContains
	}

	if nameExclude.ValueString() != "" {
		value = nameExclude.ValueString()
		filter = attr.FilterByExclude
	}

	if namePrefix.ValueString() != "" {
		value = namePrefix.ValueString()
		filter = attr.FilterByPrefix
	}

	if nameSuffix.ValueString() != "" {
		value = nameSuffix.ValueString()
		filter = attr.FilterBySuffix
	}

	return value, filter
}
