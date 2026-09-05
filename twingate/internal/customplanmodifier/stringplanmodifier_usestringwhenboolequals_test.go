package customplanmodifier

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseStringWhenBoolEquals(t *testing.T) {
	const (
		inCluster      = "in_cluster"
		address        = "address"
		defaultAddress = "kubernetes.default.svc.cluster.local"
		customAddress  = "k8s-api.example.com"
	)

	planSchema := schema.Schema{
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

	buildPlan := func(sibling tftypes.Value) tfsdk.Plan {
		return tfsdk.Plan{
			Schema: planSchema,
			Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
				inCluster: sibling,
				address:   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
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
		planValue      types.String
		expected       types.String
		expectedDetail string
	}{
		{
			name:        "sibling matches, value omitted on create - plans the fixed value",
			sibling:     siblingTrue,
			configValue: types.StringNull(),
			planValue:   types.StringUnknown(),
			expected:    types.StringValue(defaultAddress),
		},
		{
			name:        "sibling matches, value omitted with prior state - replans to the fixed value",
			sibling:     siblingTrue,
			configValue: types.StringNull(),
			planValue:   types.StringValue(customAddress),
			expected:    types.StringValue(defaultAddress),
		},
		{
			name:        "sibling does not match - leaves the plan untouched",
			sibling:     siblingFalse,
			configValue: types.StringNull(),
			planValue:   types.StringUnknown(),
			expected:    types.StringUnknown(),
		},
		{
			name:        "sibling unknown - leaves the plan untouched",
			sibling:     siblingUnknown,
			configValue: types.StringNull(),
			planValue:   types.StringUnknown(),
			expected:    types.StringUnknown(),
		},
		{
			name:        "sibling null - leaves the plan untouched",
			sibling:     siblingNull,
			configValue: types.StringNull(),
			planValue:   types.StringUnknown(),
			expected:    types.StringUnknown(),
		},
		{
			name:        "sibling matches, value pinned to the fixed value - leaves the config alone",
			sibling:     siblingTrue,
			configValue: types.StringValue(defaultAddress),
			planValue:   types.StringValue(defaultAddress),
			expected:    types.StringValue(defaultAddress),
		},
		{
			name:        "sibling does not match, custom value - never overrides the config",
			sibling:     siblingFalse,
			configValue: types.StringValue(customAddress),
			planValue:   types.StringValue(customAddress),
			expected:    types.StringValue(customAddress),
		},
		{
			name:        "sibling unknown, custom value - no error",
			sibling:     siblingUnknown,
			configValue: types.StringValue(customAddress),
			planValue:   types.StringValue(customAddress),
			expected:    types.StringValue(customAddress),
		},
		{
			name:        "sibling matches, value unknown - no error",
			sibling:     siblingTrue,
			configValue: types.StringUnknown(),
			planValue:   types.StringUnknown(),
			expected:    types.StringUnknown(),
		},
		{
			name:           "sibling defaulted in the plan, custom value - reports error",
			sibling:        siblingTrue,
			configValue:    types.StringValue(customAddress),
			planValue:      types.StringValue(customAddress),
			expected:       types.StringValue(customAddress),
			expectedDetail: `"address" must be omitted or set to "kubernetes.default.svc.cluster.local" when "in_cluster" is true.`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &planmodifier.StringResponse{PlanValue: c.planValue}

			UseStringWhenBoolEquals(path.Root(inCluster), true, defaultAddress).PlanModifyString(context.Background(), planmodifier.StringRequest{
				Path:        path.Root(address),
				ConfigValue: c.configValue,
				PlanValue:   c.planValue,
				Plan:        buildPlan(c.sibling),
			}, resp)

			assert.Equal(t, c.expected, resp.PlanValue)

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
