package twingate

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	twingateDatasource "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/datasource"
	twingateResource "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultHTTPTimeout     = "35"
	DefaultHTTPMaxRetry    = "10"
	DefaultURL             = "twingate.com"
	defaultResourceEnabled = true
	defaultGroupsEnabled   = true

	// EnvAPIToken env var for Token.
	EnvAPIToken     = "TWINGATE_API_TOKEN" // #nosec G101
	EnvNetwork      = "TWINGATE_NETWORK"
	EnvURL          = "TWINGATE_URL"
	EnvHTTPTimeout  = "TWINGATE_HTTP_TIMEOUT"
	EnvHTTPMaxRetry = "TWINGATE_HTTP_MAX_RETRY"
)

var _ provider.Provider = &Twingate{}

type Twingate struct {
	agent   string
	version string
}

type twingateProviderModel struct {
	APIToken     types.String `tfsdk:"api_token"`
	Network      types.String `tfsdk:"network"`
	URL          types.String `tfsdk:"url"`
	HTTPTimeout  types.Int64  `tfsdk:"http_timeout"`
	HTTPMaxRetry types.Int64  `tfsdk:"http_max_retry"`
	Cache        types.Object `tfsdk:"cache"`
	DefaultTags  types.Object `tfsdk:"default_tags"`
}

func New(agent, version string) func() provider.Provider {
	return func() provider.Provider {
		return &Twingate{
			agent:   agent,
			version: version,
		}
	}
}

func (t Twingate) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "twingate"
	response.Version = t.version
}

