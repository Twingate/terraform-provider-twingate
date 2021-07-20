package twingate

import (
	"context"

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
		client := NewClient(url, apiToken, network)
		return client, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create Twingate client",
		Detail:   "Unable to create anonymous Twingate client , token and network have to be provided ",
	})

	return nil, diags
}
