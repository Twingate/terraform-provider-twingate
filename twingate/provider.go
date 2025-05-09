package twingate

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	twingateDatasource "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/datasource"
	twingateResource "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
					attr.GroupsEnabled: schema.BoolAttribute{
						Optional:    true,
						Description: fmt.Sprintf("Specifies whether the provider should cache groups. The default value is `%t`.", true),
					},
				},
			},
		},
	}
}

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

	client := client.NewClient(url,
		apiToken,
		network,
		time.Duration(httpTimeout)*time.Second,
		httpMaxRetry,
		t.agent,
		t.version,
		getCacheOptions(config.Cache))

	response.DataSourceData = client
	response.ResourceData = client

	policy, _ := client.ReadSecurityPolicy(ctx, "", twingateResource.DefaultSecurityPolicyName)
	if policy != nil {
		twingateResource.DefaultSecurityPolicyID = policy.ID
	}
}

func getCacheOptions(config types.Object) client.CacheOptions {
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

	return client.CacheOptions{
		ResourceEnabled: resourceEnabled,
		GroupsEnabled:   groupsEnabled,
	}
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
