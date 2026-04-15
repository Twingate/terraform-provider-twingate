package customplanmodifier

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func UseDefaultTagsForUnknownModifier(defaultTags *map[string]string) planmodifier.Map {
	return useDefaultTagsForUnknownModifier{
		defaultTags: defaultTags,
	}
}

// useDefaultTagsForUnknownModifier implements the plan modifier.
type useDefaultTagsForUnknownModifier struct {
	defaultTags *map[string]string
}

// Description returns a human-readable description of the plan modifier.
func (m useDefaultTagsForUnknownModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute will fallback to Default Tags on unset."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useDefaultTagsForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute will fallback to Default Tags on unset."
}

// PlanModifyMap implements the plan modification logic.
func (m useDefaultTagsForUnknownModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	tags := types.MapValueMust(types.StringType, map[string]tfattr.Value{})
	req.Config.GetAttribute(ctx, path.Root(attr.Tags), &tags)

	resp.PlanValue = utils.ConvertMapValue(utils.MapUnion(*m.defaultTags, utils.ConvertMap(tags)))
}