//nolint:funlen
func (t Twingate) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			attr.APIToken: schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: fmt.Sprintf("The access key for API operations. You can retrieve this\n"+
					"from the Twingate Admin Console ([documentation](https://docs.twingate.com/docs/api-overview)).\n"+
					"Alternatively, this can be specified using the %s environment variable.", EnvAPIToken),
			},
			attr.Network: schema.StringAttribute{
				Optional: true,
				Description: fmt.Sprintf("Your Twingate network ID for API operations.\n"+
					"You can find it in the Admin Console URL, for example:\n"+
					"`autoco.twingate.com`, where `autoco` is your network ID\n"+
					"Alternatively, this can be specified using the %s environment variable.", EnvNetwork),
			},
			attr.URL: schema.StringAttribute{
				Optional: true,
				Description: fmt.Sprintf("The default is '%s'\n"+
					"This is optional and shouldn't be changed under normal circumstances.", DefaultURL),
			},
			attr.HTTPTimeout: schema.Int64Attribute{
				Optional: true,
				Description: fmt.Sprintf("Specifies a time limit in seconds for the http requests made. The default value is %s seconds.\n"+
					"Alternatively, this can be specified using the %s environment variable", DefaultHTTPTimeout, EnvHTTPTimeout),
			},
			attr.HTTPMaxRetry: schema.Int64Attribute{
				Optional: true,
				Description: fmt.Sprintf("Specifies a retry limit for the http requests made. The default value is %s.\n"+
					"Alternatively, this can be specified using the %s environment variable", DefaultHTTPMaxRetry, EnvHTTPMaxRetry),
			},
			attr.Cache: schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Specifies the cache settings for the provider.",
				Attributes: map[string]schema.Attribute{
					attr.ResourceEnabled: schema.BoolAttribute{
						Optional:    true,
						Description: fmt.Sprintf("Specifies whether the provider should cache resources. The default value is `%t`.", true),
					},
					attr.ResourcesFilter: schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Specifies the filter for the resources to be cached.",
						Attributes: map[string]schema.Attribute{
							attr.Name: schema.StringAttribute{
								Optional:    true,
								Description: "Returns only resources that exactly match this name. If no options are passed it will return all resources. Only one option can be used at a time.",
							},
							attr.Name + attr.FilterByRegexp: schema.StringAttribute{
								Optional:    true,
								Description: "The regular expression match of the name of the resource.",
							},
							attr.Name + attr.FilterByContains: schema.StringAttribute{
								Optional:    true,
								Description: "Match when the value exist in the name of the resource.",
							},
							attr.Name + attr.FilterByExclude: schema.StringAttribute{
								Optional:    true,
								Description: "Match when the exact value does not exist in the name of the resource.",
							},
							attr.Name + attr.FilterByPrefix: schema.StringAttribute{
								Optional:    true,
								Description: "The name of the resource must start with the value.",
							},
							attr.Name + attr.FilterBySuffix: schema.StringAttribute{
								Optional:    true,
								Description: "The name of the resource must end with the value.",
							},
							attr.Tags: schema.MapAttribute{
								ElementType: types.StringType,
								Optional:    true,
								Description: "Returns only resources that exactly match the given tags.",
							},
						},
					},
					attr.GroupsEnabled: schema.BoolAttribute{
						Optional:    true,
						Description: fmt.Sprintf("Specifies whether the provider should cache groups. The default value is `%t`.", true),
					},
					attr.GroupsFilter: schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Specifies the filter for the groups to be cached.",
						Attributes: map[string]schema.Attribute{
							attr.Name: schema.StringAttribute{
								Optional:    true,
								Description: "Returns only groups that exactly match this name. If no options are passed it will return all resources. Only one option can be used at a time.",
							},
							attr.Name + attr.FilterByRegexp: schema.StringAttribute{
								Optional:    true,
								Description: "The regular expression match of the name of the group.",
							},
							attr.Name + attr.FilterByContains: schema.StringAttribute{
								Optional:    true,
								Description: "Match when the value exist in the name of the group.",
							},
							attr.Name + attr.FilterByExclude: schema.StringAttribute{
								Optional:    true,
								Description: "Match when the exact value does not exist in the name of the group.",
							},
							attr.Name + attr.FilterByPrefix: schema.StringAttribute{
								Optional:    true,
								Description: "The name of the group must start with the value.",
							},
							attr.Name + attr.FilterBySuffix: schema.StringAttribute{
								Optional:    true,
								Description: "The name of the group must end with the value.",
							},
							attr.IsActive: schema.BoolAttribute{
								Optional:    true,
								Description: "Returns only Groups matching the specified state.",
							},
							attr.Types: schema.SetAttribute{
								Optional:    true,
								ElementType: types.StringType,
								Description: fmt.Sprintf("Returns groups that match a list of types. valid types: `%s`, `%s`, `%s`.", model.GroupTypeManual, model.GroupTypeSynced, model.GroupTypeSystem),
								Validators: []validator.Set{
									setvalidator.ValueStringsAre(stringvalidator.OneOf(model.GroupTypeManual, model.GroupTypeSynced, model.GroupTypeSystem)),
								},
							},
						},
					},
				},
			},
			attr.DefaultTags: schema.SingleNestedAttribute{
				Optional:    true,
				Description: "A default set of tags applied globally to all resources created by the provider.",
				Attributes: map[string]schema.Attribute{
					attr.Tags: schema.MapAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "A map of key-value pair tags to be set on all resources by default.",
					},
				},
			},
		},
	}
}

//nolint:funlen
func (t Twingate) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config twingateProviderModel

	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)

	if response.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	apiToken := os.Getenv(EnvAPIToken)
	network := os.Getenv(EnvNetwork)
	url := withDefault(os.Getenv(EnvURL), DefaultURL)
	httpTimeout := mustGetInt(withDefault(os.Getenv(EnvHTTPTimeout), DefaultHTTPTimeout))
	httpMaxRetry := mustGetInt(withDefault(os.Getenv(EnvHTTPMaxRetry), DefaultHTTPMaxRetry))

	apiToken = overrideStrWithConfig(config.APIToken, apiToken)
	network = overrideStrWithConfig(config.Network, network)
	url = overrideStrWithConfig(config.URL, url)
	httpTimeout = overrideIntWithConfig(config.HTTPTimeout, httpTimeout)
	httpMaxRetry = overrideIntWithConfig(config.HTTPMaxRetry, httpMaxRetry)

	if network == "" {
		response.Diagnostics.AddAttributeError(
			path.Root(attr.Network),
			"Missing Twingate "+attr.Network,
			fmt.Sprintf("The provider cannot create the Twingate API client as there is a missing or empty value for the Twingate %s. "+
				"Set the %s value in the configuration or use the %s environment variable. "+
				"If either is already set, ensure the value is not empty.", attr.Network, attr.Network, EnvNetwork),
		)

		return
	}

	cacheOpts, err := getCacheOptions(config.Cache)
	if err != nil {
		response.Diagnostics.AddAttributeError(
			path.Root(attr.Cache),
			"Issue in configuring Twingate "+attr.Cache,
			fmt.Sprintf("Error: %v", err.Error()),
		)

		return
	}

	client := client.NewClient(url,
		apiToken,
		network,
		time.Duration(httpTimeout)*time.Second,
		httpMaxRetry,
		t.agent,
		t.version,
		cacheOpts)

	response.DataSourceData = client
	response.ResourceData = client
	response.EphemeralResourceData = client

	policy, _ := client.ReadSecurityPolicy(ctx, "", twingateResource.DefaultSecurityPolicyName)
	if policy != nil {
		twingateResource.DefaultSecurityPolicyID = policy.ID
	}

	twingateResource.DefaultTags = getDefaultTags(config.DefaultTags)
}

