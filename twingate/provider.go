package twingate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TWINGATE_API_TOKEN", nil),
				Description: "The access key for API operations. You can retrieve this\n" +
					"from the 'Settings -> API' section of the Twingate console.",
			},
			"network": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc("TWINGATE_NETWORK", nil),
				Description: "The name of twingate network for API operations\n" +
					"You can find it in URL, for example:\n" +
					"mycompany.twingate.com where 'mycompany' is network",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc("TWINGATE_URL", "twingate.com"),
				Description: "Default is 'twingate.com'\n" +
					"Optional and shouldn't be changed",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"twingate_remote_network":   resourceRemoteNetwork(),
			"twingate_connector":        resourceConnector(),
			"twingate_connector_tokens": resourceConnectorTokens(),
			"twingate_resource":         resourceResource(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiToken := d.Get("api_token").(string)
	network := d.Get("network").(string)
	url := d.Get("url").(string)

	var diags diag.Diagnostics

	if (apiToken != "") && (network != "") {
		client := NewClient(network, apiToken, url)

		err := client.ping()

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Twingate client",
				Detail:   fmt.Sprintf("Unable to authenticate with provided api token and network %s : %s", network, err),
			})

			return nil, diags
		}

		return client, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create Twingate client",
		Detail:   "Unable to create anonymous Twingate client , token and network have to be provided ",
	})

	return nil, diags
}
