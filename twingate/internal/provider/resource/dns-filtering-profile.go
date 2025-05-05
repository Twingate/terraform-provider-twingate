package resource

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ resource.Resource = &dnsFilteringProfile{}
var _ resource.ResourceWithImportState = &dnsFilteringProfile{}

func NewDNSFilteringProfile() resource.Resource {
	return &dnsFilteringProfile{}
}

type dnsFilteringProfile struct {
	client *client.Client
}

type dnsFilteringProfileModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Priority           types.Float64 `tfsdk:"priority"`
	FallbackMethod     types.String  `tfsdk:"fallback_method"`
	Groups             types.Set     `tfsdk:"groups"`
	AllowedDomains     types.Object  `tfsdk:"allowed_domains"`
	DeniedDomains      types.Object  `tfsdk:"denied_domains"`
	ContentCategories  types.Object  `tfsdk:"content_categories"`
	SecurityCategories types.Object  `tfsdk:"security_categories"`
	PrivacyCategories  types.Object  `tfsdk:"privacy_categories"`
}

func (r *dnsFilteringProfile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateDNSFilteringProfile
}

func (r *dnsFilteringProfile) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

func (r *dnsFilteringProfile) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(attr.ID), req, resp)

	profile, err := r.client.ReadDNSFilteringProfile(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("failed to import state", err.Error())

		return
	}

	resp.State.SetAttribute(ctx, path.Root(attr.FallbackMethod), types.StringValue(profile.FallbackMethod))
	resp.State.SetAttribute(ctx, path.Root(attr.Groups), convertStringListToSet(profile.Groups))

	if len(profile.AllowedDomains) > 0 {
		resp.State.SetAttribute(ctx, path.Root(attr.AllowedDomains), convertDomainsToTerraform(profile.AllowedDomains, nil))
	}

	if len(profile.DeniedDomains) > 0 {
		resp.State.SetAttribute(ctx, path.Root(attr.DeniedDomains), convertDomainsToTerraform(profile.DeniedDomains, nil))
	}

	if profile.ContentCategories != nil {
		resp.State.SetAttribute(ctx, path.Root(attr.ContentCategories), convertContentCategoriesToTerraform(profile.ContentCategories))
	}

	if profile.SecurityCategories != nil {
		resp.State.SetAttribute(ctx, path.Root(attr.SecurityCategories), convertSecurityCategoriesToTerraform(profile.SecurityCategories))
	}

	if profile.PrivacyCategories != nil {
		resp.State.SetAttribute(ctx, path.Root(attr.PrivacyCategories), convertPrivacyCategoriesToTerraform(profile.PrivacyCategories))
	}
}

