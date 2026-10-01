package customplanmodifier

import (
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

// Attribute names mirroring the in_cluster/address pair of
// twingate_kubernetes_resource, the first user of this plan modifier.
const (
	testInCluster      = "in_cluster"
	testAddress        = "address"
	testDefaultAddress = "kubernetes.default.svc.cluster.local"
	testCustomAddress  = "k8s-api.example.com"
)

var (
	siblingTrue    = tftypes.NewValue(tftypes.Bool, true)
	siblingFalse   = tftypes.NewValue(tftypes.Bool, false)
	siblingNull    = tftypes.NewValue(tftypes.Bool, nil)
	siblingUnknown = tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)

	inClusterAddressType = tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			testInCluster: tftypes.Bool,
			testAddress:   tftypes.String,
		},
	}
)

// inClusterAddressPlan wraps raw in a plan whose schema holds one bool and one
// string attribute.
func inClusterAddressPlan(raw tftypes.Value) tfsdk.Plan {
	return tfsdk.Plan{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				testInCluster: schema.BoolAttribute{Optional: true, Computed: true},
				testAddress:   schema.StringAttribute{Optional: true, Computed: true},
			},
		},
		Raw: raw,
	}
}

// planWithSibling builds a plan where the bool sibling has the given value and the
// string attribute is still unknown, as it is before the modifier runs on create.
func planWithSibling(sibling tftypes.Value) tfsdk.Plan {
	return inClusterAddressPlan(tftypes.NewValue(inClusterAddressType, map[string]tftypes.Value{
		testInCluster: sibling,
		testAddress:   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	}))
}

func TestUseStringWhenBoolEquals(t *testing.T) {
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
			expected:    types.StringValue(testDefaultAddress),
		},
		{
			name:        "sibling matches, value omitted with prior state - replans to the fixed value",
			sibling:     siblingTrue,
			configValue: types.StringNull(),
			planValue:   types.StringValue(testCustomAddress),
			expected:    types.StringValue(testDefaultAddress),
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
			configValue: types.StringValue(testDefaultAddress),
			planValue:   types.StringValue(testDefaultAddress),
			expected:    types.StringValue(testDefaultAddress),
		},
		{
			name:        "sibling does not match, custom value - never overrides the config",
			sibling:     siblingFalse,
			configValue: types.StringValue(testCustomAddress),
			planValue:   types.StringValue(testCustomAddress),
			expected:    types.StringValue(testCustomAddress),
		},
		{
			name:        "sibling unknown, custom value - no error",
			sibling:     siblingUnknown,
			configValue: types.StringValue(testCustomAddress),
			planValue:   types.StringValue(testCustomAddress),
			expected:    types.StringValue(testCustomAddress),
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
			configValue:    types.StringValue(testCustomAddress),
			planValue:      types.StringValue(testCustomAddress),
			expected:       types.StringValue(testCustomAddress),
			expectedDetail: `"address" must be omitted or set to "kubernetes.default.svc.cluster.local" when "in_cluster" is true.`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &planmodifier.StringResponse{PlanValue: c.planValue}

			UseStringWhenBoolEquals(path.Root(testInCluster), true, testDefaultAddress).PlanModifyString(t.Context(), planmodifier.StringRequest{
				Path:        path.Root(testAddress),
				ConfigValue: c.configValue,
				PlanValue:   c.planValue,
				Plan:        planWithSibling(c.sibling),
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

func TestUseStringWhenBoolEqualsDescription(t *testing.T) {
	modifier := UseStringWhenBoolEquals(path.Root(testInCluster), true, testDefaultAddress)

	expected := `value defaults to "kubernetes.default.svc.cluster.local" when "in_cluster" is true`
	assert.Equal(t, expected, modifier.Description(t.Context()))
	assert.Equal(t, expected, modifier.MarkdownDescription(t.Context()))
}

// On destroy the whole plan is null: there is no sibling to read and nothing to
// pin, so the modifier must leave the plan untouched without diagnostics.
func TestUseStringWhenBoolEqualsSkipsDestroy(t *testing.T) {
	resp := &planmodifier.StringResponse{PlanValue: types.StringNull()}

	UseStringWhenBoolEquals(path.Root(testInCluster), true, testDefaultAddress).PlanModifyString(t.Context(), planmodifier.StringRequest{
		Path:        path.Root(testAddress),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringNull(),
		Plan:        inClusterAddressPlan(tftypes.NewValue(inClusterAddressType, nil)),
	}, resp)

	assert.Equal(t, types.StringNull(), resp.PlanValue)
	assert.False(t, resp.Diagnostics.HasError())
}

// A sibling path missing from the schema is a wiring mistake in the resource: it
// must surface as a diagnostic instead of silently skipping the modifier.
func TestUseStringWhenBoolEqualsSiblingLookupError(t *testing.T) {
	resp := &planmodifier.StringResponse{PlanValue: types.StringUnknown()}

	UseStringWhenBoolEquals(path.Root("missing"), true, testDefaultAddress).PlanModifyString(t.Context(), planmodifier.StringRequest{
		Path:        path.Root(testAddress),
		ConfigValue: types.StringNull(),
		PlanValue:   types.StringUnknown(),
		Plan:        planWithSibling(siblingTrue),
	}, resp)

	assert.Equal(t, types.StringUnknown(), resp.PlanValue)
	require.True(t, resp.Diagnostics.HasError())
	require.Len(t, resp.Diagnostics.Errors(), 1)
	assert.Equal(t, "Plan Read Error", resp.Diagnostics.Errors()[0].Summary())
}
