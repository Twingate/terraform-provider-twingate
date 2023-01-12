package attrs

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/stretchr/testify/assert"
)

func TestAttrLen(t *testing.T) {
	cases := []struct {
		attributes []string

		expected string
	}{
		{
			attributes: nil,
			expected:   "",
		},
		{
			attributes: []string{"key"},
			expected:   "key.#",
		},
		{
			attributes: []string{"access", "key"},
			expected:   "access.0.key.#",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expected, attr.Len(c.attributes...))
		})
	}
}
