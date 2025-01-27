package models

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDnsFilteringProfileModel(t *testing.T) {
	cases := []struct {
		group model.DNSFilteringProfile

		expectedName string
		expectedID   string
	}{
		{
			group: model.DNSFilteringProfile{},
		},
		{
			group: model.DNSFilteringProfile{
				ID:   "id",
				Name: "name",
			},
			expectedID:   "id",
			expectedName: "name",
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case_%d", n), func(t *testing.T) {
			assert.Equal(t, c.expectedID, c.group.GetID())
			assert.Equal(t, c.expectedName, c.group.GetName())
		})
	}
}
