package resource

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestMapUnion(t *testing.T) {
	cases := []struct {
		mapA     map[string]string
		mapB     map[string]string
		expected map[string]string
	}{
		{
			mapA:     map[string]string{"a": "1", "b": "2", "c": "3"},
			mapB:     map[string]string{"d": "4", "e": "5", "f": "6"},
			expected: map[string]string{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5", "f": "6"},
		},
		{
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     map[string]string{"b": "3", "c": "4"},
			expected: map[string]string{"a": "1", "b": "3", "c": "4"}, // b key is overwritten by mapB's value
		},
		{
			mapA:     nil,
			mapB:     map[string]string{"a": "1", "b": "2"},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     nil,
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			mapA:     nil,
			mapB:     nil,
			expected: nil,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := mapUnion(c.mapA, c.mapB)

			assert.Equal(t, c.expected, actual)
		})
	}
}
