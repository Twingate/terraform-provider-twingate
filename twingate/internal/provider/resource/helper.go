package resource

import (
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

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
func setDifference(inputA, inputB []string) []string {
	if len(inputA) == 0 {
		return nil
	}

	if len(inputB) == 0 {
		return inputA
	}

	setA := utils.MakeLookupMap(inputA)
	setB := utils.MakeLookupMap(inputB)
	result := make([]string, 0, len(setA))

	for key := range setA {
		if !setB[key] {
			result = append(result, key)
		}
	}

	return result
}

func withDefaultValue(str, defaultValue string) string {
	if str != "" {
		return str
	}

	return defaultValue
}

func addErr(diagnostics *diag.Diagnostics, err error, operation, resource string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operation, resource),
		err.Error(),
	)
}
