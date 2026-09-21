package resource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testResourceModel(t *testing.T) resourceModel {
	t.Helper()

	return resourceModel{
		ID:                       types.StringUnknown(),
		Name:                     types.StringValue(testResourceName),
		Address:                  types.StringValue(testResourceAddress),
		RemoteNetworkID:          types.StringValue(testRemoteNetworkID),
		IsAuthoritative:          types.BoolNull(),
		Protocols:                defaultProtocolsObject(),
		AccessPolicy:             makeObjectsSetNull(t.Context(), accessPolicyAttributeTypes()),
		GroupAccess:              makeObjectsSetNull(t.Context(), accessGroupAttributeTypes()),
		ServiceAccess:            makeObjectsSetNull(t.Context(), accessServiceAccountAttributeTypes()),
		IsActive:                 types.BoolValue(true),
		IsVisible:                types.BoolNull(),
		IsBrowserShortcutEnabled: types.BoolNull(),
		Alias:                    types.StringNull(),
		SecurityPolicyID:         types.StringNull(),
		RoutingMode:              types.StringNull(),
		Tags:                     types.MapNull(types.StringType),
		TagsAll:                  types.MapNull(types.StringType),
	}
}

// The API receives tags_all (user tags merged with the provider default_tags),
// never the user-declared tags attribute on its own.
func TestConvertResourceTags(t *testing.T) {
	cases := []struct {
		name     string
		tags     types.Map
		tagsAll  types.Map
		expected map[string]string
	}{
		{
			name:     "no tags - nil",
			tags:     types.MapNull(types.StringType),
			tagsAll:  types.MapNull(types.StringType),
			expected: nil,
		},
		{
			name:     "tags_all not yet known - nil",
			tags:     stringMap(map[string]string{"env": "prod"}),
			tagsAll:  types.MapUnknown(types.StringType),
			expected: nil,
		},
		{
			name:     "tags_all - converted, user tags alone are ignored",
			tags:     stringMap(map[string]string{"env": "prod"}),
			tagsAll:  stringMap(map[string]string{"env": "prod", "owner": "platform"}),
			expected: map[string]string{"env": "prod", "owner": "platform"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			model := testResourceModel(t)
			model.Tags = c.tags
			model.TagsAll = c.tagsAll

			converted, err := convertResource(&model)
			require.NoError(t, err)

			assert.Equal(t, c.expected, converted.Tags)
		})
	}
}
