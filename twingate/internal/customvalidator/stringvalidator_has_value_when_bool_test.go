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

func TestHasValueWhenBoolEquals(t *testing.T) {
	cases := []struct {
		name           string
		sibling        tftypes.Value
		configValue    types.String
		rawValue       tftypes.Value
		expectedDetail string
	}{
		{
			name:        "sibling matches, value omitted - no error",
			sibling:     siblingTrue,
			configValue: types.StringNull(),
			rawValue:    tftypes.NewValue(tftypes.String, nil),
		},
		{
			name:        "sibling matches, value pinned to the fixed value - no error",
			sibling:     siblingTrue,
			configValue: types.StringValue(testDefaultAddress),
			rawValue:    tftypes.NewValue(tftypes.String, testDefaultAddress),
		},
		{
			name:        "sibling does not match, custom value - no error",
			sibling:     siblingFalse,
			configValue: types.StringValue(testCustomAddress),
			rawValue:    tftypes.NewValue(tftypes.String, testCustomAddress),
		},
		{
			name:        "sibling null falls back to default, custom value - no error",
			sibling:     siblingNull,
			configValue: types.StringValue(testCustomAddress),
			rawValue:    tftypes.NewValue(tftypes.String, testCustomAddress),
		},
		{
			name:        "sibling unknown, custom value - no error",
			sibling:     siblingUnknown,
			configValue: types.StringValue(testCustomAddress),
			rawValue:    tftypes.NewValue(tftypes.String, testCustomAddress),
		},
		{
			name:        "value unknown - no error",
			sibling:     siblingTrue,
			configValue: types.StringUnknown(),
			rawValue:    tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		{
			name:           "sibling matches, custom value - reports error",
			sibling:        siblingTrue,
			configValue:    types.StringValue(testCustomAddress),
			rawValue:       tftypes.NewValue(tftypes.String, testCustomAddress),
			expectedDetail: `"address" must be omitted or set to "kubernetes.default.svc.cluster.local" when "in_cluster" is true.`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &validator.StringResponse{}

			HasValueWhenBoolEquals(path.Root(testInCluster), true, testDefaultAddress).ValidateString(t.Context(), validator.StringRequest{
				Path:        path.Root(testAddress),
				ConfigValue: c.configValue,
				Config:      boolSiblingConfig(testInCluster, testAddress, c.sibling, c.rawValue),
			}, resp)

			if c.expectedDetail == "" {
				assert.False(t, resp.Diagnostics.HasError())

				return
			}

			require.True(t, resp.Diagnostics.HasError())
			require.Len(t, resp.Diagnostics.Errors(), 1)
			assert.Equal(t, "Invalid attribute combination", resp.Diagnostics.Errors()[0].Summary())
			assert.Equal(t, c.expectedDetail, resp.Diagnostics.Errors()[0].Detail())
		})
	}
}

func TestHasValueWhenBoolEqualsDescription(t *testing.T) {
	v := HasValueWhenBoolEquals(path.Root(testInCluster), true, testDefaultAddress)

	expected := `string must be omitted or set to "kubernetes.default.svc.cluster.local" when "in_cluster" is true`
	assert.Equal(t, expected, v.Description(t.Context()))
	assert.Equal(t, expected, v.MarkdownDescription(t.Context()))
}

// On destroy Terraform hands validators a null config, so a value that would
// otherwise conflict with the sibling must not be reported.
func TestHasValueWhenBoolEqualsSkipsDestroy(t *testing.T) {
	resp := &validator.StringResponse{}

	HasValueWhenBoolEquals(path.Root(testInCluster), true, testDefaultAddress).ValidateString(t.Context(), validator.StringRequest{
		Path:        path.Root(testAddress),
		ConfigValue: types.StringValue(testCustomAddress),
		Config:      boolSiblingNullConfig(testInCluster, testAddress),
	}, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

// A sibling path missing from the schema is a wiring mistake in the resource: it
// must surface as a diagnostic instead of silently accepting the value.
func TestHasValueWhenBoolEqualsSiblingLookupError(t *testing.T) {
	resp := &validator.StringResponse{}

	HasValueWhenBoolEquals(path.Root("missing"), true, testDefaultAddress).ValidateString(t.Context(), validator.StringRequest{
		Path:        path.Root(testAddress),
		ConfigValue: types.StringValue(testCustomAddress),
		Config:      boolSiblingConfig(testInCluster, testAddress, siblingTrue, tftypes.NewValue(tftypes.String, testCustomAddress)),
	}, resp)

	require.True(t, resp.Diagnostics.HasError())
	require.Len(t, resp.Diagnostics.Errors(), 1)
	assert.Equal(t, "Configuration Read Error", resp.Diagnostics.Errors()[0].Summary())
}
