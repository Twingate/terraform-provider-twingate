package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &dnsFilteringProfile{}

func NewDNSFilteringProfileDatasource() datasource.DataSource {
	return &dnsFilteringProfile{}
}

type dnsFilteringProfile struct {
	client *client.Client
}

type dnsFilteringProfileModel struct {
	ID                 types.String             `tfsdk:"id"`
	Name               types.String             `tfsdk:"name"`
	Priority           types.Float64            `tfsdk:"priority"`
	FallbackMethod     types.String             `tfsdk:"fallback_method"`
	Groups             types.Set                `tfsdk:"groups"`
	AllowedDomains     *domainsModel            `tfsdk:"allowed_domains"`
	DeniedDomains      *domainsModel            `tfsdk:"denied_domains"`
	ContentCategories  *contentCategoriesModel  `tfsdk:"content_categories"`
	SecurityCategories *securityCategoriesModel `tfsdk:"security_categories"`
	PrivacyCategories  *privacyCategoriesModel  `tfsdk:"privacy_categories"`
}

type domainsModel struct {
	Domains types.Set `tfsdk:"domains"`
}

type privacyCategoriesModel struct {
	BlockAffiliateLinks    types.Bool `tfsdk:"block_affiliate_links"`
	BlockDisguisedTrackers types.Bool `tfsdk:"block_disguised_trackers"`
	BlockAdsAndTrackers    types.Bool `tfsdk:"block_ads_and_trackers"`
}

type securityCategoriesModel struct {
	EnableThreatIntelligenceFeeds   types.Bool `tfsdk:"enable_threat_intelligence_feeds"`
	EnableGoogleSafeBrowsing        types.Bool `tfsdk:"enable_google_safe_browsing"`
	BlockCryptojacking              types.Bool `tfsdk:"block_cryptojacking"`
	BlockIdnHomoglyph               types.Bool `tfsdk:"block_idn_homoglyph"`
	BlockTyposquatting              types.Bool `tfsdk:"block_typosquatting"`
	BlockDNSRebinding               types.Bool `tfsdk:"block_dns_rebinding"`
	BlockNewlyRegisteredDomains     types.Bool `tfsdk:"block_newly_registered_domains"`
	BlockDomainGenerationAlgorithms types.Bool `tfsdk:"block_domain_generation_algorithms"`
	BlockParkedDomains              types.Bool `tfsdk:"block_parked_domains"`
}

type contentCategoriesModel struct {
	BlockGambling               types.Bool `tfsdk:"block_gambling"`
	BlockDating                 types.Bool `tfsdk:"block_dating"`
	BlockAdultContent           types.Bool `tfsdk:"block_adult_content"`
	BlockSocialMedia            types.Bool `tfsdk:"block_social_media"`
	BlockGames                  types.Bool `tfsdk:"block_games"`
	BlockStreaming              types.Bool `tfsdk:"block_streaming"`
	BlockPiracy                 types.Bool `tfsdk:"block_piracy"`
	EnableYoutubeRestrictedMode types.Bool `tfsdk:"enable_youtube_restricted_mode"`
	EnableSafesearch            types.Bool `tfsdk:"enable_safesearch"`
}

func (d *dnsFilteringProfile) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateDNSFilteringProfile
}

func (d *dnsFilteringProfile) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *dnsFilteringProfile) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) { //nolint:funlen
	resp.Schema = schema.Schema{
		Description: "DNS filtering gives you the ability to control what websites your users can access. DNS filtering is only available on certain plans. For more information, see Twingate's [documentation](https://www.twingate.com/docs/dns-filtering). DNS filtering must be enabled for this data source to work. If DNS filtering isn't enabled, the provider will throw an error.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The DNS filtering profile's ID.",
			},
			// computed
			attr.Name: schema.StringAttribute{
				Computed:    true,
				Description: "The DNS filtering profile's name.",
			},
			attr.Priority: schema.Float64Attribute{
				Computed:    true,
				Description: "A floating point number representing the profile's priority.",
			},
			attr.FallbackMethod: schema.StringAttribute{
				Computed:    true,
				Description: "The DNS filtering profile's fallback method. One of AUTOMATIC or STRICT.",
			},
			attr.Groups: schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A set of group IDs that have this as their DNS filtering profile. Defaults to an empty set.",
			},
		},

		Blocks: map[string]schema.Block{
			attr.AllowedDomains: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.Domains: schema.SetAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "A set of allowed domains.",
					},
				},
			},
			attr.DeniedDomains: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.Domains: schema.SetAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "A set of denied domains.",
					},
				},
			},

			attr.PrivacyCategories: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.BlockAffiliateLinks: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block affiliate links.",
					},
					attr.BlockDisguisedTrackers: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block disguised third party trackers.",
					},
					attr.BlockAdsAndTrackers: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block ads and trackers.",
					},
				},
			},

			attr.SecurityCategories: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.EnableThreatIntelligenceFeeds: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to filter content using threat intelligence feeds.",
					},
					attr.EnableGoogleSafeBrowsing: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to use Google Safe browsing lists to block content.",
					},
					attr.BlockCryptojacking: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block cryptojacking sites.",
					},
					attr.BlockIdnHomoglyph: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block homoglyph attacks.",
					},
					attr.BlockTyposquatting: schema.BoolAttribute{
						Computed:    true,
						Description: "Blocks typosquatted domains.",
					},
					attr.BlockDNSRebinding: schema.BoolAttribute{
						Computed:    true,
						Description: "Blocks public DNS entries from returning private IP addresses.",
					},
					attr.BlockNewlyRegisteredDomains: schema.BoolAttribute{
						Computed:    true,
						Description: "Blocks newly registered domains.",
					},
					attr.BlockDomainGenerationAlgorithms: schema.BoolAttribute{
						Computed:    true,
						Description: "Blocks DGA domains.",
					},
					attr.BlockParkedDomains: schema.BoolAttribute{
						Computed:    true,
						Description: "Block parked domains.",
					},
				},
			},

			attr.ContentCategories: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.BlockGambling: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block gambling content.",
					},
					attr.BlockDating: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block dating content.",
					},
					attr.BlockAdultContent: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block adult content.",
					},
					attr.BlockSocialMedia: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block social media.",
					},
					attr.BlockGames: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block games.",
					},
					attr.BlockStreaming: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block streaming content.",
					},
					attr.BlockPiracy: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to block piracy sites.",
					},
					attr.EnableYoutubeRestrictedMode: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to force YouTube to use restricted mode.",
					},
					attr.EnableSafesearch: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to force safe search.",
					},
				},
			},
		},
	}
}

