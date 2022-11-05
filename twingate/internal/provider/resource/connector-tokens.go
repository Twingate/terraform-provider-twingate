package resource

import (
	"context"
	"fmt"
	"log"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ConnectorTokens() *schema.Resource {
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

func resourceConnectorTokensCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	connectorID := resourceData.Get("connector_id").(string)
	resourceData.SetId(connectorID)

	tokens, err := client.GenerateConnectorTokens(ctx, connectorID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set("access_token", tokens.AccessToken); err != nil {
		return diag.FromErr(fmt.Errorf("error setting access_token: %w ", err))
	}

	if err := resourceData.Set("refresh_token", tokens.RefreshToken); err != nil {
		return diag.FromErr(fmt.Errorf("error setting refresh_token: %w ", err))
	}

	return resourceConnectorTokensRead(ctx, resourceData, meta)
}

func resourceConnectorTokensDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	var diags diag.Diagnostics

	// Just calling generate new tokens for the connector so the old ones are invalidated
	_, err := client.GenerateConnectorTokens(ctx, resourceData.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Invalidating Connector Tokens id %s", resourceData.Id())
	resourceData.SetId("")

	return diags
}

func resourceConnectorTokensRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	var diags diag.Diagnostics

	accessToken := resourceData.Get("access_token").(string)
	refreshToken := resourceData.Get("refresh_token").(string)

	err := client.VerifyConnectorTokens(ctx, refreshToken, accessToken)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "can't to verify connector tokens",
			Detail:   fmt.Sprintf("can't verify connector %s tokens, assuming not valid and needs to be recreated", resourceData.Id()),
		})

		resourceData.SetId("")

		return diags
	}

	return diags
}
