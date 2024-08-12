package resource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func setIntersectionGroupAccess(inputA, inputB []model.AccessGroup) []model.AccessGroup {
	setA := map[string]model.AccessGroup{}
	setB := map[string]model.AccessGroup{}

	for _, access := range inputA {
		setA[access.GroupID] = access
	}

	for _, access := range inputB {
		setB[access.GroupID] = access
	}

	result := make([]model.AccessGroup, 0, len(setA))

	for key := range setA {
		if val, exist := setB[key]; exist {
			result = append(result, val)
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

func setDifferenceGroupAccess(inputA, inputB []model.AccessGroup) []model.AccessGroup {
	setA := map[string]model.AccessGroup{}
	setB := map[string]model.AccessGroup{}

	for _, access := range inputA {
		setA[access.GroupID] = access
	}

	for _, access := range inputB {
		setB[access.GroupID] = access
	}

	result := make([]model.AccessGroup, 0, len(setA))

	for key, valA := range setA {
		if valB, exist := setB[key]; !exist || !valA.Equals(valB) {
			result = append(result, valA)
		}
	}

	return result
}

func setDifferenceGroups(inputA, inputB []model.AccessGroup) []string {
	groupsA := utils.Map(inputA, func(item model.AccessGroup) string {
		return item.GroupID
	})

	groupsB := utils.Map(inputB, func(item model.AccessGroup) string {
		return item.GroupID
	})

	return setDifference(groupsA, groupsB)
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

func makeNullObject(attributeTypes map[string]tfattr.Type) types.Object {
	return types.ObjectNull(attributeTypes)
}

func makeObjectsSetNull(ctx context.Context, attributeTypes map[string]tfattr.Type) types.Set {
	return types.SetNull(types.ObjectNull(attributeTypes).Type(ctx))
}

func makeObjectsSet(ctx context.Context, objects ...types.Object) (types.Set, diag.Diagnostics) {
	obj := objects[0]

	items := utils.Map(objects, func(item types.Object) tfattr.Value {
		return tfattr.Value(item)
	})

	return types.SetValue(obj.Type(ctx), items)
}

// setUnion - for given two sets A and B,
// If A = {1, 2} and B = {3, 4}, then the union of A and B is {1, 2, 3, 4}.
func setUnion(setA, setB []string) []string {
	if len(setA) == 0 {
		return setB
	}

	if len(setB) == 0 {
		return setA
	}

	set := utils.MakeLookupMap(setA)

	for _, key := range setB {
		set[key] = true
	}

	result := make([]string, 0, len(set))
	for key := range set {
		result = append(result, key)
	}

	return result
}
