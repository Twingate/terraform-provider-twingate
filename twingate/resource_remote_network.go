package twingate

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRemoteNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRemoteNetworkCreate,
		ReadContext:   resourceRemoteNetworkRead,
		UpdateContext: resourceRemoteNetworkUpdate,
		DeleteContext: resourceRemoteNetworkDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceRemoteNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// // Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	remoteNetworkName := d.Get("name").(string)
	log.Printf("[INFO] Creating remote network with name %s", remoteNetworkName)
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		mutation{
		  remoteNetworkCreate(name: "%s") {
			ok
			entity {
			  id
			}
		  }
		}
        `, remoteNetworkName),
	}
	body, err := c.doGraphqlRequest(mutation)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create remote network",
			Detail:   "Unable to create remote network",
		})

		return diags
	}
	mutationRemoteNetwork, err := gabs.ParseJSON(body)

	if err != nil {
		return diag.FromErr(err)
	}
	status := mutationRemoteNetwork.Path("data.remoteNetworkCreate.ok").Data().(bool)
	if !status {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create remote network",
			Detail:   fmt.Sprintf("Unable to create remote network %s", mutationRemoteNetwork.Path("data.remoteNetworkCreate.error").String()),
		})

		return diags
	}
	d.SetId(strings.Trim(mutationRemoteNetwork.Path("data.remoteNetworkCreate.entity.id").String(), "\""))

	resourceRemoteNetworkRead(ctx, d, m)

	return diags
}

func resourceRemoteNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceRemoteNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceRemoteNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}