func (r *dnsFilteringProfile) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) { //nolint
	resp.Schema = schema.Schema{
		Description: "DNS filtering gives you the ability to control what websites your users can access. DNS filtering is only available on certain plans. For more information, see Twingate's [documentation](https://www.twingate.com/docs/dns-filtering). DNS filtering must be enabled for this resources to work. If DNS filtering isn't enabled, the provider will throw an error.",
		Attributes: map[string]schema.Attribute{
			attr.Name: schema.StringAttribute{
				Required:    true,
				Description: "The DNS filtering profile's name.",
			},
			attr.Priority: schema.Float64Attribute{
				Required:    true,
				Description: "A floating point number representing the profile's priority.",
			},
			// optional
			attr.FallbackMethod: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `The DNS filtering profile's fallback method. One of "AUTO" or "STRICT". Defaults to "STRICT".`,
				Default:     stringdefault.StaticString(model.FallbackMethodStrict),
				Validators: []validator.String{
					stringvalidator.OneOf(model.FallbackMethods...),
				},
			},
			attr.Groups: schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "A set of group IDs that have this as their DNS filtering profile. Defaults to an empty set.",
				Default:     setdefault.StaticValue(defaultEmptySet()),
			},

			// computed
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated ID of the DNS filtering profile.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},

		Blocks: map[string]schema.Block{
			attr.AllowedDomains: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.IsAuthoritative: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether Terraform should override changes made outside of Terraform. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.Domains: schema.SetAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Description: "A set of allowed domains. Defaults to an empty set.",
						Default:     setdefault.StaticValue(defaultEmptySet()),
					},
				},
			},

			attr.DeniedDomains: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.IsAuthoritative: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether Terraform should override changes made outside of Terraform. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.Domains: schema.SetAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Description: "A set of denied domains. Defaults to an empty set.",
						Default:     setdefault.StaticValue(defaultEmptySet()),
					},
				},
			},

			//nolint:dupl
			attr.ContentCategories: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.BlockGambling: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block gambling content. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.BlockDating: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block dating content. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.BlockAdultContent: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block adult content. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.BlockSocialMedia: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block social media. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.BlockGames: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block games. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.BlockStreaming: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block streaming content. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.BlockPiracy: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block piracy sites. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.EnableYoutubeRestrictedMode: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to force YouTube to use restricted mode. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.EnableSafesearch: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to force safe search. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
				},
			},

			//nolint:dupl
			attr.SecurityCategories: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.EnableThreatIntelligenceFeeds: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to filter content using threat intelligence feeds. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.EnableGoogleSafeBrowsing: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to use Google Safe browsing lists to block content. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.BlockCryptojacking: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block cryptojacking sites. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.BlockIdnHomoglyph: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block homoglyph attacks. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.BlockTyposquatting: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Blocks typosquatted domains. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.BlockDNSRebinding: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Blocks public DNS entries from returning private IP addresses. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.BlockNewlyRegisteredDomains: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Blocks newly registered domains. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.BlockDomainGenerationAlgorithms: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Blocks DGA domains. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
					attr.BlockParkedDomains: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Block parked domains. Defaults to true.",
						Default:     booldefault.StaticBool(true),
					},
				},
			},

			attr.PrivacyCategories: schema.SingleNestedBlock{
				Description: "A block with the following attributes.",
				Attributes: map[string]schema.Attribute{
					attr.BlockAffiliateLinks: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block affiliate links. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.BlockDisguisedTrackers: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block disguised third party trackers. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
					attr.BlockAdsAndTrackers: schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to block ads and trackers. Defaults to false.",
						Default:     booldefault.StaticBool(false),
					},
				},
			},
		},
	}
}

func defaultEmptySet() types.Set {
	return types.SetValueMust(types.StringType, []tfattr.Value{})
}

