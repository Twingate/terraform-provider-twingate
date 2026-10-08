package model

import (
	"regexp"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
)

// matchName reports whether value satisfies the given name filter operation.
// An unknown operation matches everything: those are delegated to the API by
// HasNotSupportedFilters before we get here.
func matchName(value, name, filterBy string) bool {
	switch filterBy {
	case "":
		return value == name

	case attr.FilterByContains:
		return strings.Contains(value, name)

	case attr.FilterByExclude:
		return !strings.Contains(value, name)

	case attr.FilterByPrefix:
		return strings.HasPrefix(value, name)

	case attr.FilterBySuffix:
		return strings.HasSuffix(value, name)

	case attr.FilterByRegexp:
		matched, err := regexp.MatchString(name, value)

		return err == nil && matched

	default:
		return true
	}
}
