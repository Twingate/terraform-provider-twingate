package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

type ReadDNSFilteringProfile struct {
	DnsFilteringProfile *gqlDNSFilteringProfile `graphql:"dnsFilteringProfile(id: $id)"`
}

type gqlDNSFilteringProfile struct {
	IDName
	Priority               float64
	AllowedDomains         []string
	DeniedDomains          []string
	FallbackMethod         string
	Groups                 gqlGroupIDs `graphql:"groups(after: $groupsEndCursor, first: $pageLimit)"`
	PrivacyCategoryConfig  *PrivacyCategoryConfig
	SecurityCategoryConfig *SecurityCategoryConfig
	ContentCategoryConfig  *ContentCategoryConfig
}

type gqlGroupIDs struct {
	PaginatedResource[*GroupIDEdge]
}

func (g gqlGroupIDs) ToModel() []string {
	return utils.Map[*GroupIDEdge, string](g.Edges, func(edge *GroupIDEdge) string {
		return string(edge.Node.ID)
	})
}

type GroupIDEdge struct {
	Node *gqlGroupID
}

type gqlGroupID struct {
	IDName
}

type PrivacyCategoryConfig struct {
	BlockAffiliate         bool `json:"blockAffiliate"`
	BlockDisguisedTrackers bool `json:"blockDisguisedTrackers"`
	BlockAdsAndTrackers    bool `json:"blockAdsAndTrackers"`
}

type SecurityCategoryConfig struct {
	EnableThreatIntelligenceFeeds   bool `json:"enableThreatIntelligenceFeeds"`
	EnableGoogleSafeBrowsing        bool `json:"enableGoogleSafeBrowsing"`
	BlockCryptojacking              bool `json:"blockCryptojacking"`
	BlockIdnHomographs              bool `json:"blockIdnHomographs"`
	BlockTyposquatting              bool `json:"blockTyposquatting"`
	BlockDnsRebinding               bool `json:"blockDnsRebinding"`
	BlockNewlyRegisteredDomains     bool `json:"blockNewlyRegisteredDomains"`
	BlockDomainGenerationAlgorithms bool `json:"blockDomainGenerationAlgorithms"`
	BlockParkedDomains              bool `json:"blockParkedDomains"`
}

type ContentCategoryConfig struct {
	BlockGambling               bool `json:"blockGambling"`
	BlockDating                 bool `json:"blockDating"`
	BlockAdultContent           bool `json:"blockAdultContent"`
	BlockSocialMedia            bool `json:"blockSocialMedia"`
	BlockGames                  bool `json:"blockGames"`
	BlockStreaming              bool `json:"blockStreaming"`
	BlockPiracy                 bool `json:"blockPiracy"`
	EnableYoutubeRestrictedMode bool `json:"enableYoutubeRestrictedMode"`
	EnableSafeSearch            bool `json:"enableSafeSearch"`
}

func (q ReadDNSFilteringProfile) IsEmpty() bool {
	return q.DnsFilteringProfile == nil
}

func (q ReadDNSFilteringProfile) ToModel() *model.DNSFilteringProfile {
	if q.DnsFilteringProfile == nil {
		return nil
	}

	return q.DnsFilteringProfile.ToModel()
}

func (p gqlDNSFilteringProfile) ToModel() *model.DNSFilteringProfile {
	profile := &model.DNSFilteringProfile{
		ID:             string(p.ID),
		Name:           p.Name,
		Priority:       p.Priority,
		FallbackMethod: p.FallbackMethod,
		AllowedDomains: p.AllowedDomains,
		DeniedDomains:  p.DeniedDomains,
		Groups:         p.Groups.ToModel(),
	}

	if p.PrivacyCategoryConfig != nil {
		profile.PrivacyCategories = &model.PrivacyCategories{
			BlockAffiliate:         p.PrivacyCategoryConfig.BlockAffiliate,
			BlockDisguisedTrackers: p.PrivacyCategoryConfig.BlockDisguisedTrackers,
			BlockAdsAndTrackers:    p.PrivacyCategoryConfig.BlockAdsAndTrackers,
		}
	}

	if p.SecurityCategoryConfig != nil {
		profile.SecurityCategories = &model.SecurityCategory{
			EnableThreatIntelligenceFeeds:   p.SecurityCategoryConfig.EnableThreatIntelligenceFeeds,
			EnableGoogleSafeBrowsing:        p.SecurityCategoryConfig.EnableGoogleSafeBrowsing,
			BlockCryptojacking:              p.SecurityCategoryConfig.BlockCryptojacking,
			BlockIdnHomographs:              p.SecurityCategoryConfig.BlockIdnHomographs,
			BlockTyposquatting:              p.SecurityCategoryConfig.BlockTyposquatting,
			BlockDnsRebinding:               p.SecurityCategoryConfig.BlockDnsRebinding,
			BlockNewlyRegisteredDomains:     p.SecurityCategoryConfig.BlockNewlyRegisteredDomains,
			BlockDomainGenerationAlgorithms: p.SecurityCategoryConfig.BlockDomainGenerationAlgorithms,
			BlockParkedDomains:              p.SecurityCategoryConfig.BlockParkedDomains,
		}
	}

	if p.ContentCategoryConfig != nil {
		profile.ContentCategories = &model.ContentCategory{
			BlockGambling:               p.ContentCategoryConfig.BlockGambling,
			BlockDating:                 p.ContentCategoryConfig.BlockDating,
			BlockAdultContent:           p.ContentCategoryConfig.BlockAdultContent,
			BlockSocialMedia:            p.ContentCategoryConfig.BlockSocialMedia,
			BlockGames:                  p.ContentCategoryConfig.BlockGames,
			BlockStreaming:              p.ContentCategoryConfig.BlockStreaming,
			BlockPiracy:                 p.ContentCategoryConfig.BlockPiracy,
			EnableYoutubeRestrictedMode: p.ContentCategoryConfig.EnableYoutubeRestrictedMode,
			EnableSafeSearch:            p.ContentCategoryConfig.EnableSafeSearch,
		}
	}

	return profile
}

type ReadDNSFilteringProfileGroups struct {
	DnsFilteringProfile *gqlDNSFilteringProfileGroups `graphql:"dnsFilteringProfile(id: $id)"`
}

func (q ReadDNSFilteringProfileGroups) IsEmpty() bool {
	return q.DnsFilteringProfile == nil
}

type gqlDNSFilteringProfileGroups struct {
	IDName
	Groups gqlGroupIDs `graphql:"groups(after: $groupsEndCursor, first: $pageLimit)"`
}
