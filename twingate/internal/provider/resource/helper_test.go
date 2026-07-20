package resource

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSetIntersection(t *testing.T) {
	cases := []struct {
		a        []string
		b        []string
		expected []string
	}{
		{
			a:        []string{"1", "2", "3"},
			b:        []string{"0", "2", "1", "5"},
			expected: []string{"1", "2"},
		},
		{
			a:        []string{"0", "2", "1", "5"},
			b:        []string{"1", "2", "3"},
			expected: []string{"1", "2"},
		},
		{
			a:        []string{"0", "2", "1", "5", "2"},
			b:        []string{"1", "2", "3"},
			expected: []string{"1", "2"},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := setIntersection(c.a, c.b)

			assert.ElementsMatch(t, c.expected, actual)
		})
	}
}

func TestSetDifference(t *testing.T) {
	cases := []struct {
		a        []string
		b        []string
		expected []string
	}{
		{
			a:        []string{"1", "2", "3"},
			b:        []string{"0", "2"},
			expected: []string{"1", "3"},
		},
		{
			a:        []string{"0", "2", "1", "5"},
			b:        []string{"1", "2", "3"},
			expected: []string{"0", "5"},
		},
		{
			a:        []string{"1"},
			b:        []string{"2"},
			expected: []string{"1"},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := setDifference(c.a, c.b)

			assert.ElementsMatch(t, c.expected, actual)
		})
	}
}

func TestStringPtr(t *testing.T) {
	val := "value"
	emptyStr := ""

	cases := []struct {
		input    string
		expected *string
	}{
		{
			input:    emptyStr,
			expected: &emptyStr,
		},
		{
			input:    val,
			expected: &val,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, stringPtr(c.input))
		})
	}
}

func TestBoolPtr(t *testing.T) {
	valTrue := true
	valFalse := false

	cases := []struct {
		input    bool
		expected *bool
	}{
		{
			input:    valTrue,
			expected: &valTrue,
		},
		{
			input:    valFalse,
			expected: &valFalse,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, boolPtr(c.input))
		})
	}
}

func TestWithDefaultValue(t *testing.T) {
	cases := []struct {
		input      string
		defaultVal string
		expected   string
	}{
		{
			input:      "",
			defaultVal: "default",
			expected:   "default",
		},
		{
			input:      "val",
			defaultVal: "default",
			expected:   "val",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, withDefaultValue(c.input, c.defaultVal))
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestIsWildcardAddress(t *testing.T) {
	cases := []struct {
		address  string
		expected bool
	}{
		{
			address:  "hello.com",
			expected: false,
		},
		{
			address:  "*.hello.com",
			expected: true,
		},
		{
			address:  "redis-?-blah.internal",
			expected: true,
		},
		{
			address:  "redis-*-blah.internal",
			expected: true,
		},
		{
			address:  "10.0.0.0/16",
			expected: true,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, isWildcardAddress(c.address))
		})
	}
}

func TestValidateBypassRoutingMode(t *testing.T) {
	through := model.RoutingModeThroughTwingate
	bypass := model.RoutingModeBypassTwingate
	policy := "policy-id"

	allowAll := &model.Protocols{TCP: model.DefaultProtocol(), UDP: model.DefaultProtocol()}
	restrictedTCP := &model.Protocols{TCP: model.NewProtocol(model.PolicyRestricted, nil), UDP: model.DefaultProtocol()}
	denyAllUDP := &model.Protocols{TCP: model.DefaultProtocol(), UDP: model.NewProtocol(model.PolicyDenyAll, nil)}

	cases := []struct {
		name          string
		routingMode   *string
		address       string
		securityPol   types.String
		accessGroups  []model.AccessGroup
		protocols     *model.Protocols
		expectedError error
	}{
		{name: "nil routing mode", routingMode: nil, address: "*.example.com", securityPol: types.StringNull(), expectedError: nil},
		{name: "through ignores wildcard", routingMode: &through, address: "*.example.com", securityPol: types.StringNull(), expectedError: nil},
		{name: "through ignores restricted ports", routingMode: &through, address: "public.example.com", securityPol: types.StringNull(), protocols: restrictedTCP, expectedError: nil},
		{name: "bypass clean", routingMode: &bypass, address: "public.example.com", securityPol: types.StringNull(), expectedError: nil},
		{name: "bypass allow-all protocols", routingMode: &bypass, address: "public.example.com", securityPol: types.StringNull(), protocols: allowAll, expectedError: nil},
		{name: "bypass wildcard star", routingMode: &bypass, address: "*.example.com", securityPol: types.StringNull(), expectedError: ErrBypassRoutingWithWildcardAddress},
		{name: "bypass wildcard question", routingMode: &bypass, address: "a?.example.com", securityPol: types.StringNull(), expectedError: ErrBypassRoutingWithWildcardAddress},
		{name: "bypass resource security policy", routingMode: &bypass, address: "public.example.com", securityPol: types.StringValue(policy), expectedError: ErrBypassRoutingWithSecurityPolicy},
		{name: "bypass group security policy", routingMode: &bypass, address: "public.example.com", securityPol: types.StringNull(), accessGroups: []model.AccessGroup{{GroupID: "g1", SecurityPolicyID: &policy}}, expectedError: ErrBypassRoutingWithSecurityPolicy},
		{name: "bypass restricted tcp ports", routingMode: &bypass, address: "public.example.com", securityPol: types.StringNull(), protocols: restrictedTCP, expectedError: ErrBypassRoutingWithPortRestriction},
		{name: "bypass deny-all udp ports", routingMode: &bypass, address: "public.example.com", securityPol: types.StringNull(), protocols: denyAllUDP, expectedError: ErrBypassRoutingWithPortRestriction},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateBypassRoutingMode(c.routingMode, c.address, c.securityPol, c.accessGroups, c.protocols)
			assert.ErrorIs(t, err, c.expectedError)
		})
	}
}
