package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

			assert.Equal(t, c.expected, isZeroValue(c.val))
		})
	}
}

func TestGetDefaultGQLValue(t *testing.T) {
	var (
		defaultString *string
		defaultInt    *int
		defaultBool   *bool
		defaultFloat  *float64
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

			assert.Equal(t, c.expected, getNullableValue(c.val))
		})
	}
}
