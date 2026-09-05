package customvalidator

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonEmptyWhenBoolEquals(t *testing.T) {
	const (
		inCluster       = "in_cluster"
		bearerTokenFile = "bearer_token_file"
	)

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			inCluster:       schema.BoolAttribute{Optional: true, Computed: true},
			bearerTokenFile: schema.StringAttribute{Optional: true, Computed: true},
		},
	}

	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			inCluster:       tftypes.Bool,
			bearerTokenFile: tftypes.String,
		},
	}

	buildConfig := func(sibling, value tftypes.Value) tfsdk.Config {
		return tfsdk.Config{
			Schema: configSchema,
			Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
				inCluster:       sibling,
				bearerTokenFile: value,
			}),
		}
	}

	var (
		siblingFalse   = tftypes.NewValue(tftypes.Bool, false)
		siblingTrue    = tftypes.NewValue(tftypes.Bool, true)
		siblingNull    = tftypes.NewValue(tftypes.Bool, nil)
		siblingUnknown = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)
	)

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

			NonEmptyWhenBoolEquals(path.Root(inCluster), false).ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root(bearerTokenFile),
				ConfigValue: c.configValue,
				Config:      buildConfig(c.sibling, c.rawValue),
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
