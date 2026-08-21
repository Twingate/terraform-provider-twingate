package customvalidator

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Map = caseInsensitiveUniqueKeysValidator{}

type caseInsensitiveUniqueKeysValidator struct{}

func (v caseInsensitiveUniqueKeysValidator) Description(_ context.Context) string {
	return "map keys must be unique when compared case-insensitively"
}

func (v caseInsensitiveUniqueKeysValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v caseInsensitiveUniqueKeysValidator) ValidateMap(_ context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	keys := make([]string, 0, len(req.ConfigValue.Elements()))
	for key := range req.ConfigValue.Elements() {
		keys = append(keys, key)
	}

	// Element order is not stable, so sort to keep the reported collision
	// deterministic across runs.
	sort.Strings(keys)

	seen := make(map[string]string, len(keys))

	for _, key := range keys {
		normalized := strings.ToLower(key)

		if original, exists := seen[normalized]; exists {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate key",
				fmt.Sprintf("Keys %q and %q differ only by case.", original, key),
			)

			return
		}

		seen[normalized] = key
	}
}

// CaseInsensitiveUniqueKeys returns a validator that rejects a map whose keys are
// only distinct by case. Terraform treats map keys as case-sensitive, so without
// this the collision would surface as an opaque API error during apply.
func CaseInsensitiveUniqueKeys() validator.Map {
	return caseInsensitiveUniqueKeysValidator{}
}
