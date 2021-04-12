package twingate

import (
	"context"
	"fmt"
	"log"

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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the remote network",
			},
			"is_active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the network is enabled or disabled",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRemoteNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkName := d.Get("name").(string)
	isActive := d.Get("is_active").(bool)

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
	mutationRemoteNetwork, err := c.doGraphqlRequest(mutation)

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
	d.SetId(mutationRemoteNetwork.Path("data.remoteNetworkCreate.entity.id").Data().(string))
	log.Printf("[INFO] Remote network %s created with id %s", remoteNetworkName, d.Id())
	resourceRemoteNetworkRead(ctx, d, m)

	return diags
}

func resourceRemoteNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkName := d.Get("name").(string)
	isActive := d.Get("is_active").(bool)

	if d.HasChanges("is_active", "name") {
		remoteNetworkId := d.Id()
		log.Printf("[INFO] Updating remote network id %s", remoteNetworkId)
		mutation := map[string]string{
			"query": fmt.Sprintf(`
				mutation {
					remoteNetworkUpdate(id: "%s", name: "%s", isActive: %t){
						ok
						error
					}
				}
        `, remoteNetworkId, remoteNetworkName, isActive),
		}
		mutationRemoteNetwork, err := c.doGraphqlRequest(mutation)
		if err != nil {
			return diag.FromErr(err)
		}
		status := mutationRemoteNetwork.Path("data.remoteNetworkUpdate.ok").Data().(bool)
		if !status {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update network",
				Detail:   fmt.Sprintf("Unable to update network %s", mutationRemoteNetwork.Path("data.remoteNetworkUpdate.error").String()),
			})

			return diags
		}
	}

	return resourceRemoteNetworkRead(ctx, d, m)
}

func resourceRemoteNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkId := d.Id()

	mutation := map[string]string{
		"query": fmt.Sprintf(`
		 mutation {
		  remoteNetworkDelete(id: "%s"){
			ok
			error
		  }
		}
		`, remoteNetworkId),
	}
	deleteRemoteNetwork, err := c.doGraphqlRequest(mutation)

	if err != nil {
		return diag.FromErr(err)
	}

	status := deleteRemoteNetwork.Path("data.remoteNetworkDelete.ok").Data().(bool)
	if !status {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete remote network",
			Detail:   fmt.Sprintf("Unable to delete remote network \"%s\"", deleteRemoteNetwork.Path("data.remoteNetworkDelete.error").String()),
		})

		return diags
	}
	log.Printf("[INFO] Deleted remote network id %s", d.Id())

	return diags
}

func resourceRemoteNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	var diags diag.Diagnostics

	remoteNetworkId := d.Id()

	log.Printf("[INFO] Reading remote network id %s", d.Id())

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
	queryRemoteNetwork, err := c.doGraphqlRequest(mutation)
	if err != nil {
		return diag.FromErr(err)
	}

	remoteNetwork := queryRemoteNetwork.Path("data.remoteNetwork")
	if remoteNetwork.Data() == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read remote network",
			Detail:   fmt.Sprintf("Unable to read remote network with id \"%s\"", remoteNetworkId),
		})

		return diags
	}

	err = d.Set("is_active", remoteNetwork.Path("isActive").Data().(bool))
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("name", remoteNetwork.Path("name").Data().(string))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
