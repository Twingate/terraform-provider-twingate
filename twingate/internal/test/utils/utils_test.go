package utils

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	type testCase struct {
		items    []string
		element  string
		expected bool
	}

	cases := []testCase{
		{
			items:    []string{"1", "2", "3"},
			element:  "2",
			expected: true,
		},
		{
			items:    []string{"1", "2", "3"},
			element:  "0",
			expected: false,
		},
		{
			items:    []string{},
			element:  "0",
			expected: false,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case: %d", n), func(t *testing.T) {
			actual := utils.Contains(c.items, c.element)

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestMapKeys(t *testing.T) {
	type testCase struct {
		lookup   map[string]bool
		expected []string
	}

	cases := []testCase{
		{
			lookup:   map[string]bool{},
			expected: []string{},
		},
		{
			lookup:   map[string]bool{"1": true, "3": true, "2": true},
			expected: []string{"1", "2", "3"},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case: %d", n), func(t *testing.T) {
			actual := utils.MapKeys(c.lookup)

			assert.ElementsMatch(t, c.expected, actual)
		})
	}
}

func TestMakeLookupMap(t *testing.T) {
	type testCase struct {
		items    []string
		expected map[string]bool
	}

	cases := []testCase{
		{
			items:    []string{},
			expected: map[string]bool{},
		},
		{
			items:    []string{"1", "2", "3"},
			expected: map[string]bool{"1": true, "2": true, "3": true},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case: %d", n), func(t *testing.T) {
			actual := utils.MakeLookupMap(c.items)

			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestDocList(t *testing.T) {
	type testCase struct {
		items    []string
		expected string
	}

	cases := []testCase{
		{
			items:    []string{},
			expected: "",
		},
		{
			items:    []string{"1"},
			expected: "1",
		},
		{
			items:    []string{"1", "2"},
			expected: "1 or 2",
		},
		{
			items:    []string{"1", "2", "3"},
			expected: "1, 2 or 3",
		},
		{
			items:    []string{"1", "2", "3", "4"},
			expected: "1, 2, 3 or 4",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case: %d", n), func(t *testing.T) {
			actual := utils.DocList(c.items)

			assert.Equal(t, c.expected, actual)
		})
	}
}
