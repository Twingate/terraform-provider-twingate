package datasource

import (
	"fmt"
	"slices"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func CountOptionalAttributes(attributes ...types.String) int {
	var count int

	for _, attr := range attributes {
		if attr.ValueString() != "" {
			count++
		}
	}

	return count
}

func GetNameFilter(name, nameRegexp, nameContains, nameExclude, namePrefix, nameSuffix types.String) (string, string) {
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

func CountSetAttributes(attributes ...types.Set) int {
	var count int

	for _, attribute := range attributes {
		if len(attribute.Elements()) > 0 {
			count++
		}
	}

	return count
}

// SetValues converts a set of strings into a sorted slice. The order is stable so
// that the datasource id does not depend on the order in the config.
// Null or unknown sets and elements, and empty strings, are skipped.
func SetValues(set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var values []string

	for _, item := range set.Elements() {
		value, ok := item.(types.String)
		if !ok || value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
			continue
		}

		values = append(values, value.ValueString())
	}

	if len(values) == 0 {
		return nil
	}

	slices.Sort(values)

	return values
}

// GetStringFilter builds the filter on a single string field from the mutually
// exclusive datasource attributes describing it.
func GetStringFilter(valueIn types.Set, value, valueRegexp, valueContains, valueExclude, valuePrefix, valueSuffix types.String) *client.StringFilter {
	filter := &client.StringFilter{}
	filter.Name, filter.Filter = GetNameFilter(value, valueRegexp, valueContains, valueExclude, valuePrefix, valueSuffix)

	if values := SetValues(valueIn); len(values) > 0 {
		filter.Values = values
		filter.Filter = attr.FilterByIn
	}

	return filter
}

// InFilterAttribute describes a `<field>_in` filter: match any of the exact values in the list.
func InFilterAttribute(description string) schema.SetAttribute {
	return schema.SetAttribute{
		Optional:    true,
		ElementType: types.StringType,
		Description: description,
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
			setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
		},
	}
}
