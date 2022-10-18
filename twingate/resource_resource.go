package twingate

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/twingate/go-graphql-client"
)

const (
	policyRestricted = "RESTRICTED"
	policyAllowAll   = "ALLOW_ALL"
	policyDenyAll    = "DENY_ALL"
)

func castToStrings(a, b interface{}) (string, string) {
	return a.(string), b.(string)
}

func protocolDiff(k, oldValue, newValue string, d *schema.ResourceData) bool {
	keys := []string{"protocols.0.tcp.0.policy", "protocols.0.udp.0.policy"}
	for _, key := range keys {
		if strings.HasPrefix(k, key) {
			oldPolicy, newPolicy := castToStrings(d.GetChange(key))
			if oldPolicy == policyRestricted && newPolicy == policyDenyAll {
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
	oldPorts, newPorts := convertPortsToSlice(a.([]interface{})), convertPortsToSlice(b.([]interface{}))

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

func convertPortsRangeToMap(portsRange []*PortRangeInput) map[int32]struct{} {
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

func resourceResource() *schema.Resource { //nolint:funlen
	portsResource := schema.Resource{
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{policyRestricted, policyAllowAll, policyDenyAll}, false),
				Description:  fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", policyRestricted, policyAllowAll, policyDenyAll),
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
		CreateContext: resourceResourceCreate,
		UpdateContext: resourceResourceUpdate,
		ReadContext:   resourceResourceRead,
		DeleteContext: resourceResourceDelete,

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
				Description: "List of Group IDs that have permission to access the Resource. These IDs can be retrieved from the twingate_group and twingate_groups data sources, or directly from the Twingate API",
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

func convertGroupsGraphql(a []interface{}) []*graphql.ID {
	res := []*graphql.ID{}

	for _, elem := range a {
		id := graphql.ID(elem.(string))
		res = append(res, &id)
	}

	return res
}

func extractProtocolsFromContext(p interface{}) *StringProtocolsInput {
	protocolsMap := p.(map[string]interface{})
	protocols := &StringProtocolsInput{}
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
	case policyAllowAll:
		return policy, ports
	case policyDenyAll:
		return policyRestricted, nil
	}

	p := convertPortsToSlice(input["ports"].([]interface{}))
	if len(p) > 0 {
		ports = p
	}

	return policy, ports
}

func extractResource(resourceData *schema.ResourceData) (*Resource, error) {
	protocols, err := extractProtocols(resourceData)
	if err != nil {
		return nil, err
	}

	resource := &Resource{
		Name:            graphql.String(resourceData.Get("name").(string)),
		RemoteNetworkID: graphql.ID(resourceData.Get("remote_network_id").(string)),
		Address:         graphql.String(resourceData.Get("address").(string)),
		GroupsIds:       convertGroupsGraphql(resourceData.Get("group_ids").(*schema.Set).List()),
		Protocols:       protocols,
	}

	if resourceData.Id() != "" {
		resource.ID = resourceData.Id()
	}

	return resource, nil
}

func extractProtocols(resourceData *schema.ResourceData) (*ProtocolsInput, error) {
	p := resourceData.Get("protocols").([]interface{})
	if len(p) == 0 {
		return newEmptyProtocols(), nil
	}

	protocols, err := extractProtocolsFromContext(p[0]).convertToGraphql()
	if err != nil {
		return nil, err
	}

	return protocols, nil
}

func resourceResourceCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	req, err := extractResource(resourceData)
	if err != nil {
		return diag.FromErr(err)
	}

	resource, err := client.createResource(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Created resource %s", resource.Name)

	return resourceResourceReadHelper(ctx, client, resourceData, resource, nil)
}

func resourceResourceUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)
	resource, err := extractResource(resourceData)

	if err != nil {
		return diag.FromErr(err)
	}

	resource, err = client.updateResource(ctx, resource)

	return resourceResourceReadHelper(ctx, client, resourceData, resource, err)
}

func resourceResourceDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var diags diag.Diagnostics

	resourceID := resourceData.Id()

	err := client.deleteResource(ctx, resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] deleted resource id %s", resourceData.Id())

	return diags
}

func resourceResourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	resource, err := client.readResource(ctx, resourceData.Id())

	return resourceResourceReadHelper(ctx, client, resourceData, resource, err)
}

func resourceResourceReadHelper(ctx context.Context, client *Client, resourceData *schema.ResourceData, resource *Resource, err error) diag.Diagnostics {
	if err != nil {
		if errors.Is(err, ErrGraphqlResultIsEmpty) {
			// clear state
			resourceData.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	if resource.Protocols == nil {
		resource.Protocols = newEmptyProtocols()
	}

	if !resource.IsActive {
		// fix set active state for the resource on `terraform apply`
		err = client.updateResourceActiveState(ctx, &Resource{
			ID:       resource.ID,
			IsActive: true,
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceData.SetId(resource.ID.(string))

	return resourceResourceReadDiagnostics(resourceData, resource)
}

func resourceResourceReadDiagnostics(resourceData *schema.ResourceData, resource *Resource) diag.Diagnostics {
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

	if err := resourceData.Set("group_ids", resource.stringGroups()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting group_ids: %w ", err))
	}

	protocols := resource.Protocols.flattenProtocols()
	if err := resourceData.Set("protocols", protocols); err != nil {
		return diag.FromErr(fmt.Errorf("error setting protocols: %w ", err))
	}

	return diags
}
