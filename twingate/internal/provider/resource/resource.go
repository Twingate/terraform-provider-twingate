package resource

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Resource() *schema.Resource { //nolint:funlen
	portsSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			attr.Policy: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(model.Policies, false),
				Description:  fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", model.PolicyRestricted, model.PolicyAllowAll, model.PolicyDenyAll),
			},
			attr.Ports: {
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

	protocolsSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			attr.AllowIcmp: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to allow ICMP (ping) traffic",
			},
			attr.TCP: {
				Type:                  schema.TypeList,
				Required:              true,
				MaxItems:              1,
				Elem:                  portsSchema,
				DiffSuppressOnRefresh: true,
				DiffSuppressFunc:      protocolDiff,
			},
			attr.UDP: {
				Type:                  schema.TypeList,
				Required:              true,
				MaxItems:              1,
				Elem:                  portsSchema,
				DiffSuppressOnRefresh: true,
				DiffSuppressFunc:      protocolDiff,
			},
		},
	}

	accessSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			attr.GroupIDs: {
				Type:         schema.TypeSet,
				Elem:         &schema.Schema{Type: schema.TypeString},
				MinItems:     1,
				Optional:     true,
				AtLeastOneOf: []string{attr.Path(attr.Access, attr.ServiceAccountIDs)},
				Description:  "List of Group IDs that will have permission to access the Resource.",
			},
			attr.ServiceAccountIDs: {
				Type:         schema.TypeSet,
				Elem:         &schema.Schema{Type: schema.TypeString},
				MinItems:     1,
				Optional:     true,
				AtLeastOneOf: []string{attr.Path(attr.Access, attr.GroupIDs)},
				Description:  "List of Service Account IDs that will have permission to access the Resource.",
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
			attr.Name: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Resource",
			},
			attr.Address: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Resource's IP/CIDR or FQDN/DNS zone",
			},
			attr.RemoteNetworkID: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Remote Network ID where the Resource lives",
			},
			// optional
			attr.IsAuthoritative: {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to `false`, assignments made outside of Terraform will be ignored.",
			},
			attr.Protocols: {
				Type:                  schema.TypeList,
				Optional:              true,
				MaxItems:              1,
				Description:           "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction, and all protocols and ports are allowed.",
				DiffSuppressOnRefresh: true,
				DiffSuppressFunc:      protocolsDiff,
				Elem:                  protocolsSchema,
			},
			attr.Access: {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Restrict access to certain groups or service accounts",
				Elem:        accessSchema,
			},
			// computed
			attr.IsVisible: {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Controls whether this Resource will be visible in the main Resource list in the Twingate Client.",
			},
			attr.IsBrowserShortcutEnabled: {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: `Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.`,
			},
			attr.Alias: {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Set a DNS alias address for the Resource. Must be a DNS-valid name string.",
				DiffSuppressFunc: aliasDiff,
			},
			attr.ID: {
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

	if err = client.AddResourceServiceAccountIDs(ctx, resource); err != nil {
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

	if err = deleteResourceGroupIDs(ctx, resourceData, resource, client); err != nil {
		return diag.FromErr(err)
	}

	if err = deleteResourceServiceAccountIDs(ctx, resourceData, resource, client); err != nil {
		return diag.FromErr(err)
	}

	resource, err = client.UpdateResource(ctx, resource)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = client.AddResourceServiceAccountIDs(ctx, resource); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Updated resource %s", resource.Name)

	return resourceResourceReadHelper(ctx, client, resourceData, resource, nil)
}

func resourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	resource, err := client.ReadResource(ctx, resourceData.Id())
	if resource != nil {
		resource.IsAuthoritative = convertAuthoritativeFlag(resourceData)
	}

	return resourceResourceReadHelper(ctx, client, resourceData, resource, err)
}

func resourceDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*client.Client)
	resourceID := resourceData.Id()

	err := c.DeleteResource(ctx, resourceID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleted resource id %s", resourceData.Id())

	return nil
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

	remoteServiceAccounts, err := resourceClient.ReadResourceServiceAccounts(ctx, resource.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	resource.ServiceAccounts = remoteServiceAccounts

	if !resource.IsAuthoritative {
		groups, serviceAccounts := convertAccess(resourceData)
		resource.ServiceAccounts = setIntersection(serviceAccounts, resource.ServiceAccounts)
		resource.Groups = setIntersection(groups, resource.Groups)
	}

	resourceData.SetId(resource.ID)

	return readDiagnostics(resourceData, resource)
}

func readDiagnostics(resourceData *schema.ResourceData, resource *model.Resource) diag.Diagnostics { //nolint:cyclop
	if err := resourceData.Set(attr.Name, resource.Name); err != nil {
		return ErrAttributeSet(err, attr.Name)
	}

	if err := resourceData.Set(attr.RemoteNetworkID, resource.RemoteNetworkID); err != nil {
		return ErrAttributeSet(err, attr.RemoteNetworkID)
	}

	if err := resourceData.Set(attr.Address, resource.Address); err != nil {
		return ErrAttributeSet(err, attr.Address)
	}

	if err := resourceData.Set(attr.IsAuthoritative, resource.IsAuthoritative); err != nil {
		return ErrAttributeSet(err, attr.IsAuthoritative)
	}

	if err := resourceData.Set(attr.Access, resource.AccessToTerraform()); err != nil {
		return ErrAttributeSet(err, attr.Access)
	}

	if err := resourceData.Set(attr.Protocols, resource.Protocols.ToTerraform()); err != nil {
		return ErrAttributeSet(err, attr.Protocols)
	}

	if resource.IsVisible != nil {
		if err := resourceData.Set(attr.IsVisible, *resource.IsVisible); err != nil {
			return ErrAttributeSet(err, attr.IsVisible)
		}
	}

	if resource.IsBrowserShortcutEnabled != nil {
		if err := resourceData.Set(attr.IsBrowserShortcutEnabled, *resource.IsBrowserShortcutEnabled); err != nil {
			return ErrAttributeSet(err, attr.IsBrowserShortcutEnabled)
		}
	}

	var alias interface{}
	if resource.Alias != nil {
		alias = *resource.Alias
	}

	if err := resourceData.Set(attr.Alias, alias); err != nil {
		return ErrAttributeSet(err, attr.Alias)
	}

	return nil
}

func protocolDiff(attribute, oldValue, newValue string, data *schema.ResourceData) bool {
	keys := []string{
		attr.Path(attr.Protocols, attr.TCP, attr.Policy),
		attr.Path(attr.Protocols, attr.UDP, attr.Policy),
	}

	for _, key := range keys {
		if strings.HasPrefix(attribute, key) {
			oldPolicy, newPolicy := castToStrings(data.GetChange(key))
			if oldPolicy == model.PolicyRestricted && newPolicy == model.PolicyDenyAll {
				return true
			}
		}
	}

	return false
}

func protocolsDiff(key, oldValue, newValue string, resourceData *schema.ResourceData) bool {
	switch key {
	case attr.Len(attr.Protocols),
		attr.Len(attr.Protocols, attr.TCP),
		attr.Len(attr.Protocols, attr.UDP):
		return oldValue == "1" && newValue == "0"

	case attr.Path(attr.Protocols, attr.TCP, attr.Policy),
		attr.Path(attr.Protocols, attr.UDP, attr.Policy):
		oldPolicy, newPolicy := castToStrings(resourceData.GetChange(key))

		return oldPolicy == newPolicy

	default:
		return false
	}
}

func aliasDiff(key, _, _ string, resourceData *schema.ResourceData) bool {
	oldVal, newVal := castToStrings(resourceData.GetChange(key))

	return oldVal == newVal
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

func convertPortsRangeToMap(portsRange []*model.PortRange) map[int]struct{} {
	out := make(map[int]struct{})

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

func portsNotChanged(attribute, oldValue, newValue string, data *schema.ResourceData) bool {
	keys := []string{
		attr.Path(attr.Protocols, attr.TCP, attr.Ports),
		attr.Path(attr.Protocols, attr.UDP, attr.Ports),
	}

	for _, key := range keys {
		if strings.HasPrefix(attribute, key) {
			return equalPorts(data.GetChange(key))
		}
	}

	return false
}

func deleteResourceGroupIDs(ctx context.Context, resourceData *schema.ResourceData, resource *model.Resource, client *client.Client) error {
	groupIDs := getIDsToDelete(ctx, resourceData, resource.Groups, attr.GroupIDs, resource, client)

	return client.DeleteResourceGroups(ctx, resource.ID, groupIDs) //nolint
}

func deleteResourceServiceAccountIDs(ctx context.Context, resourceData *schema.ResourceData, resource *model.Resource, client *client.Client) error {
	idsToDelete := getIDsToDelete(ctx, resourceData, resource.ServiceAccounts, attr.ServiceAccountIDs, resource, client)

	return client.DeleteResourceServiceAccounts(ctx, resource.ID, idsToDelete) //nolint
}

func getIDsToDelete(ctx context.Context, resourceData *schema.ResourceData, currentIDs []string, attribute string, resource *model.Resource, client *client.Client) []string {
	oldIDs := getOldIDs(ctx, resourceData, attribute, resource, client)
	if len(oldIDs) == 0 {
		return nil
	}

	return setDifference(oldIDs, currentIDs)
}

func getOldIDs(ctx context.Context, resourceData *schema.ResourceData, attribute string, resource *model.Resource, client *client.Client) []string {
	if resource.IsAuthoritative {
		return getOldIDsAuthoritative(ctx, resource, client, attribute)
	}

	return getOldIDsNonAuthoritative(resourceData, attribute)
}

func getOldIDsNonAuthoritative(resourceData *schema.ResourceData, attribute string) []string {
	if resourceData.HasChange(attribute) {
		old, _ := resourceData.GetChange(attribute)

		return convertIDs(old)
	}

	if resourceData.HasChange(attr.Path(attr.Access, attribute)) {
		old, _ := resourceData.GetChange(attr.Path(attr.Access, attribute))

		return convertIDs(old)
	}

	return nil
}

func getOldIDsAuthoritative(ctx context.Context, resource *model.Resource, client *client.Client, attribute string) []string {
	switch attribute {
	case attr.ServiceAccountIDs:
		serviceAccounts, err := client.ReadResourceServiceAccounts(ctx, resource.ID)
		if err != nil {
			return nil
		}

		return serviceAccounts

	case attr.GroupIDs:
		res, err := client.ReadResource(ctx, resource.ID)
		if err != nil {
			return nil
		}

		return res.Groups
	}

	return nil
}

func convertResource(data *schema.ResourceData) (*model.Resource, error) {
	protocols, err := convertProtocols(data)
	if err != nil {
		return nil, err
	}

	groups, serviceAccounts := convertAccess(data)
	res := &model.Resource{
		Name:            data.Get(attr.Name).(string),
		RemoteNetworkID: data.Get(attr.RemoteNetworkID).(string),
		Address:         data.Get(attr.Address).(string),
		Protocols:       protocols,
		Groups:          groups,
		ServiceAccounts: serviceAccounts,
		IsAuthoritative: convertAuthoritativeFlag(data),
		Alias:           getOptionalString(data, attr.Alias),
	}

	isVisible, ok := data.GetOkExists(attr.IsVisible) //nolint
	if val := isVisible.(bool); ok {
		res.IsVisible = &val
	}

	isBrowserShortcutEnabled, ok := data.GetOkExists(attr.IsBrowserShortcutEnabled) //nolint:staticcheck
	if val := isBrowserShortcutEnabled.(bool); ok {
		res.IsBrowserShortcutEnabled = &val
	}

	return res, nil
}

func getOptionalString(data *schema.ResourceData, attr string) *string {
	var result *string

	cfg := data.GetRawConfig()
	val := cfg.GetAttr(attr)

	if !val.IsNull() {
		str := val.AsString()
		result = &str
	}

	return result
}

func convertAccess(data *schema.ResourceData) ([]string, []string) {
	rawList := data.Get(attr.Access).([]interface{})
	if len(rawList) == 0 || rawList[0] == nil {
		return nil, nil
	}

	rawMap := rawList[0].(map[string]interface{})

	return convertIDs(rawMap[attr.GroupIDs]), convertIDs(rawMap[attr.ServiceAccountIDs])
}

func convertAuthoritativeFlag(data *schema.ResourceData) bool {
	flag, hasFlag := data.GetOkExists(attr.IsAuthoritative) //nolint:staticcheck

	if hasFlag {
		return flag.(bool)
	}

	// default value
	return true
}

func convertProtocols(data *schema.ResourceData) (*model.Protocols, error) {
	rawList := data.Get(attr.Protocols).([]interface{})
	if len(rawList) == 0 {
		return model.DefaultProtocols(), nil
	}

	rawMap := rawList[0].(map[string]interface{})

	udp, err := convertProtocol(rawMap[attr.UDP].([]interface{}))
	if err != nil {
		return nil, err
	}

	tcp, err := convertProtocol(rawMap[attr.TCP].([]interface{}))
	if err != nil {
		return nil, err
	}

	return &model.Protocols{
		UDP:       udp,
		TCP:       tcp,
		AllowIcmp: rawMap[attr.AllowIcmp].(bool),
	}, nil
}

func convertProtocol(rawList []interface{}) (*model.Protocol, error) {
	if len(rawList) == 0 {
		return nil, nil //nolint:nilnil
	}

	rawMap := rawList[0].(map[string]interface{})
	policy := rawMap[attr.Policy].(string)

	ports, err := convertPorts(rawMap[attr.Ports].([]interface{}))
	if err != nil {
		return nil, err
	}

	return model.NewProtocol(policy, ports), nil
}

func convertPorts(rawList []interface{}) ([]*model.PortRange, error) {
	var ports = make([]*model.PortRange, 0, len(rawList))

	for _, port := range rawList {
		var str string
		if port != nil {
			str = port.(string)
		}

		portRange, err := model.NewPortRange(str)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		ports = append(ports, portRange)
	}

	return ports, nil
}

func convertUsers(data *schema.ResourceData) []string {
	if ids, ok := data.GetOk(attr.UserIDs); ok {
		return convertIDs(ids)
	}

	return nil
}