func (r *dnsFilteringProfile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsFilteringProfileModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := r.client.CreateDNSFilteringProfile(ctx, plan.Name.ValueString())

	if profile != nil {
		if !plan.Priority.IsNull() {
			profile.Priority = plan.Priority.ValueFloat64()
		}

		profile.FallbackMethod = plan.FallbackMethod.ValueString()
		profile.Groups = convertSetToList(plan.Groups)
		profile.AllowedDomains = convertDomains(plan.AllowedDomains)
		profile.DeniedDomains = convertDomains(plan.DeniedDomains)
		profile.PrivacyCategories = convertPrivacyCategories(plan.PrivacyCategories)
		profile.ContentCategories = convertContentCategories(plan.ContentCategories)
		profile.SecurityCategories = convertSecurityCategories(plan.SecurityCategories)

		profile, err = r.client.UpdateDNSFilteringProfile(ctx, profile)
	}

	r.helper(ctx, profile, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func (r *dnsFilteringProfile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsFilteringProfileModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := r.client.ReadDNSFilteringProfile(ctx, state.ID.ValueString())

	if profile != nil {
		if !convertBoolDefaultTrue(state.AllowedDomains.Attributes()[attr.IsAuthoritative]) {
			profile.AllowedDomains = convertDomains(state.AllowedDomains)
		}

		if !convertBoolDefaultTrue(state.DeniedDomains.Attributes()[attr.IsAuthoritative]) {
			profile.DeniedDomains = convertDomains(state.DeniedDomains)
		}
	}

	r.helper(ctx, profile, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

func (r *dnsFilteringProfile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) { //nolint:funlen
	var state, plan dnsFilteringProfileModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	allowedDomains := convertDomains(plan.AllowedDomains)
	deniedDomains := convertDomains(plan.DeniedDomains)

	profile := &model.DNSFilteringProfile{
		ID:                 state.ID.ValueString(),
		Name:               plan.Name.ValueString(),
		Priority:           plan.Priority.ValueFloat64(),
		FallbackMethod:     plan.FallbackMethod.ValueString(),
		Groups:             convertSetToList(plan.Groups),
		AllowedDomains:     allowedDomains,
		DeniedDomains:      deniedDomains,
		PrivacyCategories:  convertPrivacyCategories(plan.PrivacyCategories),
		ContentCategories:  convertContentCategories(plan.ContentCategories),
		SecurityCategories: convertSecurityCategories(plan.SecurityCategories),
	}

	allowedDomainsIsAuthoritative := convertBoolDefaultTrue(plan.AllowedDomains.Attributes()[attr.IsAuthoritative])
	deniedDomainsIsAuthoritative := convertBoolDefaultTrue(plan.DeniedDomains.Attributes()[attr.IsAuthoritative])

	var (
		originAllowedDomains []string
		originDeniedDomains  []string
	)

	if !allowedDomainsIsAuthoritative || !deniedDomainsIsAuthoritative {
		origin, err := r.client.ReadDNSFilteringProfile(ctx, profile.ID)
		if err != nil {
			r.helper(ctx, profile, &plan, &resp.State, &resp.Diagnostics, err, operationUpdate)

			return
		}

		originAllowedDomains = origin.AllowedDomains
		originDeniedDomains = origin.AllowedDomains
	}

	if !allowedDomainsIsAuthoritative {
		profile.AllowedDomains = setUnion(profile.AllowedDomains, originAllowedDomains)
	}

	if !deniedDomainsIsAuthoritative {
		profile.DeniedDomains = setUnion(profile.DeniedDomains, originDeniedDomains)
	}

	var err error
	profile, err = r.client.UpdateDNSFilteringProfile(ctx, profile)

	if profile != nil {
		if !allowedDomainsIsAuthoritative {
			profile.AllowedDomains = allowedDomains
		}

		if !deniedDomainsIsAuthoritative {
			profile.DeniedDomains = deniedDomains
		}
	}

	r.helper(ctx, profile, &plan, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func (r *dnsFilteringProfile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsFilteringProfileModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSFilteringProfile(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationDelete, TwingateDNSFilteringProfile)
}

func (r *dnsFilteringProfile) helper(ctx context.Context, profile *model.DNSFilteringProfile, state *dnsFilteringProfileModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			respState.RemoveResource(ctx)

			return
		}

		addErr(diagnostics, err, operation, TwingateDNSFilteringProfile)

		return
	}

	state.ID = types.StringValue(profile.ID)
	state.Name = types.StringValue(profile.Name)
	state.Priority = types.Float64Value(profile.Priority)
	state.FallbackMethod = types.StringValue(profile.FallbackMethod)
	state.Groups = convertStringListToSet(profile.Groups)

	if !state.AllowedDomains.IsNull() {
		state.AllowedDomains = convertDomainsToTerraform(profile.AllowedDomains, state.AllowedDomains.Attributes()[attr.IsAuthoritative])
	}

	if !state.DeniedDomains.IsNull() {
		state.DeniedDomains = convertDomainsToTerraform(profile.DeniedDomains, state.DeniedDomains.Attributes()[attr.IsAuthoritative])
	}

	if !state.ContentCategories.IsNull() && profile.ContentCategories != nil {
		state.ContentCategories = convertContentCategoriesToTerraform(profile.ContentCategories)
	}

	if !state.SecurityCategories.IsNull() && profile.SecurityCategories != nil {
		state.SecurityCategories = convertSecurityCategoriesToTerraform(profile.SecurityCategories)
	}

	if !state.PrivacyCategories.IsNull() && profile.PrivacyCategories != nil {
		state.PrivacyCategories = convertPrivacyCategoriesToTerraform(profile.PrivacyCategories)
	}

	// Set refreshed state
	diags := respState.Set(ctx, state)
	diagnostics.Append(diags...)
}

func convertContentCategoriesToTerraform(categories *model.ContentCategory) types.Object {
	attributes := map[string]tfattr.Value{
		attr.BlockGambling:               types.BoolValue(categories.BlockGambling),
		attr.BlockDating:                 types.BoolValue(categories.BlockDating),
		attr.BlockAdultContent:           types.BoolValue(categories.BlockAdultContent),
		attr.BlockSocialMedia:            types.BoolValue(categories.BlockSocialMedia),
		attr.BlockGames:                  types.BoolValue(categories.BlockGames),
		attr.BlockStreaming:              types.BoolValue(categories.BlockStreaming),
		attr.BlockPiracy:                 types.BoolValue(categories.BlockPiracy),
		attr.EnableYoutubeRestrictedMode: types.BoolValue(categories.EnableYoutubeRestrictedMode),
		attr.EnableSafesearch:            types.BoolValue(categories.EnableSafeSearch),
	}

	return types.ObjectValueMust(contentCategoriesAttributeTypes(), attributes)
}

func contentCategoriesAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.BlockGambling:               types.BoolType,
		attr.BlockDating:                 types.BoolType,
		attr.BlockAdultContent:           types.BoolType,
		attr.BlockSocialMedia:            types.BoolType,
		attr.BlockGames:                  types.BoolType,
		attr.BlockStreaming:              types.BoolType,
		attr.BlockPiracy:                 types.BoolType,
		attr.EnableYoutubeRestrictedMode: types.BoolType,
		attr.EnableSafesearch:            types.BoolType,
	}
}

func convertSecurityCategoriesToTerraform(categories *model.SecurityCategory) types.Object {
	attributes := map[string]tfattr.Value{
		attr.EnableThreatIntelligenceFeeds:   types.BoolValue(categories.EnableThreatIntelligenceFeeds),
		attr.EnableGoogleSafeBrowsing:        types.BoolValue(categories.EnableGoogleSafeBrowsing),
		attr.BlockCryptojacking:              types.BoolValue(categories.BlockCryptojacking),
		attr.BlockIdnHomoglyph:               types.BoolValue(categories.BlockIdnHomographs),
		attr.BlockTyposquatting:              types.BoolValue(categories.BlockTyposquatting),
		attr.BlockDNSRebinding:               types.BoolValue(categories.BlockDNSRebinding),
		attr.BlockNewlyRegisteredDomains:     types.BoolValue(categories.BlockNewlyRegisteredDomains),
		attr.BlockDomainGenerationAlgorithms: types.BoolValue(categories.BlockDomainGenerationAlgorithms),
		attr.BlockParkedDomains:              types.BoolValue(categories.BlockParkedDomains),
	}

	return types.ObjectValueMust(securityCategoriesAttributeTypes(), attributes)
}

func securityCategoriesAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.EnableThreatIntelligenceFeeds:   types.BoolType,
		attr.EnableGoogleSafeBrowsing:        types.BoolType,
		attr.BlockCryptojacking:              types.BoolType,
		attr.BlockIdnHomoglyph:               types.BoolType,
		attr.BlockTyposquatting:              types.BoolType,
		attr.BlockDNSRebinding:               types.BoolType,
		attr.BlockNewlyRegisteredDomains:     types.BoolType,
		attr.BlockDomainGenerationAlgorithms: types.BoolType,
		attr.BlockParkedDomains:              types.BoolType,
	}
}