func (d *dnsFilteringProfile) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) { //nolint:funlen
	var data dnsFilteringProfileModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := d.client.ReadDNSFilteringProfile(ctx, data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateDNSFilteringProfile)

		return
	}

	data.Name = types.StringValue(profile.Name)
	data.Priority = types.Float64Value(profile.Priority)
	data.FallbackMethod = types.StringValue(profile.FallbackMethod)
	data.AllowedDomains = convertDomainsToTerraform(profile.AllowedDomains)
	data.DeniedDomains = convertDomainsToTerraform(profile.DeniedDomains)
	data.Groups = utils.MakeStringSet(profile.Groups)

	if profile.PrivacyCategories != nil {
		data.PrivacyCategories = &privacyCategoriesModel{
			BlockAffiliateLinks:    types.BoolValue(profile.PrivacyCategories.BlockAffiliate),
			BlockDisguisedTrackers: types.BoolValue(profile.PrivacyCategories.BlockDisguisedTrackers),
			BlockAdsAndTrackers:    types.BoolValue(profile.PrivacyCategories.BlockAdsAndTrackers),
		}
	}

	if profile.ContentCategories != nil {
		data.ContentCategories = &contentCategoriesModel{
			BlockGambling:               types.BoolValue(profile.ContentCategories.BlockGambling),
			BlockDating:                 types.BoolValue(profile.ContentCategories.BlockDating),
			BlockAdultContent:           types.BoolValue(profile.ContentCategories.BlockAdultContent),
			BlockSocialMedia:            types.BoolValue(profile.ContentCategories.BlockSocialMedia),
			BlockGames:                  types.BoolValue(profile.ContentCategories.BlockGames),
			BlockStreaming:              types.BoolValue(profile.ContentCategories.BlockStreaming),
			BlockPiracy:                 types.BoolValue(profile.ContentCategories.BlockPiracy),
			EnableYoutubeRestrictedMode: types.BoolValue(profile.ContentCategories.EnableYoutubeRestrictedMode),
			EnableSafesearch:            types.BoolValue(profile.ContentCategories.EnableSafeSearch),
		}
	}

	if profile.SecurityCategories != nil {
		data.SecurityCategories = &securityCategoriesModel{
			EnableThreatIntelligenceFeeds:   types.BoolValue(profile.SecurityCategories.EnableThreatIntelligenceFeeds),
			EnableGoogleSafeBrowsing:        types.BoolValue(profile.SecurityCategories.EnableGoogleSafeBrowsing),
			BlockCryptojacking:              types.BoolValue(profile.SecurityCategories.BlockCryptojacking),
			BlockIdnHomoglyph:               types.BoolValue(profile.SecurityCategories.BlockIdnHomographs),
			BlockTyposquatting:              types.BoolValue(profile.SecurityCategories.BlockTyposquatting),
			BlockDNSRebinding:               types.BoolValue(profile.SecurityCategories.BlockDNSRebinding),
			BlockNewlyRegisteredDomains:     types.BoolValue(profile.SecurityCategories.BlockNewlyRegisteredDomains),
			BlockDomainGenerationAlgorithms: types.BoolValue(profile.SecurityCategories.BlockDomainGenerationAlgorithms),
			BlockParkedDomains:              types.BoolValue(profile.SecurityCategories.BlockParkedDomains),
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
