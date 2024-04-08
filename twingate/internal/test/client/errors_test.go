package client

import (
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestUnwrapError(t *testing.T) {
	err := client.NewAPIError(errBadRequest, "read", "resource")

	assert.Equal(t, errBadRequest, err.Unwrap())
}
