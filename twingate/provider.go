package twingate

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	DefaultHTTPTimeout  = "10"
	DefaultHTTPMaxRetry = "10"
)

func Provider(version string) *schema.Provider {
	provider := &schema.Provider{
		Schema: providerOptions(),
		ResourcesMap: map[string]*schema.Resource{
			"twingate_remote_network":   resourceRemoteNetwork(),
			"twingate_connector":        resourceConnector(),
			"twingate_connector_tokens": resourceConnectorTokens(),
			"twingate_group":            resourceGroup(),
			"twingate_resource":         resourceResource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"twingate_group":          datasourceGroup(),
			"twingate_groups":         datasourceGroups(),
			"twingate_remote_network": datasourceRemoteNetwork(),
			"twingate_users":          datasourceUsers(),
			"twingate_user":           datasourceUser(),
			"twingate_connectors":     datasourceConnectors(),
			"twingate_resources":      datasourceResources(),
			"twingate_resource":       datasourceResource(),
			"twingate_connector":      datasourceConnector(),
		},
	}
	provider.ConfigureContextFunc = configure(version, provider)

	return provider
}

func providerOptions() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"api_token": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			DefaultFunc: schema.EnvDefaultFunc("TWINGATE_API_TOKEN", nil),
			Description: "The access key for API operations. You can retrieve this\n" +
				"from the Twingate Admin Console ([documentation](https://docs.twingate.com/docs/api-overview)).\n" +
				"Alternatively, this can be specified using the TWINGATE_API_TOKEN environment variable.",
		},
		"network": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   false,
			DefaultFunc: schema.EnvDefaultFunc("TWINGATE_NETWORK", nil),
			Description: "Your Twingate network ID for API operations.\n" +
				"You can find it in the Admin Console URL, for example:\n" +
				"`autoco.twingate.com`, where `autoco` is your network ID\n" +
				"Alternatively, this can be specified using the TWINGATE_NETWORK environment variable.",
		},
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   false,
			DefaultFunc: schema.EnvDefaultFunc("TWINGATE_URL", "twingate.com"),
			Description: "The default is 'twingate.com'\n" +
				"This is optional and shouldn't be changed under normal circumstances.",
		},
		"http_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("TWINGATE_HTTP_TIMEOUT", DefaultHTTPTimeout),
			Description: "Specifies a time limit in seconds for the http requests made. The default value is 10 seconds.\n" +
				"Alternatively, this can be specified using the TWINGATE_HTTP_TIMEOUT environment variable",
		},
		"http_max_retry": {
			Type:        schema.TypeInt,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("TWINGATE_HTTP_MAX_RETRY", DefaultHTTPMaxRetry),
			Description: "Specifies a retry limit for the http requests made. This default value is 10.\n" +
				"Alternatively, this can be specified using the TWINGATE_HTTP_MAX_RETRY environment variable",
		},
	}
}

func configure(version string, _ *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		apiToken := d.Get("api_token").(string)
		network := d.Get("network").(string)
		url := d.Get("url").(string)
		httpTimeout := d.Get("http_timeout").(int)
		httpMaxRetry := d.Get("http_max_retry").(int)

		var diags diag.Diagnostics

		if (apiToken != "") && (network != "") {
			client := NewClient(url, apiToken, network, time.Duration(httpTimeout)*time.Second, httpMaxRetry, version)

			return client, diags
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Twingate client",
			Detail:   "Unable to create anonymous Twingate client , token and network have to be provided ",
		})

		return nil, diags
	}
}
