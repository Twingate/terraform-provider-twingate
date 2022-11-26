package resource

import (
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