func convertPrivacyCategoriesToTerraform(categories *model.PrivacyCategories) types.Object {
	attributes := map[string]tfattr.Value{
		attr.BlockAffiliateLinks:    types.BoolValue(categories.BlockAffiliate),
		attr.BlockDisguisedTrackers: types.BoolValue(categories.BlockDisguisedTrackers),
		attr.BlockAdsAndTrackers:    types.BoolValue(categories.BlockAdsAndTrackers),
	}

	return types.ObjectValueMust(privacyCategoriesAttributeTypes(), attributes)
}

func privacyCategoriesAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.BlockAffiliateLinks:    types.BoolType,
		attr.BlockDisguisedTrackers: types.BoolType,
		attr.BlockAdsAndTrackers:    types.BoolType,
	}
}

func convertDomainsToTerraform(domains []string, isAuthoritative tfattr.Value) types.Object {
	authoritative := types.BoolValue(true)
	if isAuthoritative != nil {
		authoritative = isAuthoritative.(types.Bool)
	}

	attributes := map[string]tfattr.Value{
		attr.IsAuthoritative: authoritative,
		attr.Domains:         convertStringListToSet(domains),
	}

	return types.ObjectValueMust(domainsAttributeTypes(), attributes)
}

func domainsAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.IsAuthoritative: types.BoolType,
		attr.Domains: types.SetType{
			ElemType: types.StringType,
		},
	}
}

