package client

import (
	"fmt"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

func TestIsZeroValue(t *testing.T) {
	var (
		boolPointer    *bool
		strPointer     *string
		intPointer     *int
		int32Pointer   *int32
		int64Pointer   *int64
		float64Pointer *float64
		float32Pointer *float32
	)

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
		{
			val:      boolPointer,
			expected: true,
		},
		{
			val:      strPointer,
			expected: true,
		},
		{
			val:      intPointer,
			expected: true,
		},
		{
			val:      int32Pointer,
			expected: true,
		},
		{
			val:      int64Pointer,
			expected: true,
		},
		{
			val:      float32Pointer,
			expected: true,
		},
		{
			val:      float64Pointer,
			expected: true,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, isZeroValue(c.val))
		})
	}
}

func TestGetNullableValue(t *testing.T) {
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

func TestGqlID(t *testing.T) {
	cases := []struct {
		inputVal   interface{}
		inputNames []string
		expected   map[string]interface{}
	}{
		{
			inputVal: "test-id",
			expected: map[string]interface{}{
				"id": graphql.ID("test-id"),
			},
		},
		{
			inputVal:   "custom-id",
			inputNames: []string{"custom"},
			expected: map[string]interface{}{
				"custom": graphql.ID("custom-id"),
			},
		},
		{
			inputVal: graphql.ID("gql"),
			expected: map[string]interface{}{
				"id": graphql.ID("gql"),
			},
		},
		{
			inputVal: 101,
			expected: map[string]interface{}{
				"id": graphql.ID("101"),
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			values := make(map[string]interface{})
			gqlID(c.inputVal, c.inputNames...)(values)

			assert.Equal(t, c.expected, values)
		})
	}
}

func TestGqlNullableID(t *testing.T) {
	var defaultID *graphql.ID

	cases := []struct {
		inputVal  interface{}
		inputName string
		expected  map[string]interface{}
	}{
		{
			inputVal:  "test-id",
			inputName: "id",
			expected: map[string]interface{}{
				"id": graphql.ID("test-id"),
			},
		},
		{
			inputVal:  "",
			inputName: "custom",
			expected: map[string]interface{}{
				"custom": defaultID,
			},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {
			values := make(map[string]interface{})
			gqlNullableID(c.inputVal, c.inputName)(values)

			assert.Equal(t, c.expected, values)
		})
	}
}

func TestGetValue(t *testing.T) {
	var (
		strVal             = "str"
		boolTrue           = true
		boolFalse          = false
		intVal     int     = 1
		int32Val   int32   = 1
		int64Val   int64   = 1111
		float32Val float32 = 1.1
		float64Val float64 = 9999.99
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
			val:      &strVal,
			expected: strVal,
		},
		{
			val:      &boolTrue,
			expected: boolTrue,
		},
		{
			val:      &boolFalse,
			expected: boolFalse,
		},
		{
			val:      &intVal,
			expected: intVal,
		},
		{
			val:      &int32Val,
			expected: int32Val,
		},
		{
			val:      &int64Val,
			expected: int64Val,
		},
		{
			val:      &float32Val,
			expected: float32Val,
		},
		{
			val:      &float64Val,
			expected: float64Val,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_n%d", n), func(t *testing.T) {

			assert.Equal(t, c.expected, getValue(c.val))
		})
	}
}
