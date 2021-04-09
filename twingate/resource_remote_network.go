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
			"is_active": {
				Type:     schema.TypeBool,
				Required: false,
			},
		},
	}
}

func resourceRemoteNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// // Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	remoteNetworkName := d.Get("name").(string)
	isActive := d.Get("is_active").(bool)

	log.Printf("[INFO] Creating remote network with name %s", remoteNetworkName)
	mutation := map[string]string{
		"query": fmt.Sprintf(`
		mutation{
		  remoteNetworkCreate(name: "%s", isActive: %t) {
			ok
			entity {
			  id
			}
		  }
		}
        `, remoteNetworkName, isActive),
	}
	responseBody, err := c.doGraphqlRequest(mutation)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create remote network",
			Detail:   fmt.Sprintf("Unable to create remote network with name %s", remoteNetworkName),
		})

		return diags
	}
	mutationRemoteNetwork, err := gabs.ParseJSON(responseBody)

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
	log.Printf("[INFO] Remote network %s created with id %s", remoteNetworkName, d.Id())
	resourceRemoteNetworkRead(ctx, d, m)

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

func resourceRemoteNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// // Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	remoteNetworkId := d.Id()

	mutation := map[string]string{
		"query": fmt.Sprintf(`
		{
		  remoteNetwork(id: "%s") {
			name
			isActive
		  }
		}

        `, remoteNetworkId),
	}
	responseBody, err := c.doGraphqlRequest(mutation)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read remote network",
			Detail:   fmt.Sprintf("Unable to read remote network with ID %s", remoteNetworkId),
		})

		return diags
	}
	queryRemoteNetwork, err := gabs.ParseJSON(responseBody)

	if err != nil {
		return diag.FromErr(err)
	}
	remoteNetwork := queryRemoteNetwork.Path("data.remoteNetwork")

	d.Set("is_active", remoteNetwork.Path("isActive").Data().(bool))

	return diags
}
