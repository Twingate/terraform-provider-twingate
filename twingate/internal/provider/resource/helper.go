package resource

import (
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ErrAttributeSet(err error, attribute string) diag.Diagnostics {
	return diag.FromErr(fmt.Errorf("error setting %s: %w ", attribute, err))
}

func castToStrings(a, b interface{}) (string, string) {
	return a.(string), b.(string)
}

func convertIDs(data interface{}) []string {
	return utils.Map[interface{}, string](
		data.(*schema.Set).List(),
		func(elem interface{}) string {
			return elem.(string)
		},
	)
}

// setIntersection - for given two sets A and B,
// A ∩ B (read as A intersection B) is the set of common elements that belong to set A and B.
// If A = {1, 2, 3, 4} and B = {3, 4, 5, 7}, then the intersection of A and B is given by A ∩ B = {3, 4}.
func setIntersection(a, b []string) []string {
	setA := utils.MakeLookupMap(a)
	setB := utils.MakeLookupMap(b)
	result := make([]string, 0, len(setA))

	for key := range setA {
		if setB[key] {
			result = append(result, key)
		}
	}

	return result
}

// setDifference - difference between sets implies subtracting the elements from a set.
// The difference between sets A and set B denoted as A − B.
// If A = {1, 2, 3, 4} and B = {3, 4, 5, 7}, then the difference between sets A and B is given by A - B = {1, 2}.
func setDifference(a, b []string) []string {
	setA := utils.MakeLookupMap(a)
	setB := utils.MakeLookupMap(b)
	result := make([]string, 0, len(setA))

	for key := range setA {
		if !setB[key] {
			result = append(result, key)
		}
	}

	return result
}

func getOptionalBoolFlag(data *schema.ResourceData, attribute string) *bool {
	isVisible, ok := data.GetOkExists(attribute) //nolint
	if val := isVisible.(bool); ok {
		return &val
	}

	return nil
}
