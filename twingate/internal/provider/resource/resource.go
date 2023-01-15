package resource

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Resource() *schema.Resource { //nolint:funlen
	portsResource := schema.Resource{
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(model.Policies, false),
				Description:  fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", model.PolicyRestricted, model.PolicyAllowAll, model.PolicyDenyAll),
			},
			"ports": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of port ranges between 1 and 65535 inclusive, in the format `100-200` for a range, or `8080` for a single port",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DiffSuppressOnRefresh: true,
				DiffSuppressFunc:      portsNotChanged,
			},
		},
	}

	return &schema.Resource{
		Description:   "Resources in Twingate represent servers on the private network that clients can connect to. Resources can be defined by IP, CIDR range, FQDN, or DNS zone. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
		CreateContext: resourceCreate,
		UpdateContext: resourceUpdate,
		ReadContext:   resourceRead,
		DeleteContext: resourceDelete,

		Schema: map[string]*schema.Schema{
			// required
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Resource",
			},
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Resource's IP/CIDR or FQDN/DNS zone",
			},
			"remote_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Remote Network ID where the Resource lives",
			},
			"group_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of Group IDs that have permission to access the Resource, cannot be generated by Terraform and must be retrieved from the Twingate Admin Console or API",
				Deprecated:  "The group_ids argument is now deprecated, and the new access block argument should be used instead. The group_ids argument will be removed in a future version of the provider.",
			},
			"protocols": {
				Type:                  schema.TypeList,
				Optional:              true,
				MaxItems:              1,
				Description:           "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction, and all protocols and ports are allowed.",
				DiffSuppressOnRefresh: true,
				DiffSuppressFunc:      protocolsDiff,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_icmp": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to allow ICMP (ping) traffic",
						},
						"tcp": {
							Type:                  schema.TypeList,
							Required:              true,
							MaxItems:              1,
							Elem:                  &portsResource,
							DiffSuppressOnRefresh: true,
							DiffSuppressFunc:      protocolDiff,
						},
						"udp": {
							Type:                  schema.TypeList,
							Required:              true,
							MaxItems:              1,
							Elem:                  &portsResource,
							DiffSuppressOnRefresh: true,
							DiffSuppressFunc:      protocolDiff,
						},
					},
				},
			},
			"is_visible": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if the Resource is active.",
			},
			"is_browser_shortcut_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether this Resource will display a browser shortcut in the client.",
			},
			// computed
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Autogenerated ID of the Resource, encoded in base64",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func castToStrings(a, b interface{}) (string, string) {
	return a.(string), b.(string)
}

func protocolDiff(k, oldValue, newValue string, d *schema.ResourceData) bool {
	keys := []string{"protocols.0.tcp.0.policy", "protocols.0.udp.0.policy"}
	for _, key := range keys {
		if strings.HasPrefix(k, key) {
			oldPolicy, newPolicy := castToStrings(d.GetChange(key))
			if oldPolicy == model.PolicyRestricted && newPolicy == model.PolicyDenyAll {
				return true
			}
		}
	}

	return false
}

func protocolsDiff(key, oldValue, newValue string, resourceData *schema.ResourceData) bool {
	switch key {
	case "protocols.#", "protocols.0.tcp.#", "protocols.0.udp.#":
		return oldValue == "1" && newValue == "0"

	case "protocols.0.tcp.0.policy", "protocols.0.udp.0.policy":
		oldPolicy, newPolicy := castToStrings(resourceData.GetChange(key))

		return oldPolicy == newPolicy

	default:
		return false
	}
}

func equalPorts(a, b interface{}) bool {
	oldPorts, newPorts := a.([]interface{}), b.([]interface{})

	oldPortsRange, err := convertPorts(oldPorts)
	if err != nil {
		return false
	}

	newPortsRange, err := convertPorts(newPorts)
	if err != nil {
		return false
	}

	oldPortsMap := convertPortsRangeToMap(oldPortsRange)
	newPortsMap := convertPortsRangeToMap(newPortsRange)

	return reflect.DeepEqual(oldPortsMap, newPortsMap)
}