func getCacheOptions(config types.Object) (client.CacheOptions, error) {
	var (
		resourceEnabled = defaultResourceEnabled
		groupsEnabled   = defaultGroupsEnabled
	)

	if !config.IsNull() && !config.IsUnknown() {
		cacheAttrs := config.Attributes()
		resourceEnabledAttr := cacheAttrs[attr.ResourceEnabled].(types.Bool).ValueBoolPointer()

		if resourceEnabledAttr != nil {
			resourceEnabled = *resourceEnabledAttr
		}

		groupsEnabledAttr := cacheAttrs[attr.GroupsEnabled].(types.Bool).ValueBoolPointer()
		if groupsEnabledAttr != nil {
			groupsEnabled = *groupsEnabledAttr
		}
	}

	resourcesFilter, err := parseResourcesFilter(config)
	if err != nil {
		return client.CacheOptions{}, fmt.Errorf("failed to parse resources filter: %w", err)
	}

	groupsFilter, err := parseGroupFilter(config)
	if err != nil {
		return client.CacheOptions{}, fmt.Errorf("failed to parse groups filter: %w", err)
	}

	return client.CacheOptions{
		ResourceEnabled: resourceEnabled,
		GroupsEnabled:   groupsEnabled,
		ResourcesFilter: resourcesFilter,
		GroupsFilter:    groupsFilter,
	}, nil
}

func parseResourcesFilter(config types.Object) (*model.ResourcesFilter, error) {
	if config.IsNull() || config.IsUnknown() {
		//nolint:nilnil
		return nil, nil
	}

	filterObj := config.Attributes()[attr.ResourcesFilter].(types.Object)
	if filterObj.IsNull() || filterObj.IsUnknown() {
		//nolint:nilnil
		return nil, nil
	}

	attrs := filterObj.Attributes()

	name := attrs[attr.Name].(types.String)
	nameRegexp := attrs[attr.Name+attr.FilterByRegexp].(types.String)
	nameContains := attrs[attr.Name+attr.FilterByContains].(types.String)
	nameExclude := attrs[attr.Name+attr.FilterByExclude].(types.String)
	namePrefix := attrs[attr.Name+attr.FilterByPrefix].(types.String)
	nameSuffix := attrs[attr.Name+attr.FilterBySuffix].(types.String)

	value, filter := twingateDatasource.GetNameFilter(name, nameRegexp, nameContains, nameExclude, namePrefix, nameSuffix)

	if twingateDatasource.CountOptionalAttributes(name, nameRegexp, nameContains, nameExclude, namePrefix, nameSuffix) > 1 {
		return nil, twingateDatasource.ErrResourcesDatasourceShouldSetOneOptionalNameAttribute
	}

	tags := attrs[attr.Tags].(types.Map)

	return &model.ResourcesFilter{
		Name:       &value,
		NameFilter: filter,
		Tags:       twingateDatasource.GetTags(tags),
	}, nil
}

