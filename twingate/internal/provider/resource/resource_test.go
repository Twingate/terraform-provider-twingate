package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestResourceResourceReadDiagnosticsError(t *testing.T) {
	t.Run("Test Twingate Resource : Resource Read Diagnostics Error", func(t *testing.T) {
		res := &model.Resource{
			Groups:    []string{},
			Protocols: &model.Protocols{},
		}
		d := &schema.ResourceData{}
		diags := readDiagnostics(d, res)
		assert.True(t, diags.HasError())
	})
}

func TestIntersection(t *testing.T) {
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