func convertPortsRangeToMap(portsRange []*model.PortRange) map[int32]struct{} {
	out := make(map[int32]struct{})

	for _, port := range portsRange {
		if port.Start == port.End {
			out[port.Start] = struct{}{}

			continue
		}

		for i := port.Start; i <= port.End; i++ {
			out[i] = struct{}{}
		}
	}

	return out
}

func portsNotChanged(k, oldValue, newValue string, d *schema.ResourceData) bool {
	keys := []string{"protocols.0.tcp.0.ports", "protocols.0.udp.0.ports"}
	for _, key := range keys {
		if strings.HasPrefix(k, key) {
			return equalPorts(d.GetChange(key))
		}
	}

	return false
}

func resourceCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	resource, err := convertResource(resourceData)
	if err != nil {
		return diag.FromErr(err)
	}

	resource, err = client.CreateResource(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Created resource %s", resource.Name)

	return resourceResourceReadHelper(ctx, client, resourceData, resource, nil)
}

func resourceUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	resource, err := convertResource(resourceData)
	if err != nil {
		return diag.FromErr(err)
	}

	resource.ID = resourceData.Id()

	resource, err = client.UpdateResource(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Created resource %s", resource.Name)

	return resourceResourceReadHelper(ctx, client, resourceData, resource, nil)
}

func resourceDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	resourceID := resourceData.Id()

	err := c.DeleteResource(ctx, resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] deleted resource id %s", resourceData.Id())

	return nil
}

func resourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)
	resource, err := client.ReadResource(ctx, resourceData.Id())

	if resource != nil {
		_, exists := resourceData.GetOkExists("is_visible") //nolint
		if !exists {
			resource.IsVisible = nil
		}

		_, exists = resourceData.GetOkExists("is_browser_shortcut_enabled") //nolint
		if !exists {
			resource.IsBrowserShortcutEnabled = nil
		}
	}

	return resourceResourceReadHelper(ctx, client, resourceData, resource, err)
}

func resourceResourceReadHelper(ctx context.Context, resourceClient *client.Client, resourceData *schema.ResourceData, resource *model.Resource, err error) diag.Diagnostics {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if resource.Protocols == nil {
		resource.Protocols = model.DefaultProtocols()
	}

	if !resource.IsActive {
		// fix set active state for the resource on `terraform apply`
		err = resourceClient.UpdateResourceActiveState(ctx, &model.Resource{
			ID:       resource.ID,
			IsActive: true,
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceData.SetId(resource.ID)

	return readDiagnostics(resourceData, resource)
}

func readDiagnostics(resourceData *schema.ResourceData, resource *model.Resource) diag.Diagnostics {
	if err := resourceData.Set("name", resource.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %w ", err))
	}

	if err := resourceData.Set("remote_network_id", resource.RemoteNetworkID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting remote network: %w ", err))
	}

	if err := resourceData.Set("address", resource.Address); err != nil {
		return diag.FromErr(fmt.Errorf("error setting address: %w ", err))
	}

	if err := resourceData.Set("group_ids", resource.Groups); err != nil {
		return diag.FromErr(fmt.Errorf("error setting group_ids: %w ", err))
	}

	if err := resourceData.Set("protocols", resource.Protocols.ToTerraform()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting protocols: %w ", err))
	}

	if resource.IsVisible != nil {
		if err := resourceData.Set("is_visible", *resource.IsVisible); err != nil {
			return diag.FromErr(fmt.Errorf("error setting is_visible: %w ", err))
		}
	}

	if resource.IsBrowserShortcutEnabled != nil {
		if err := resourceData.Set("is_browser_shortcut_enabled", *resource.IsBrowserShortcutEnabled); err != nil {
			return diag.FromErr(fmt.Errorf("error setting is_browser_shortcut_enabled: %w ", err))
		}
	}

	return nil
}
