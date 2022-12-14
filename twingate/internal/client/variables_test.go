package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func Test_convertToGQL(t *testing.T) {
	cases := []struct {
		val      interface{}
		expected interface{}
	}{
		{
			val:      nil,
			expected: nil,
		},
		{
			val:      "123",
			expected: graphql.String("123"),
		},
		{
			val:      101,
			expected: graphql.Int(101),
		},
		{
			val:      int32(102),
			expected: graphql.Int(102),
		},
		{
			val:      int64(103),
			expected: graphql.Int(103),
		},
		{
			val:      101.5,
			expected: graphql.Float(101.5),
		},
		{
			val:      float32(101.25),
			expected: graphql.Float(101.25),
		},
		{
			val:      true,
			expected: graphql.Boolean(true),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("test case #%d", i+1), func(t *testing.T) {
			actual := convertToGQL(c.val)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestIsDefaultValue(t *testing.T) {
	cases := []struct {
		val      interface{}
		expected bool
	}{
		{
			val:      nil,
			expected: true,
		},
		{
			val:      "",
			expected: true,
		},
		{
			val:      "a",
			expected: false,
		},
		{
			val:      0,
			expected: true,
		},
		{
			val:      1,
			expected: false,
		},
		{
			val:      int32(0),
			expected: true,
		},
		{
			val:      int32(1),
			expected: false,
		},
		{
			val:      int64(0),
			expected: true,
		},
		{
			val:      int64(1),
			expected: false,
		},
		{
			val:      false,
			expected: true,
		},
		{
			val:      true,
			expected: false,
		},
		{
			val:      float64(0),
			expected: true,
		},
		{
			val:      float64(1),
			expected: false,
		},
		{
			val:      float32(0),
			expected: true,
		},
		{
			val:      float32(1),
			expected: false,
		},
		{
			val:      []interface{}{},
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, isDefaultValue(c.val))
		})
	}
}

func TestGetDefaultGQLValue(t *testing.T) {
	var (
		defaultString *graphql.String
		defaultInt    *graphql.Int
		defaultBool   *graphql.Boolean
		defaultFloat  *graphql.Float
	)

	cases := []struct {
		val      interface{}
		expected interface{}
	}{
		{
			val:      nil,
			expected: nil,
		},
		{
			val:      "str",
			expected: defaultString,
		},
		{
			val:      true,
			expected: defaultBool,
		},
		{
			val:      1,
			expected: defaultInt,
		},
		{
			val:      int32(1),
			expected: defaultInt,
		},
		{
			val:      int64(1),
			expected: defaultInt,
		},
		{
			val:      float32(1.0),
			expected: defaultFloat,
		},
		{
			val:      1.0,
			expected: defaultFloat,
		},
		{
			val:      []interface{}{},
			expected: nil,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, getDefaultGQLValue(c.val))
		})
	}
}
