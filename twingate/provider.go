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
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
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
		apiToken = config.APIToken.String()
	}

	if !config.Network.IsNull() {
		network = config.Network.String()
	}

	if !config.URL.IsNull() {
		url = config.URL.String()
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

func withDefault[T comparable](val, defaultVal T) T {
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

//func Provider(version string) *schema.Provider {
//	provider := &schema.Provider{
//		Schema: providerOptions(),
//		ResourcesMap: map[string]*schema.Resource{
//			resource.TwingateRemoteNetwork: resource.RemoteNetwork(),
//			resource.TwingateConnector:         resource.Connector(),
//			resource.TwingateConnectorTokens:   resource.ConnectorTokens(),
//			resource.TwingateGroup:             resource.Group(),
//			resource.TwingateResource:          resource.Resource(),
//			resource.TwingateServiceAccount:    resource.ServiceAccount(),
//			resource.TwingateServiceAccountKey: resource.ServiceKey(),
//			resource.TwingateUser:              resource.User(),
//		},
//		DataSourcesMap: map[string]*schema.Resource{
//			datasource.TwingateGroup:            datasource.Group(),
//			datasource.TwingateGroups:           datasource.Groups(),
//			datasource.TwingateRemoteNetwork:    datasource.RemoteNetwork(),
//			datasource.TwingateRemoteNetworks:   datasource.RemoteNetworks(),
//			datasource.TwingateUser:             datasource.User(),
//			datasource.TwingateUsers:            datasource.Users(),
//			datasource.TwingateConnector:        datasource.Connector(),
//			datasource.TwingateConnectors:       datasource.Connectors(),
//			datasource.TwingateResource:         datasource.Resource(),
//			datasource.TwingateResources:        datasource.Resources(),
//			datasource.TwingateServiceAccounts:  datasource.ServiceAccounts(),
//			datasource.TwingateSecurityPolicy:   datasource.SecurityPolicy(),
//			datasource.TwingateSecurityPolicies: datasource.SecurityPolicies(),
//		},
//	}
//	provider.ConfigureContextFunc = configure(version, provider)
//
//	return provider
//}

//func providerOptions() map[string]*schema.Schema {
//	return map[string]*schema.Schema{
//		attr.APIToken: {
//			Type:        schema.TypeString,
//			Optional:    true,
//			Sensitive:   true,
//			DefaultFunc: schema.EnvDefaultFunc(EnvAPIToken, nil),
//			Description: fmt.Sprintf("The access key for API operations. You can retrieve this\n"+
//				"from the Twingate Admin Console ([documentation](https://docs.twingate.com/docs/api-overview)).\n"+
//				"Alternatively, this can be specified using the %s environment variable.", EnvAPIToken),
//		},
//		attr.Network: {
//			Type:        schema.TypeString,
//			Optional:    true,
//			Sensitive:   false,
//			DefaultFunc: schema.EnvDefaultFunc(EnvNetwork, nil),
//			Description: fmt.Sprintf("Your Twingate network ID for API operations.\n"+
//				"You can find it in the Admin Console URL, for example:\n"+
//				"`autoco.twingate.com`, where `autoco` is your network ID\n"+
//				"Alternatively, this can be specified using the %s environment variable.", EnvNetwork),
//		},
//		attr.URL: {
//			Type:        schema.TypeString,
//			Optional:    true,
//			Sensitive:   false,
//			DefaultFunc: schema.EnvDefaultFunc(EnvURL, DefaultURL),
//			Description: fmt.Sprintf("The default is '%s'\n"+
//				"This is optional and shouldn't be changed under normal circumstances.", DefaultURL),
//		},
//		attr.HTTPTimeout: {
//			Type:        schema.TypeInt,
//			Optional:    true,
//			DefaultFunc: schema.EnvDefaultFunc(EnvHTTPTimeout, DefaultHTTPTimeout),
//			Description: fmt.Sprintf("Specifies a time limit in seconds for the http requests made. The default value is %s seconds.\n"+
//				"Alternatively, this can be specified using the %s environment variable", DefaultHTTPTimeout, EnvHTTPTimeout),
//		},
//		attr.HTTPMaxRetry: {
//			Type:        schema.TypeInt,
//			Optional:    true,
//			DefaultFunc: schema.EnvDefaultFunc(EnvHTTPMaxRetry, DefaultHTTPMaxRetry),
//			Description: fmt.Sprintf("Specifies a retry limit for the http requests made. The default value is %s.\n"+
//				"Alternatively, this can be specified using the %s environment variable", DefaultHTTPMaxRetry, EnvHTTPMaxRetry),
//		},
//	}
//}

//func configure(version string, _ *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
//	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
//		apiToken := d.Get(attr.APIToken).(string)
//		network := d.Get(attr.Network).(string)
//		url := d.Get(attr.URL).(string)
//		httpTimeout := d.Get(attr.HTTPTimeout).(int)
//		httpMaxRetry := d.Get(attr.HTTPMaxRetry).(int)
//
//		if network != "" {
//			return client.NewClient(url,
//					apiToken,
//					network,
//					time.Duration(httpTimeout)*time.Second,
//					httpMaxRetry,
//					version),
//				nil
//		}
//
//		return nil, diag.Diagnostics{
//			diag.Diagnostic{
//				Severity: diag.Error,
//				Summary:  "Unable to create Twingate client",
//				Detail:   "Unable to create anonymous Twingate client, network has to be provided",
//			},
//		}
//	}
//}
