// Package twingate is a Terraform provider for Twingate platform
package twingate

import (
	"context"
	"fmt"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/datasource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func Provider(version string) *schema.Provider {
	provider := &schema.Provider{
		Schema: providerOptions(),
		ResourcesMap: map[string]*schema.Resource{
			"twingate_remote_network":   resource.RemoteNetwork(),
			"twingate_connector":        resource.Connector(),
			"twingate_connector_tokens": resource.ConnectorTokens(),
			"twingate_group":            resource.Group(),
			"twingate_resource":         resource.Resource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"twingate_group":          datasource.Group(),
			"twingate_groups":         datasource.Groups(),
			"twingate_remote_network": datasource.RemoteNetwork(),
			"twingate_users":          datasource.Users(),
			"twingate_user":           datasource.User(),
			"twingate_connectors":     datasource.Connectors(),
			"twingate_resources":      datasource.Resources(),
			"twingate_resource":       datasource.Resource(),
			"twingate_connector":      datasource.Connector(),
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
			DefaultFunc: schema.EnvDefaultFunc(EnvAPIToken, nil),
			Description: fmt.Sprintf("The access key for API operations. You can retrieve this\n"+
				"from the Twingate Admin Console ([documentation](https://docs.twingate.com/docs/api-overview)).\n"+
				"Alternatively, this can be specified using the %s environment variable.", EnvAPIToken),
		},
		"network": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   false,
			DefaultFunc: schema.EnvDefaultFunc(EnvNetwork, nil),
			Description: fmt.Sprintf("Your Twingate network ID for API operations.\n"+
				"You can find it in the Admin Console URL, for example:\n"+
				"`autoco.twingate.com`, where `autoco` is your network ID\n"+
				"Alternatively, this can be specified using the %s environment variable.", EnvNetwork),
		},
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   false,
			DefaultFunc: schema.EnvDefaultFunc(EnvURL, DefaultURL),
			Description: fmt.Sprintf("The default is '%s'\n"+
				"This is optional and shouldn't be changed under normal circumstances.", DefaultURL),
		},
		"http_timeout": {
			Type:        schema.TypeInt,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc(EnvHTTPTimeout, DefaultHTTPTimeout),
			Description: fmt.Sprintf("Specifies a time limit in seconds for the http requests made. The default value is %s seconds.\n"+
				"Alternatively, this can be specified using the %s environment variable", DefaultHTTPTimeout, EnvHTTPTimeout),
		},
		"http_max_retry": {
			Type:        schema.TypeInt,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc(EnvHTTPMaxRetry, DefaultHTTPMaxRetry),
			Description: fmt.Sprintf("Specifies a retry limit for the http requests made. This setting is %s.\n"+
				"Alternatively, this can be specified using the %s environment variable", DefaultHTTPMaxRetry, EnvHTTPMaxRetry),
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
			client := transport.NewClient(url, apiToken, network, time.Duration(httpTimeout)*time.Second, httpMaxRetry, version)

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
