package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestServiceAccountKeyModel(t *testing.T) {
	cases := []struct {
		key model.ServiceKey

		expectedName   string
		expectedID     string
		expectedActive bool
	}{
		{
			key: model.ServiceKey{},
		},
		{
			key: model.ServiceKey{
				ID:     "id",
				Name:   "name",
				Status: model.StatusActive,
			},
			expectedID:     "id",
			expectedName:   "name",
			expectedActive: true,
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expectedID, c.key.GetID())
			assert.Equal(t, c.expectedName, c.key.GetName())
			assert.Equal(t, c.expectedActive, c.key.IsActive())
		})
	}
}
