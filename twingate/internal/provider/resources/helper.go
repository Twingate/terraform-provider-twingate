package resources

import "github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"

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

// setJoin - joins two sets.
// The join of sets A and set B denoted as A + B.
// If A = {1, 2, 3, 4} and B = {3, 4, 5, 7}, then the join of sets A and B is given by A + B = {1, 2, 3, 4, 5, 7}.
func setJoin(a, b []string) []string {
	result := make([]string, 0, len(a)+len(b))
	result = append(result, a...)
	result = append(result, b...)

	return utils.MapKeys(utils.MakeLookupMap(result))
}
