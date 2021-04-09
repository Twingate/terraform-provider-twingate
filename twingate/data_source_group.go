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

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIngredientsRead,
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

func dataSourceIngredientsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// // Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	groupName := d.Get("name").(string)

	query := map[string]string{
		"query": `
		{
		  groups {
			edges {
			  node {
				name
				id
			  }
			}
		  }
		}
        `,
	}

	body, err := c.doGraphqlRequest(query)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to fetch groups",
			Detail:   "Unable to fetch groups",
		})

		return diags
	}
	queryGroups, err := gabs.ParseJSON(body)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, group := range queryGroups.Path("data.groups.edges").Children() {
		name := strings.Trim(group.Path("node.name").String(), "\"")
		id := strings.Trim(group.Path("node.id").String(), "\"")
		log.Printf("[INFO] iterating over %s with ID %s", name, id)
		if groupName == name {
			_ = d.Set("name", name)
			d.SetId(id)
			log.Printf("[INFO] Found group named %s with ID %s", name, id)

			break
		}
	}
	if d.Id() == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find group",
			Detail:   fmt.Sprintf("Unable to find group with Name %s", groupName),
		})
	}

	return diags
}
