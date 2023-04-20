package resource

import (
	"context"
	"fmt"
	"log"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
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
			attr.ConnectorID: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the parent Connector",
			},
			// optional
			attr.Keepers: {
				Description: "Arbitrary map of values that, when changed, will trigger recreation of resource. Use this to automatically rotate Connector tokens on a schedule.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
			},
			// Computed
			attr.AccessToken: {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The Access Token of the parent Connector",
			},
			attr.RefreshToken: {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The Refresh Token of the parent Connector",
			},
		},
	}
}

func resourceConnectorTokensCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)

	connectorID := resourceData.Get(attr.ConnectorID).(string)
	resourceData.SetId(connectorID)

	tokens, err := c.GenerateConnectorTokens(ctx, connectorID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := resourceData.Set(attr.AccessToken, tokens.AccessToken); err != nil {
		return ErrAttributeSet(err, attr.AccessToken)
	}

	if err := resourceData.Set(attr.RefreshToken, tokens.RefreshToken); err != nil {
		return ErrAttributeSet(err, attr.RefreshToken)
	}

	return resourceConnectorTokensRead(ctx, resourceData, meta)
}

func resourceConnectorTokensDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)

	// Just calling generate new tokens for the connector so the old ones are invalidated
	_, err := c.GenerateConnectorTokens(ctx, resourceData.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Invalidating Connector Tokens id %s", resourceData.Id())
	resourceData.SetId("")

	return nil
}

func resourceConnectorTokensRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	accessToken := resourceData.Get(attr.AccessToken).(string)
	refreshToken := resourceData.Get(attr.RefreshToken).(string)

	err := c.VerifyConnectorTokens(ctx, refreshToken, accessToken)
	if err != nil {
		resourceData.SetId("")

		return diag.Diagnostics{
			{
				Severity: diag.Warning,
				Summary:  "can't to verify connector tokens",
				Detail:   fmt.Sprintf("can't verify connector %s tokens, assuming not valid and needs to be recreated", resourceData.Id()),
			},
		}
	}

	return nil
}
