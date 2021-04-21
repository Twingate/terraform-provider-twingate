package twingate

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConnectorTokens() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorTokensCreate,
		ReadContext:   resourceConnectorTokensRead,
		DeleteContext: resourceConnectorTokensDelete,

		Schema: map[string]*schema.Schema{
			// required
			"connector_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the parent Connector",
			},
			// optional
			"keepers": {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of resource.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
			},
			// Computed
			"access_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The Access token of the parent Connector",
			},
			"refresh_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The Refresh Token of the parent Connector",
			},
		},
	}
}

func resourceConnectorTokensCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	connectorId := d.Get("connector_id").(string)

	d.SetId(connectorId)

	connector := Connector{Id: connectorId}
	err := client.generateConnectorTokens(&connector)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("access_token", connector.AccessToken); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting access_token: %w ", err))
	}
	if err := d.Set("refresh_token", connector.RefreshToken); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting refresh_token: %w ", err))
	}
	return resourceConnectorTokensRead(ctx, d, m)
}

func resourceConnectorTokensDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("[INFO] Destroyed ConnectorTokens id %s", d.Id())
	d.SetId("")

	return diags
}

func resourceConnectorTokensRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	accessToken := d.Get("access_token").(string)
	refreshToken := d.Get("refresh_token").(string)
	// Confirming the tokens are still valid
	err := client.verifyConnectorTokens(&refreshToken, &accessToken)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Unable to verify connector tokens",
			Detail:   fmt.Sprintf("Unable to verify connector %s tokens, assuming not valid and needs to be recreated", d.Id()),
		})
		d.SetId("")
		return diags
	}

	return diags
}
