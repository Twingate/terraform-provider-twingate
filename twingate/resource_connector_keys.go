package twingate

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConnectorKeys() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorKeysCreate,
		ReadContext:   resourceConnectorKeysRead,
		DeleteContext: resourceConnectorKeysDelete,

		Schema: map[string]*schema.Schema{
			//required
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
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Description: `The name used for this key pair`,
			},
			"created_at": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceConnectorKeysCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	connectorId := d.Get("connector_id").(string)

	decodedId, err := base64.StdEncoding.DecodeString(connectorId)
	if err != nil {
		return diag.FromErr(err)
	}

	encodedId := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s-key", string(decodedId))))
	d.SetId(encodedId)

	connector := Connector{Id: connectorId}
	err = client.generateConnectorTokens(&connector)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("access_token", connector.AccessToken); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting access_token: %s ", err))
	}
	if err := d.Set("refresh_token", connector.RefreshToken); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting refresh_token: %s ", err))
	}
	if err := d.Set("name", fmt.Sprintf("%s-key", string(decodedId))); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting name: %s ", err))
	}
	if err := d.Set("created_at", time.Now().Format(time.RFC850)); err != nil {
		return diag.FromErr(fmt.Errorf("Error setting created_at: %s ", err))
	}
	return resourceConnectorKeysRead(ctx, d, m)

}

func resourceConnectorKeysDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("[INFO] Destroyed ConnectorKeys id %s", d.Id())
	d.SetId("")

	return diags
}

func resourceConnectorKeysRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	accessToken := d.Get("access_token").(string)
	refreshToken := d.Get("refresh_token").(string)
	// Confirming the tokens are still valid
	err := client.verifyConnectorTokens(&refreshToken, &accessToken)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Unable to verify connector  tokens",
			Detail:   fmt.Sprintf("Unable to verify connector %s tokens , assuming not valid and needs to be recreated", d.Id()),
		})
		d.SetId("")
		return diags
	}

	return diags

}
