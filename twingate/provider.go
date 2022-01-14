package twingate

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	DefaultHTTPTimeout  = 10
	DefaultHTTPRetryMax = 10
)

func Provider(version string) *schema.Provider { //nolint:funlen
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
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
			"transport": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Transport settings",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_timeout": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     DefaultHTTPTimeout,
							Description: "Specifies a time limit in seconds for the http requests made",
						},
						"http_retry_max": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     DefaultHTTPRetryMax,
							Description: "Specifies a retry limit for the http requests made",
						},
					},
				},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"twingate_remote_network":   resourceRemoteNetwork(),
			"twingate_connector":        resourceConnector(),
			"twingate_connector_tokens": resourceConnectorTokens(),
			"twingate_resource":         resourceResource(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}
	provider.ConfigureContextFunc = configure(version, provider)

	return provider
}

func configure(version string, _ *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		apiToken := d.Get("api_token").(string)
		network := d.Get("network").(string)
		url := d.Get("url").(string)

		transportArray := d.Get("transport").([]interface{})

		var (
			httpTimeout  = DefaultHTTPTimeout
			httpRetryMax = DefaultHTTPRetryMax
		)

		if len(transportArray) > 0 {
			transportItem := transportArray[0].(map[string]interface{})
			httpTimeout = transportItem["http_timeout"].(int)
			httpRetryMax = transportItem["http_retry_max"].(int)
		}

		var diags diag.Diagnostics

		if (apiToken != "") && (network != "") {
			client := NewClient(url, apiToken, network, time.Duration(httpTimeout)*time.Second, httpRetryMax, version)

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
