package customvalidator

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonEmptyWhenBoolEquals(t *testing.T) {
	cases := []struct {
		name           string
		sibling        tftypes.Value
		configValue    types.String
		rawValue       tftypes.Value
		expectedDetail string
	}{
		{
			name:        "sibling matches, value set - no error",
			sibling:     siblingFalse,
			configValue: types.StringValue("/path/to/token"),
			rawValue:    tftypes.NewValue(tftypes.String, "/path/to/token"),
		},
		{
			name:        "sibling does not match, value null - no error",
			sibling:     siblingTrue,
			configValue: types.StringNull(),
			rawValue:    tftypes.NewValue(tftypes.String, nil),
		},
		{
			name:        "sibling null falls back to default, value null - no error",
			sibling:     siblingNull,
			configValue: types.StringNull(),
			rawValue:    tftypes.NewValue(tftypes.String, nil),
		},
		{
			name:        "sibling unknown, value null - no error",
			sibling:     siblingUnknown,
			configValue: types.StringNull(),
			rawValue:    tftypes.NewValue(tftypes.String, nil),
		},
		{
			name:        "value unknown - no error",
			sibling:     siblingFalse,
			configValue: types.StringUnknown(),
			rawValue:    tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		{
			name:           "sibling matches, value null - reports error",
			sibling:        siblingFalse,
			configValue:    types.StringNull(),
			rawValue:       tftypes.NewValue(tftypes.String, nil),
			expectedDetail: `"bearer_token_file" must be set to a non-empty value when "in_cluster" is false.`,
		},
		{
			name:           "sibling matches, value empty - reports error",
			sibling:        siblingFalse,
			configValue:    types.StringValue(""),
			rawValue:       tftypes.NewValue(tftypes.String, ""),
			expectedDetail: `"bearer_token_file" must be set to a non-empty value when "in_cluster" is false.`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &validator.StringResponse{}

			NonEmptyWhenBoolEquals(path.Root(testInCluster), false).ValidateString(t.Context(), validator.StringRequest{
				Path:        path.Root(testBearerTokenFile),
				ConfigValue: c.configValue,
				Config:      boolSiblingConfig(testInCluster, testBearerTokenFile, c.sibling, c.rawValue),
			}, resp)

			if c.expectedDetail == "" {
				assert.False(t, resp.Diagnostics.HasError())

				return
			}

			require.True(t, resp.Diagnostics.HasError())
			require.Len(t, resp.Diagnostics.Errors(), 1)
			assert.Equal(t, "Missing required attribute", resp.Diagnostics.Errors()[0].Summary())
			assert.Equal(t, c.expectedDetail, resp.Diagnostics.Errors()[0].Detail())
		})
	}
}

func TestNonEmptyWhenBoolEqualsDescription(t *testing.T) {
	v := NonEmptyWhenBoolEquals(path.Root(testInCluster), false)

	expected := `string must be set to a non-empty value when "in_cluster" is false`
	assert.Equal(t, expected, v.Description(t.Context()))
	assert.Equal(t, expected, v.MarkdownDescription(t.Context()))
}

// On destroy Terraform hands validators a null config, so a missing value must
// not be reported even though the sibling would otherwise require it.
func TestNonEmptyWhenBoolEqualsSkipsDestroy(t *testing.T) {
	resp := &validator.StringResponse{}

	NonEmptyWhenBoolEquals(path.Root(testInCluster), false).ValidateString(t.Context(), validator.StringRequest{
		Path:        path.Root(testBearerTokenFile),
		ConfigValue: types.StringNull(),
		Config:      boolSiblingNullConfig(testInCluster, testBearerTokenFile),
	}, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

// A sibling path missing from the schema is a wiring mistake in the resource: it
// must surface as a diagnostic instead of silently accepting the missing value.
func TestNonEmptyWhenBoolEqualsSiblingLookupError(t *testing.T) {
	resp := &validator.StringResponse{}

	NonEmptyWhenBoolEquals(path.Root("missing"), false).ValidateString(t.Context(), validator.StringRequest{
		Path:        path.Root(testBearerTokenFile),
		ConfigValue: types.StringNull(),
		Config:      boolSiblingConfig(testInCluster, testBearerTokenFile, siblingFalse, tftypes.NewValue(tftypes.String, nil)),
	}, resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Len(t, resp.Diagnostics.Errors(), 1)
	assert.Equal(t, "Configuration Read Error", resp.Diagnostics.Errors()[0].Summary())
}
