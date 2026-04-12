package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
