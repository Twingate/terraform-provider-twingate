package attr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstAttr(t *testing.T) {
	cases := []struct {
		name       string
		attributes []string
		expected   string
	}{
		{
			name:       "No attributes provided",
			attributes: []string{},
			expected:   "",
		},
		{
			name:       "PathAttr returns valid string",
			attributes: []string{"attr1", "attr2"},
			expected:   "attr1.attr2.0",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, FirstAttr(c.attributes...))
		})
	}
}

func TestLenAttr(t *testing.T) {
	cases := []struct {
		name       string
		attributes []string
		expected   string
	}{
		{
			name:       "No attributes provided",
			attributes: []string{},
			expected:   "",
		},
		{
			name:       "PathAttr returns valid string",
			attributes: []string{"attr1", "attr2"},
			expected:   "attr1.attr2.#",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, LenAttr(c.attributes...))
		})
	}
}
