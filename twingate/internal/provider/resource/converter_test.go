package resource

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestConvertProtocol(t *testing.T) {

	cases := []struct {
		input       types.Object
		expected    *model.Protocol
		expectedErr error
	}{
		{},
		{
			input: types.ObjectValueMust(protocolAttributeTypes(), map[string]tfattr.Value{
				attr.Policy: types.StringValue(model.PolicyAllowAll),
				attr.Ports:  makeTestSet("-"),
			}),
			expectedErr: errors.New("failed to parse protocols port range"),
		},
		{
			input: types.ObjectValueMust(protocolAttributeTypes(), map[string]tfattr.Value{
				attr.Policy: types.StringValue(model.PolicyRestricted),
				attr.Ports:  makeTestSet("80-88"),
			}),
			expected: &model.Protocol{
				Policy: model.PolicyRestricted,
				Ports: []*model.PortRange{
					{Start: 80, End: 88},
				},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {

			protocol, err := convertProtocol(c.input)

			assert.Equal(t, c.expected, protocol)
			if c.expectedErr != nil {
				assert.ErrorContains(t, err, c.expectedErr.Error())
			}

		})
	}

}

func TestConvertPortsRangeToMap(t *testing.T) {
	cases := []struct {
		portsRange []*model.PortRange
		expected   map[int]struct{}
	}{
		{
			portsRange: nil,
			expected:   map[int]struct{}{},
		},
		{
			portsRange: []*model.PortRange{
				{
					Start: 70,
					End:   70,
				},
				{
					Start: 81,
					End:   85,
				},
			},
			expected: map[int]struct{}{
				70: {},
				81: {},
				82: {},
				83: {},
				84: {},
				85: {},
			},
		},
		{
			portsRange: []*model.PortRange{
				{
					Start: 80,
					End:   83,
				},
				{
					Start: 81,
					End:   85,
				},
				{
					Start: 81,
					End:   82,
				},
			},
			expected: map[int]struct{}{
				80: {},
				81: {},
				82: {},
				83: {},
				84: {},
				85: {},
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := convertPortsRangeToMap(c.portsRange)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func makeTestSet(values ...string) types.Set {
	elements := make([]tfattr.Value, 0, len(values))
	for _, val := range values {
		elements = append(elements, types.StringValue(val))
	}

	return types.SetValueMust(types.StringType, elements)
}

func TestEqualPorts(t *testing.T) {
	cases := []struct {
		inputA   types.Set
		inputB   types.Set
		expected bool
	}{
		{
			inputA:   makeTestSet(""),
			inputB:   makeTestSet(""),
			expected: false,
		},
		{
			inputA:   makeTestSet("80"),
			inputB:   makeTestSet(""),
			expected: false,
		},
		{
			inputA:   makeTestSet("80"),
			inputB:   makeTestSet("90"),
			expected: false,
		},
		{
			inputA:   makeTestSet("80"),
			inputB:   makeTestSet("80"),
			expected: true,
		},
		{
			inputA:   makeTestSet("80-81"),
			inputB:   makeTestSet("80", "81"),
			expected: true,
		},
		{
			inputA:   makeTestSet("80-81", "70"),
			inputB:   makeTestSet("70", "80", "81"),
			expected: true,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := equalPorts(c.inputA, c.inputB)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestConvertGroupsAccessToTerraform(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name        string
		groupAccess []model.AccessGroup
		reference   types.Set
		expected    types.Set
	}{
		{
			name: "null security_policy_id in the reference is preserved when the API returns the effective policy",
			groupAccess: []model.AccessGroup{
				{GroupID: "test-group-id", SecurityPolicyID: stringPtr("test-policy-id")},
			},
			reference: makeTestAccessGroupSet(ctx, makeTestAccessGroup(ctx, "test-group-id", types.StringNull())),
			expected:  makeTestAccessGroupSet(ctx, makeTestAccessGroup(ctx, "test-group-id", types.StringNull())),
		},
		{
			name: "empty security_policy_id in the reference is preserved when the API returns the effective policy",
			groupAccess: []model.AccessGroup{
				{GroupID: "test-group-id", SecurityPolicyID: stringPtr("test-policy-id")},
			},
			reference: makeTestAccessGroupSet(ctx, makeTestAccessGroup(ctx, "test-group-id", types.StringValue(""))),
			expected:  makeTestAccessGroupSet(ctx, makeTestAccessGroup(ctx, "test-group-id", types.StringValue(""))),
		},
		{
			name: "security_policy_id set in the reference keeps the API value",
			groupAccess: []model.AccessGroup{
				{GroupID: "test-group-id", SecurityPolicyID: stringPtr("test-policy-id")},
			},
			reference: makeTestAccessGroupSet(ctx, makeTestAccessGroup(ctx, "test-group-id", types.StringValue("test-policy-id"))),
			expected:  makeTestAccessGroupSet(ctx, makeTestAccessGroup(ctx, "test-group-id", types.StringValue("test-policy-id"))),
		},
		{
			name: "group missing from the reference keeps the API value",
			groupAccess: []model.AccessGroup{
				{GroupID: "test-group-id", SecurityPolicyID: stringPtr("test-policy-id")},
			},
			reference: makeObjectsSetNull(ctx, accessGroupAttributeTypes()),
			expected:  makeTestAccessGroupSet(ctx, makeTestAccessGroup(ctx, "test-group-id", types.StringValue("test-policy-id"))),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, diags := convertGroupsAccessToTerraform(ctx, c.groupAccess, c.reference)

			assert.False(t, diags.HasError())
			assert.Equal(t, c.expected, actual)
		})
	}
}

func makeTestAccessGroup(ctx context.Context, groupID string, securityPolicyID types.String) types.Object {
	return types.ObjectValueMust(accessGroupAttributeTypes(), map[string]tfattr.Value{
		attr.GroupID:          types.StringValue(groupID),
		attr.SecurityPolicyID: securityPolicyID,
		attr.AccessPolicy:     makeObjectsSetNull(ctx, accessPolicyAttributeTypes()),
	})
}

func makeTestAccessGroupSet(ctx context.Context, groups ...types.Object) types.Set {
	elements := make([]tfattr.Value, 0, len(groups))
	for _, group := range groups {
		elements = append(elements, group)
	}

	return types.SetValueMust(types.ObjectNull(accessGroupAttributeTypes()).Type(ctx), elements)
}