func parseGroupFilter(config types.Object) (*model.GroupsFilter, error) {
	if config.IsNull() || config.IsUnknown() {
		//nolint:nilnil
		return nil, nil
	}

	filterObj := config.Attributes()[attr.GroupsFilter].(types.Object)
	if filterObj.IsNull() || filterObj.IsUnknown() {
		//nolint:nilnil
		return nil, nil
	}

	attrs := filterObj.Attributes()

	name := attrs[attr.Name].(types.String)
	nameRegexp := attrs[attr.Name+attr.FilterByRegexp].(types.String)
	nameContains := attrs[attr.Name+attr.FilterByContains].(types.String)
	nameExclude := attrs[attr.Name+attr.FilterByExclude].(types.String)
	namePrefix := attrs[attr.Name+attr.FilterByPrefix].(types.String)
	nameSuffix := attrs[attr.Name+attr.FilterBySuffix].(types.String)

	value, filter := twingateDatasource.GetNameFilter(name, nameRegexp, nameContains, nameExclude, namePrefix, nameSuffix)

	if twingateDatasource.CountOptionalAttributes(name, nameRegexp, nameContains, nameExclude, namePrefix, nameSuffix) > 1 {
		return nil, twingateDatasource.ErrResourcesDatasourceShouldSetOneOptionalNameAttribute
	}

	groupFilter := &model.GroupsFilter{
		Name:       &value,
		NameFilter: filter,
		IsActive:   attrs[attr.IsActive].(types.Bool).ValueBoolPointer(),
	}

	groupTypes := attrs[attr.Types].(types.Set).Elements()

	if len(groupTypes) > 0 {
		groupFilter.Types = utils.Map(groupTypes, func(item tfattr.Value) string {
			return item.(types.String).ValueString()
		})
	}

	if groupFilter.Name == nil && len(groupFilter.Types) == 0 && groupFilter.IsActive == nil {
		//nolint:nilnil
		return nil, nil
	}

	return groupFilter, nil
}

func mustGetInt(str string) int {
	if val, err := strconv.Atoi(str); err == nil {
		return val
	}

	return 0
}

func overrideStrWithConfig(cfg types.String, defaultValue string) string {
	if !cfg.IsNull() {
		return cfg.ValueString()
	}

	return defaultValue
}

func overrideIntWithConfig(cfg types.Int64, defaultValue int) int {
	if !cfg.IsNull() {
		return int(cfg.ValueInt64())
	}

	return defaultValue
}

func withDefault[T comparable](val, defaultVal T) T {
	var zeroValue T
	if val == zeroValue {
		return defaultVal
	}

	return val
}

func getDefaultTags(defaultTags types.Object) map[string]string {
	if defaultTags.IsNull() || defaultTags.IsUnknown() {
		return nil
	}

	rawTags := defaultTags.Attributes()[attr.Tags].(types.Map)
	if rawTags.IsNull() || rawTags.IsUnknown() || len(rawTags.Elements()) == 0 {
		return nil
	}

	tags := make(map[string]string, len(rawTags.Elements()))

	for key, val := range rawTags.Elements() {
		tags[key] = val.(types.String).ValueString()
	}

	return tags
}

func (t Twingate) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		twingateDatasource.NewConnectorDatasource,
		twingateDatasource.NewConnectorsDatasource,
		twingateDatasource.NewGroupDatasource,
		twingateDatasource.NewGroupsDatasource,
		twingateDatasource.NewRemoteNetworkDatasource,
		twingateDatasource.NewRemoteNetworksDatasource,
		twingateDatasource.NewServiceAccountsDatasource,
		twingateDatasource.NewUserDatasource,
		twingateDatasource.NewUsersDatasource,
		twingateDatasource.NewSecurityPolicyDatasource,
		twingateDatasource.NewSecurityPoliciesDatasource,
		twingateDatasource.NewResourceDatasource,
		twingateDatasource.NewResourcesDatasource,
		twingateDatasource.NewDNSFilteringProfileDatasource,
	}
}

func (t Twingate) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		twingateResource.NewConnectorTokensResource,
		twingateResource.NewConnectorResource,
		twingateResource.NewGroupResource,
		twingateResource.NewRemoteNetworkResource,
		twingateResource.NewServiceAccountResource,
		twingateResource.NewServiceKeyResource,
		twingateResource.NewUserResource,
		twingateResource.NewResourceResource,
		twingateResource.NewDNSFilteringProfile,
	}
}

func (t Twingate) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		func() ephemeral.EphemeralResource {
			return twingateResource.NewEphemeralConnectorTokens()
		},
	}
}
