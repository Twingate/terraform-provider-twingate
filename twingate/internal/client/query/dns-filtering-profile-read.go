package query

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

type ReadDNSFilteringProfile struct {
	DnsFilteringProfile *gqlDNSFilteringProfile `graphql:"dnsFilteringProfile(id: $id)"`
}

type gqlDNSFilteringProfile struct {
	IDName
	Priority               float64                 `json:"priority"`
	AllowedDomains         []string                `json:"allowedDomains"`
	DeniedDomains          []string                `json:"deniedDomains"`
	FallbackMethod         string                  `json:"fallbackMethod"`
	PrivacyCategories      *PrivacyCategoryConfig  `json:"privacyCategoryConfig"`
	SecurityCategoryConfig *SecurityCategoryConfig `json:"securityCategoryConfig"`
	ContentCategoryConfig  *ContentCategoryConfig  `json:"contentCategoryConfig"`
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
	}

	if p.PrivacyCategories != nil {
		profile.PrivacyCategories = &model.PrivacyCategories{
			BlockAffiliate:         p.PrivacyCategories.BlockAffiliate,
			BlockDisguisedTrackers: p.PrivacyCategories.BlockDisguisedTrackers,
			BlockAdsAndTrackers:    p.PrivacyCategories.BlockAdsAndTrackers,
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
