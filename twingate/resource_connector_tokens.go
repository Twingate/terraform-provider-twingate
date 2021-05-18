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
		Description:   "This resource type will generate tokens for a Connector, which are needed to successfully provision one on your network. The Connector itself has its own resource type and must be created before you can provision tokens.",
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
				Description: "Arbitrary map of values that, when changed, will trigger recreation of resource. Use this to automatically rotate Connector tokens on a schedule.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
			},
			// Computed
			"access_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The Access Token of the parent Connector",
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

	connectorID := d.Get("connector_id").(string)

	d.SetId(connectorID)

	connector := Connector{ID: connectorID}
	err := client.generateConnectorTokens(&connector)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("access_token", connector.ConnectorTokens.AccessToken); err != nil {
		return diag.FromErr(fmt.Errorf("error setting access_token: %w ", err))
	}

	if err := d.Set("refresh_token", connector.ConnectorTokens.RefreshToken); err != nil {
		return diag.FromErr(fmt.Errorf("error setting refresh_token: %w ", err))
	}

	return resourceConnectorTokensRead(ctx, d, m)
}

func resourceConnectorTokensDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	connector := Connector{ID: d.Id()}
	// Just calling generate new tokens for the connector so the old ones are invalidated
	err := client.generateConnectorTokens(&connector)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Invalidating Connector Tokens id %s", d.Id())
	d.SetId("")

	return diags
}

func resourceConnectorTokensRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	accessToken := d.Get("access_token").(string)
	refreshToken := d.Get("refresh_token").(string)

	err := client.verifyConnectorTokens(refreshToken, accessToken)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "can't to verify connector tokens",
			Detail:   fmt.Sprintf("can't verify connector %s tokens, assuming not valid and needs to be recreated", d.Id()),
		})

		d.SetId("")

		return diags
	}

	return diags
}