func convertPrivacyCategories(obj types.Object) *model.PrivacyCategories {
	attrs := obj.Attributes()

	return &model.PrivacyCategories{
		BlockAffiliate:         convertBoolDefaultFalse(attrs[attr.BlockAffiliateLinks]),
		BlockDisguisedTrackers: convertBoolDefaultFalse(attrs[attr.BlockDisguisedTrackers]),
		BlockAdsAndTrackers:    convertBoolDefaultFalse(attrs[attr.BlockAdsAndTrackers]),
	}
}

func convertSecurityCategories(obj types.Object) *model.SecurityCategory {
	attrs := obj.Attributes()

	return &model.SecurityCategory{
		EnableThreatIntelligenceFeeds:   convertBoolDefaultTrue(attrs[attr.EnableThreatIntelligenceFeeds]),
		EnableGoogleSafeBrowsing:        convertBoolDefaultTrue(attrs[attr.EnableGoogleSafeBrowsing]),
		BlockCryptojacking:              convertBoolDefaultTrue(attrs[attr.BlockCryptojacking]),
		BlockIdnHomographs:              convertBoolDefaultTrue(attrs[attr.BlockIdnHomoglyph]),
		BlockTyposquatting:              convertBoolDefaultTrue(attrs[attr.BlockTyposquatting]),
		BlockDNSRebinding:               convertBoolDefaultTrue(attrs[attr.BlockDNSRebinding]),
		BlockNewlyRegisteredDomains:     convertBoolDefaultTrue(attrs[attr.BlockNewlyRegisteredDomains]),
		BlockDomainGenerationAlgorithms: convertBoolDefaultTrue(attrs[attr.BlockDomainGenerationAlgorithms]),
		BlockParkedDomains:              convertBoolDefaultTrue(attrs[attr.BlockParkedDomains]),
	}
}

func convertContentCategories(obj types.Object) *model.ContentCategory {
	attrs := obj.Attributes()

	return &model.ContentCategory{
		BlockGambling:               convertBoolDefaultFalse(attrs[attr.BlockGambling]),
		BlockDating:                 convertBoolDefaultFalse(attrs[attr.BlockDating]),
		BlockAdultContent:           convertBoolDefaultFalse(attrs[attr.BlockAdultContent]),
		BlockSocialMedia:            convertBoolDefaultFalse(attrs[attr.BlockSocialMedia]),
		BlockGames:                  convertBoolDefaultFalse(attrs[attr.BlockGames]),
		BlockStreaming:              convertBoolDefaultFalse(attrs[attr.BlockStreaming]),
		BlockPiracy:                 convertBoolDefaultFalse(attrs[attr.BlockPiracy]),
		EnableYoutubeRestrictedMode: convertBoolDefaultFalse(attrs[attr.EnableYoutubeRestrictedMode]),
		EnableSafeSearch:            convertBoolDefaultFalse(attrs[attr.EnableSafesearch]),
	}
}

func convertBoolDefaultTrue(boolAttr tfattr.Value) bool {
	return convertBoolWithDefault(boolAttr, true)
}

func convertBoolDefaultFalse(boolAttr tfattr.Value) bool {
	return convertBoolWithDefault(boolAttr, false)
}

func convertBoolWithDefault(value tfattr.Value, defaultValue bool) bool {
	if value == nil || value.IsNull() || value.IsUnknown() {
		return defaultValue
	}

	return boolAttr(value)
}

func boolAttr(boolAttr tfattr.Value) bool {
	return boolAttr.(types.Bool).ValueBool()
}

func convertDomains(obj types.Object) []string {
	if obj.IsNull() || obj.IsUnknown() {
		return []string{}
	}

	return convertSetToList(obj.Attributes()[attr.Domains].(types.Set))
}

func convertSetToList(set types.Set) []string {
	return utils.Map(set.Elements(),
		func(item tfattr.Value) string {
			return item.(types.String).ValueString()
		},
	)
}

func convertStringListToSet(items []string) types.Set {
	values := utils.Map(items, func(item string) tfattr.Value {
		return types.StringValue(item)
	})

	return types.SetValueMust(types.StringType, values)
}
