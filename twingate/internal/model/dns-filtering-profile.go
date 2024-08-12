package model

const (
	FallbackMethodAuto   = "AUTO"
	FallbackMethodStrict = "STRICT"
)

var FallbackMethods = []string{FallbackMethodAuto, FallbackMethodStrict} //nolint

type DNSFilteringProfile struct {
	ID                 string
	Name               string
	AllowedDomains     []string
	DeniedDomains      []string
	Groups             []string
	FallbackMethod     string
	Priority           float64
	PrivacyCategories  *PrivacyCategories
	SecurityCategories *SecurityCategory
	ContentCategories  *ContentCategory
}

type PrivacyCategories struct {
	BlockAffiliate         bool
	BlockDisguisedTrackers bool
	BlockAdsAndTrackers    bool
}

type SecurityCategory struct {
	EnableThreatIntelligenceFeeds   bool
	EnableGoogleSafeBrowsing        bool
	BlockCryptojacking              bool
	BlockIdnHomographs              bool
	BlockTyposquatting              bool
	BlockDNSRebinding               bool
	BlockNewlyRegisteredDomains     bool
	BlockDomainGenerationAlgorithms bool
	BlockParkedDomains              bool
}

type ContentCategory struct {
	BlockGambling               bool
	BlockDating                 bool
	BlockAdultContent           bool
	BlockSocialMedia            bool
	BlockGames                  bool
	BlockStreaming              bool
	BlockPiracy                 bool
	EnableYoutubeRestrictedMode bool
	EnableSafeSearch            bool
}

func (p DNSFilteringProfile) GetName() string {
	return p.Name
}

func (p DNSFilteringProfile) GetID() string {
	return p.ID
}
