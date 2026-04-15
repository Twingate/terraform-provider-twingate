package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapDifference(t *testing.T) {
	cases := []struct {
		mapA     map[string]string
		mapB     map[string]string
		expected map[string]string
	}{
		{
			// Keys in A that don't appear in B survive.
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     map[string]string{"c": "3"},
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			// Overlapping key is removed; non-overlapping key survives.
			mapA:     map[string]string{"a": "1", "b": "2"},
			mapB:     map[string]string{"b": "99"},
			expected: map[string]string{"a": "1"},
		},
		{
			// Full overlap → nil.
			mapA:     map[string]string{"a": "1"},
			mapB:     map[string]string{"a": "99"},
			expected: nil,
		},
		{
			// Nil A → nil.
			mapA:     nil,
			mapB:     map[string]string{"a": "1"},
			expected: nil,
		},
		{
			// Nil B → all of A survives.
			mapA:     map[string]string{"a": "1"},
			mapB:     nil,
			expected: map[string]string{"a": "1"},
		},
		{
			// Both nil → nil.
			mapA:     nil,
			mapB:     nil,
			expected: nil,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			actual := MapDifference(c.mapA, c.mapB)

			assert.Equal(t, c.expected, actual)
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
			actual := MapUnion(c.mapA, c.mapB)

			assert.Equal(t, c.expected, actual)
		})
	}
}
