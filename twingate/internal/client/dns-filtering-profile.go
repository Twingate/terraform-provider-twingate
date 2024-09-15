package client

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client/query"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

func (client *Client) ReadShallowDNSFilteringProfiles(ctx context.Context) ([]*model.DNSFilteringProfile, error) {
	opr := resourceDNSFilteringProfile.read().withCustomName("readShallowDNSFilteringProfiles")

	response := query.ReadDNSFilteringProfiles{}
	if err := client.query(ctx, &response, nil, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	return response.ToModel(), nil
}

func (client *Client) ReadDNSFilteringProfile(ctx context.Context, profileID string) (*model.DNSFilteringProfile, error) {
	opr := resourceDNSFilteringProfile.read()

	if profileID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(profileID),
		cursor(query.CursorGroups),
		pageLimit(defaultPageLimit),
	)

	response := query.ReadDNSFilteringProfile{}
	if err := client.queryWithTimeout(ctx, &response, variables, opr, attr{id: profileID}); err != nil {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.DNSFilteringProfile.Groups.FetchPages(oprCtx, client.readDNSFilteringProfileGroupsAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.ToModel(), nil
}

func (client *Client) CreateDNSFilteringProfile(ctx context.Context, name string) (*model.DNSFilteringProfile, error) {
	opr := resourceDNSFilteringProfile.create()

	if name == "" {
		return nil, opr.apiError(ErrGraphqlNameIsEmpty)
	}

	variables := newVars(
		gqlVar(name, "name"),
		cursor(query.CursorGroups),
		pageLimit(defaultPageLimit),
	)

	var response query.CreateDNSFilteringProfile
	if err := client.mutate(ctx, &response, variables, opr, attr{name: name}); err != nil {
		return nil, err
	}

	return response.Entity.ToModel(), nil
}

type PrivacyCategoryConfigInput struct {
	BlockAdsAndTrackers    bool `json:"blockAdsAndTrackers"`
	BlockAffiliate         bool `json:"blockAffiliate"`
	BlockDisguisedTrackers bool `json:"blockDisguisedTrackers"`
}

func newPrivacyCategoryConfigInput(input *model.PrivacyCategories) *PrivacyCategoryConfigInput {
	if input == nil {
		return nil
	}

	return &PrivacyCategoryConfigInput{
		BlockAdsAndTrackers:    input.BlockAdsAndTrackers,
		BlockAffiliate:         input.BlockAffiliate,
		BlockDisguisedTrackers: input.BlockDisguisedTrackers,
	}
}

type SecurityCategoryConfigInput struct {
	BlockCryptojacking              bool `json:"blockCryptojacking"`
	BlockDNSRebinding               bool `json:"blockDnsRebinding"`
	BlockDomainGenerationAlgorithms bool `json:"blockDomainGenerationAlgorithms"`
	BlockIdnHomographs              bool `json:"blockIdnHomographs"`
	BlockNewlyRegisteredDomains     bool `json:"blockNewlyRegisteredDomains"`
	BlockParkedDomains              bool `json:"blockParkedDomains"`
	BlockTyposquatting              bool `json:"blockTyposquatting"`
	EnableGoogleSafeBrowsing        bool `json:"enableGoogleSafeBrowsing"`
	EnableThreatIntelligenceFeeds   bool `json:"enableThreatIntelligenceFeeds"`
}

func newSecurityCategoryConfigInput(input *model.SecurityCategory) *SecurityCategoryConfigInput {
	if input == nil {
		return nil
	}

	return &SecurityCategoryConfigInput{
		BlockCryptojacking:              input.BlockCryptojacking,
		BlockDNSRebinding:               input.BlockDNSRebinding,
		BlockDomainGenerationAlgorithms: input.BlockDomainGenerationAlgorithms,
		BlockIdnHomographs:              input.BlockIdnHomographs,
		BlockNewlyRegisteredDomains:     input.BlockNewlyRegisteredDomains,
		BlockParkedDomains:              input.BlockParkedDomains,
		BlockTyposquatting:              input.BlockTyposquatting,
		EnableGoogleSafeBrowsing:        input.EnableGoogleSafeBrowsing,
		EnableThreatIntelligenceFeeds:   input.EnableThreatIntelligenceFeeds,
	}
}

type ContentCategoryConfigInput struct {
	BlockAdultContent           bool `json:"blockAdultContent"`
	BlockDating                 bool `json:"blockDating"`
	BlockGambling               bool `json:"blockGambling"`
	BlockGames                  bool `json:"blockGames"`
	BlockPiracy                 bool `json:"blockPiracy"`
	BlockSocialMedia            bool `json:"blockSocialMedia"`
	BlockStreaming              bool `json:"blockStreaming"`
	EnableSafeSearch            bool `json:"enableSafeSearch"`
	EnableYoutubeRestrictedMode bool `json:"enableYoutubeRestrictedMode"`
}

func newContentCategoryConfigInput(input *model.ContentCategory) *ContentCategoryConfigInput {
	if input == nil {
		return nil
	}

	return &ContentCategoryConfigInput{
		BlockAdultContent:           input.BlockAdultContent,
		BlockDating:                 input.BlockDating,
		BlockGambling:               input.BlockGambling,
		BlockGames:                  input.BlockGames,
		BlockPiracy:                 input.BlockPiracy,
		BlockSocialMedia:            input.BlockSocialMedia,
		BlockStreaming:              input.BlockStreaming,
		EnableSafeSearch:            input.EnableSafeSearch,
		EnableYoutubeRestrictedMode: input.EnableYoutubeRestrictedMode,
	}
}

type DohFallbackMethod string

func (client *Client) UpdateDNSFilteringProfile(ctx context.Context, input *model.DNSFilteringProfile) (*model.DNSFilteringProfile, error) {
	opr := resourceDNSFilteringProfile.update()

	if input == nil || input.ID == "" {
		return nil, opr.apiError(ErrGraphqlIDIsEmpty)
	}

	variables := newVars(
		gqlID(input.ID, "id"),
		gqlNullable(input.Name, "name"),
		gqlNullable(input.Priority, "priority"),
		gqlVar(input.AllowedDomains, "allowedDomains"),
		gqlVar(input.DeniedDomains, "deniedDomains"),
		gqlVar(DohFallbackMethod(input.FallbackMethod), "fallbackMethod"),
		gqlVar(input.Groups, "groups"),
		gqlVar(newPrivacyCategoryConfigInput(input.PrivacyCategories), "privacyCategoryConfig"),
		gqlVar(newSecurityCategoryConfigInput(input.SecurityCategories), "securityCategoryConfig"),
		gqlVar(newContentCategoryConfigInput(input.ContentCategories), "contentCategoryConfig"),
		cursor(query.CursorGroups),
		pageLimit(defaultPageLimit),
	)

	var response query.UpdateDNSFilteringProfile
	if err := client.mutate(ctx, &response, variables, opr, attr{id: input.ID}); err != nil {
		return nil, err
	}

	oprCtx := withOperationCtx(ctx, opr)

	if err := response.Entity.Groups.FetchPages(oprCtx, client.readDNSFilteringProfileGroupsAfter, variables); err != nil {
		return nil, err //nolint
	}

	return response.Entity.ToModel(), nil
}

func (client *Client) readDNSFilteringProfileGroupsAfter(ctx context.Context, variables map[string]interface{}, cursor string) (*query.PaginatedResource[*query.GroupIDEdge], error) {
	opr := resourceGroup.read().withCustomName("readDNSFilteringProfileGroupsAfter")

	variables[query.CursorGroups] = cursor

	response := query.ReadDNSFilteringProfileGroups{}
	if err := client.query(ctx, &response, variables, opr, attr{id: "All"}); err != nil {
		return nil, err
	}

	return &response.DNSFilteringProfile.Groups.PaginatedResource, nil
}

func (client *Client) DeleteDNSFilteringProfile(ctx context.Context, profileID string) error {
	opr := resourceDNSFilteringProfile.delete()

	if profileID == "" {
		return opr.apiError(ErrGraphqlIDIsEmpty)
	}

	var response query.DeleteDNSFilteringProfile

	return client.mutate(ctx, &response, newVars(gqlID(profileID)), opr, attr{id: profileID})
}
