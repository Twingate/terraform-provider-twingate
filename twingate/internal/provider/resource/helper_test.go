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
