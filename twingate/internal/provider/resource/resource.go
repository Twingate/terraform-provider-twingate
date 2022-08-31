package resource

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/twingate/go-graphql-client"
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
	oldPorts, newPorts := transport.ConvertPortsToSlice(a.([]interface{})), transport.ConvertPortsToSlice(b.([]interface{}))

	oldPortsRange, err := transport.ConvertPorts(oldPorts)
	if err != nil {
		return false
	}

	newPortsRange, err := transport.ConvertPorts(newPorts)
	if err != nil {
		return false
	}

	oldPortsMap := convertPortsRangeToMap(oldPortsRange)
	newPortsMap := convertPortsRangeToMap(newPortsRange)

	return reflect.DeepEqual(oldPortsMap, newPortsMap)
}

func convertPortsRangeToMap(portsRange []*transport.PortRangeInput) map[int32]struct{} {
	out := make(map[int32]struct{})

	for _, port := range portsRange {
		if port.Start == port.End {
			out[int32(port.Start)] = struct{}{}

			continue
		}

		for i := int32(port.Start); i <= int32(port.End); i++ {
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

func convertGroupsGraphql(a []interface{}) []*graphql.ID {
	res := []*graphql.ID{}

	for _, elem := range a {
		id := graphql.ID(elem.(string))
		res = append(res, &id)
	}

	return res
}

func extractProtocolsFromContext(p interface{}) *transport.StringProtocolsInput {
	protocolsMap := p.(map[string]interface{})
	protocols := &transport.StringProtocolsInput{}
	protocols.AllowIcmp = protocolsMap["allow_icmp"].(bool)

	u := protocolsMap["udp"].([]interface{})
	if len(u) > 0 {
		udp := u[0].(map[string]interface{})
		protocols.UDPPolicy, protocols.UDPPorts = parseProtocol(udp)
	}

	t := protocolsMap["tcp"].([]interface{})
	if len(t) > 0 {
		tcp := t[0].(map[string]interface{})
		protocols.TCPPolicy, protocols.TCPPorts = parseProtocol(tcp)
	}

	return protocols
}

func parseProtocol(input map[string]interface{}) (string, []string) {
	var ports []string

	policy := input["policy"].(string)

	switch policy {
	case model.PolicyAllowAll:
		return policy, ports
	case model.PolicyDenyAll:
		return model.PolicyRestricted, nil
	}

	p := transport.ConvertPortsToSlice(input["ports"].([]interface{}))
	if len(p) > 0 {
		ports = p
	}

	return policy, ports
}

func extractResource(resourceData *schema.ResourceData) (*transport.Resource, error) {
	resource := &transport.Resource{
		Name:            graphql.String(resourceData.Get("name").(string)),
		RemoteNetworkID: graphql.ID(resourceData.Get("remote_network_id").(string)),
		Address:         graphql.String(resourceData.Get("address").(string)),
		GroupsIds:       convertGroupsGraphql(resourceData.Get("group_ids").(*schema.Set).List()),
	}

	p := resourceData.Get("protocols").([]interface{})

	if len(p) > 0 {
		p, err := extractProtocolsFromContext(p[0]).ConvertToGraphql()
		if err != nil {
			return nil, err
		}

		resource.Protocols = p
	} else {
		resource.Protocols = transport.NewEmptyProtocols()
	}

	return resource, nil
}

func resourceCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	resource, err := extractResource(resourceData)
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.CreateResource(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	resourceData.SetId(resource.ID.(string))
	log.Printf("[INFO] Created resource %s", resource.Name)

	waitForResourceAvailability()

	return resourceRead(ctx, resourceData, meta)
}

func resourceUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	if resourceData.HasChanges("protocols", "remote_network_id", "name", "address", "group_ids") {
		resource, err := extractResource(resourceData)
		if err != nil {
			return diag.FromErr(err)
		}

		resource.ID = resourceData.Id()

		err = client.UpdateResource(ctx, resource)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRead(ctx, resourceData, meta)
}

func resourceDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)

	var diags diag.Diagnostics

	resourceID := resourceData.Id()

	err := client.DeleteResource(ctx, resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] deleted resource id %s", resourceData.Id())

	return diags
}

func resourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*transport.Client)
	resourceID := resourceData.Id()

	resource, err := client.ReadResource(ctx, resourceID)
	if err != nil {
		if errors.Is(err, transport.ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if resource.Protocols == nil {
		resource.Protocols = transport.NewEmptyProtocols()
	}

	if !resource.IsActive {
		// fix set active state for the resource on `terraform apply`
		err = client.UpdateResourceActiveState(ctx, &transport.Resource{
			ID:       resourceID,
			IsActive: true,
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceReadDiagnostics(resourceData, resource)
}

func ResourceReadDiagnostics(resourceData *schema.ResourceData, resource *transport.Resource) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := resourceData.Set("name", resource.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %w ", err))
	}

	if err := resourceData.Set("remote_network_id", resource.RemoteNetworkID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting remote network: %w ", err))
	}

	if err := resourceData.Set("address", resource.Address); err != nil {
		return diag.FromErr(fmt.Errorf("error setting address: %w ", err))
	}

	if err := resourceData.Set("group_ids", resource.StringGroups()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting group_ids: %w ", err))
	}

	protocols := resource.Protocols.FlattenProtocols()
	if err := resourceData.Set("protocols", protocols); err != nil {
		return diag.FromErr(fmt.Errorf("error setting protocols: %w ", err))
	}

	return diags
}

func waitForResourceAvailability() {
	time.Sleep(time.Second)
}
