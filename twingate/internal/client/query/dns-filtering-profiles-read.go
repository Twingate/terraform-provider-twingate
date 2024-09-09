package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

type ReadDNSFilteringProfiles struct {
	DNSFilteringProfiles []*gqlShallowDNSFilteringProfile `graphql:"dnsFilteringProfiles"`
}

type gqlShallowDNSFilteringProfile struct {
	IDName
	Priority float64
}

func (q ReadDNSFilteringProfiles) IsEmpty() bool {
	return len(q.DNSFilteringProfiles) == 0
}

func (q ReadDNSFilteringProfiles) ToModel() []*model.DNSFilteringProfile {
	return utils.Map(q.DNSFilteringProfiles, func(item *gqlShallowDNSFilteringProfile) *model.DNSFilteringProfile {
		return &model.DNSFilteringProfile{
			ID:       string(item.ID),
			Name:     item.Name,
			Priority: item.Priority,
		}
	})
}
