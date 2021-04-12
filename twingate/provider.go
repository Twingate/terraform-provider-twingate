package twingate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("TWINGATE_TOKEN", nil),
			},
			"network": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc("TWINGATE_NETWORK", nil),
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   false,
				DefaultFunc: schema.EnvDefaultFunc("TWINGATE_URL", "twingate.com"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"twingate_remote_network": resourceRemoteNetwork(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"twingate_group": dataSourceGroup(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("token").(string)
	network := d.Get("network").(string)
	url := d.Get("url").(string)
	var diags diag.Diagnostics

	if (token != "") && (network != "") {
		c, err := NewClient(&network, &token, &url)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Twingate client",
				Detail:   "Unable to authenticate with provided token / network",
			})

			return nil, diags
		}

		return c, diags
	}

	c, err := NewClient(nil, nil, &url)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Twingate client",
			Detail:   "Unable to create anonymous Twingate client , token and network have to be provided",
		})

		return nil, diags
	}

	return c, diags
}
