package twingate

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/datasources"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultHTTPTimeout  = "10"
	DefaultHTTPMaxRetry = "10"
	DefaultURL          = "twingate.com"

	// EnvAPIToken env var for Token.
	EnvAPIToken     = "TWINGATE_API_TOKEN" //#nosec
	EnvNetwork      = "TWINGATE_NETWORK"
	EnvURL          = "TWINGATE_URL"
	EnvHTTPTimeout  = "TWINGATE_HTTP_TIMEOUT"
	EnvHTTPMaxRetry = "TWINGATE_HTTP_MAX_RETRY"
)

var _ provider.Provider = &Twingate{}

type Twingate struct {
	version string
}

type twingateProviderModel struct {
	APIToken     types.String `tfsdk:"api_token"`
	Network      types.String `tfsdk:"network"`
	URL          types.String `tfsdk:"url"`
	HTTPTimeout  types.Int64  `tfsdk:"http_timeout"`
	HTTPMaxRetry types.Int64  `tfsdk:"http_max_retry"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Twingate{
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

	var httpTimeout, httpMaxRetry int

	if val, err := strconv.Atoi(withDefault(os.Getenv(EnvHTTPTimeout), DefaultHTTPTimeout)); err == nil {
		httpTimeout = val
	}

	if val, err := strconv.Atoi(withDefault(os.Getenv(EnvHTTPMaxRetry), DefaultHTTPMaxRetry)); err == nil {
		httpMaxRetry = val
	}

	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	}

	if !config.Network.IsNull() {
		network = config.Network.ValueString()
	}

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.HTTPTimeout.IsNull() {
		httpTimeout = int(config.HTTPTimeout.ValueInt64())
	}

	if !config.HTTPMaxRetry.IsNull() {
		httpMaxRetry = int(config.HTTPMaxRetry.ValueInt64())
	}

	if network == "" {
		response.Diagnostics.AddAttributeError(
			path.Root(attr.Network),
			fmt.Sprintf("Missing Twingate %s", attr.Network),
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
		t.version)

	response.DataSourceData = client
	response.ResourceData = client
}

func withDefault[T comparable](val, defaultVal T) T { //nolint:ireturn
	var zeroValue T
	if val == zeroValue {
		return defaultVal
	}

	return val
}

func (t Twingate) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewConnectorDatasource,
		datasources.NewConnectorsDatasource,
		datasources.NewGroupDatasource,
		datasources.NewGroupsDatasource,
		datasources.NewRemoteNetworkDatasource,
		datasources.NewRemoteNetworksDatasource,
		datasources.NewUserDatasource,
		datasources.NewUsersDatasource,
		datasources.NewServiceAccountsDatasource,
		datasources.NewSecurityPolicyDatasource,
		datasources.NewSecurityPoliciesDatasource,
		datasources.NewResourceDatasource,
		datasources.NewResourcesDatasource,
	}
}

func (t Twingate) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewConnectorResource,
		resources.NewRemoteNetworkResource,
		resources.NewConnectorTokensResource,
		resources.NewGroupResource,
		resources.NewServiceAccountResource,
		resources.NewServiceKeyResource,
		resources.NewUserResource,
		resources.NewResourceResource,
	}
}
