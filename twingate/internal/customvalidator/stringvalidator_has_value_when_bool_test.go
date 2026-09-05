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

func TestFixedValueWhenBoolEquals(t *testing.T) {
	const (
		inCluster      = "in_cluster"
		address        = "address"
		defaultAddress = "kubernetes.default.svc.cluster.local"
	)

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			inCluster: schema.BoolAttribute{Optional: true, Computed: true},
			address:   schema.StringAttribute{Optional: true, Computed: true},
		},
	}

	objectType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			inCluster: tftypes.Bool,
			address:   tftypes.String,
		},
	}

	buildConfig := func(sibling, value tftypes.Value) tfsdk.Config {
		return tfsdk.Config{
			Schema: configSchema,
			Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
				inCluster: sibling,
				address:   value,
			}),
		}
	}

	var (
		siblingTrue    = tftypes.NewValue(tftypes.Bool, true)
		siblingFalse   = tftypes.NewValue(tftypes.Bool, false)
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
			name:        "sibling matches, value omitted - no error",
			sibling:     siblingTrue,
			configValue: types.StringNull(),
			rawValue:    tftypes.NewValue(tftypes.String, nil),
		},
		{
			name:        "sibling matches, value pinned to the fixed value - no error",
			sibling:     siblingTrue,
			configValue: types.StringValue(defaultAddress),
			rawValue:    tftypes.NewValue(tftypes.String, defaultAddress),
		},
		{
			name:        "sibling does not match, custom value - no error",
			sibling:     siblingFalse,
			configValue: types.StringValue("k8s-api.example.com"),
			rawValue:    tftypes.NewValue(tftypes.String, "k8s-api.example.com"),
		},
		{
			name:        "sibling null falls back to default, custom value - no error",
			sibling:     siblingNull,
			configValue: types.StringValue("k8s-api.example.com"),
			rawValue:    tftypes.NewValue(tftypes.String, "k8s-api.example.com"),
		},
		{
			name:        "sibling unknown, custom value - no error",
			sibling:     siblingUnknown,
			configValue: types.StringValue("k8s-api.example.com"),
			rawValue:    tftypes.NewValue(tftypes.String, "k8s-api.example.com"),
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
			configValue:    types.StringValue("k8s-api.example.com"),
			rawValue:       tftypes.NewValue(tftypes.String, "k8s-api.example.com"),
			expectedDetail: `"address" must be omitted or set to "kubernetes.default.svc.cluster.local" when "in_cluster" is true.`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &validator.StringResponse{}

			HasValueWhenBoolEquals(path.Root(inCluster), true, defaultAddress).ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root(address),
				ConfigValue: c.configValue,
				Config:      buildConfig(c.sibling, c.rawValue),
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
