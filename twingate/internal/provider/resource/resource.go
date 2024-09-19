package resource

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	DefaultSecurityPolicyName       = "Default Policy"
	schemaVersion             int64 = 2
)

var (
	DefaultSecurityPolicyID               string //nolint:gochecknoglobals
	ErrPortsWithPolicyAllowAll            = errors.New(model.PolicyAllowAll + " policy does not allow specifying ports.")
	ErrPortsWithPolicyDenyAll             = errors.New(model.PolicyDenyAll + " policy does not allow specifying ports.")
	ErrPolicyRestrictedWithoutPorts       = errors.New(model.PolicyRestricted + " policy requires specifying ports.")
	ErrInvalidAttributeCombination        = errors.New("invalid attribute combination")
	ErrWildcardAddressWithEnabledShortcut = errors.New("Resources with a CIDR range or wildcard can't have the browser shortcut enabled.")
	ErrDefaultPolicyNotSet                = errors.New("default policy not set")
	ErrWrongGlobalID                      = errors.New("Unable to parse global ID")
)

// Ensure the implementation satisfies the desired interfaces.
var _ resource.Resource = &twingateResource{}

func NewResourceResource() resource.Resource {
	return &twingateResource{}
}

type twingateResource struct {
	client *client.Client
}

type resourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Address                  types.String `tfsdk:"address"`
	RemoteNetworkID          types.String `tfsdk:"remote_network_id"`
	IsAuthoritative          types.Bool   `tfsdk:"is_authoritative"`
	Protocols                types.Object `tfsdk:"protocols"`
	GroupAccess              types.Set    `tfsdk:"access_group"`
	ServiceAccess            types.Set    `tfsdk:"access_service"`
	IsActive                 types.Bool   `tfsdk:"is_active"`
	IsVisible                types.Bool   `tfsdk:"is_visible"`
	IsBrowserShortcutEnabled types.Bool   `tfsdk:"is_browser_shortcut_enabled"`
	Alias                    types.String `tfsdk:"alias"`
	SecurityPolicyID         types.String `tfsdk:"security_policy_id"`
	DLPPolicyID              types.String `tfsdk:"dlp_policy_id"`
}

func (r *twingateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateResource
}

func (r *twingateResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

func (r *twingateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(attr.ID), req, resp)

	res, err := r.client.ReadResource(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("failed to import state", err.Error())

		return
	}

	if res.Protocols != nil {
		protocols, diags := convertProtocolsToTerraform(res.Protocols, nil)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.State.SetAttribute(ctx, path.Root(attr.Protocols), protocols)
	}

	if len(res.GroupsAccess) > 0 {
		accessGroup, diags := convertGroupsAccessToTerraform(ctx, res.GroupsAccess)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.State.SetAttribute(ctx, path.Root(attr.AccessGroup), accessGroup)
	}

	if len(res.ServiceAccounts) > 0 {
		accessServiceAccount, diags := convertAccessServiceAccountsToTerraform(ctx, res.ServiceAccounts)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.State.SetAttribute(ctx, path.Root(attr.AccessService), accessServiceAccount)
	}
}

//nolint:funlen
func (r *twingateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     schemaVersion,
		Description: "Resources in Twingate represent servers on the private network that clients can connect to. Resources can be defined by IP, CIDR range, FQDN, or DNS zone. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
		Attributes: map[string]schema.Attribute{
			attr.Name: schema.StringAttribute{
				Required:    true,
				Description: "The name of the Resource",
			},
			attr.Address: schema.StringAttribute{
				Required:    true,
				Description: "The Resource's IP/CIDR or FQDN/DNS zone",
			},
			attr.RemoteNetworkID: schema.StringAttribute{
				Required:    true,
				Description: "Remote Network ID where the Resource lives",
			},
			// optional
			attr.IsActive: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Set the resource as active or inactive. Default is `true`.",
				Default:     booldefault.StaticBool(true),
			},
			attr.IsAuthoritative: schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to `false`, assignments made outside of Terraform will be ignored.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			attr.Alias: schema.StringAttribute{
				Optional:      true,
				Description:   "Set a DNS alias address for the Resource. Must be a DNS-valid name string.",
				PlanModifiers: []planmodifier.String{CaseInsensitiveDiff()},
			},
			attr.Protocols: protocols(),
			// computed
			attr.SecurityPolicyID: schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "The ID of a `twingate_security_policy` to set as this Resource's Security Policy. Default is `Default Policy`.",
				Default:       stringdefault.StaticString(DefaultSecurityPolicyID),
				PlanModifiers: []planmodifier.String{UseDefaultPolicyForUnknownModifier()},
			},
			attr.DLPPolicyID: schema.StringAttribute{
				Optional: true,
				//Computed:    true,
				Description: "The ID of a DLP policy to be used as the default DLP policy for this Resource. Defaults to null.",
				//Default:     stringdefault.StaticString(""),
				//PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			attr.IsVisible: schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Controls whether this Resource will be visible in the main Resource list in the Twingate Client. Default is `true`.",
				Default:       booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			attr.IsBrowserShortcutEnabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Controls whether an \"Open in Browser\" shortcut will be shown for this Resource in the Twingate Client. Default is `false`.",
				Default:     booldefault.StaticBool(false),
			},
			attr.ID: schema.StringAttribute{
				Computed:      true,
				Description:   "Autogenerated ID of the Resource, encoded in base64",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},

		Blocks: map[string]schema.Block{
			attr.AccessGroup:   groupAccessBlock(),
			attr.AccessService: serviceAccessBlock(),
		},
	}
}

func (r *twingateResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 1 (Schema.Version)
		0: upgradeResourceStateV0(),
		// State upgrade implementation from schema version 1 to 2
		1: upgradeResourceStateV1(),
	}
}

func protocols() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Computed: true,
		Default:  objectdefault.StaticValue(defaultProtocolsObject()),
		Attributes: map[string]schema.Attribute{
			attr.AllowIcmp: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to allow ICMP (ping) traffic",
			},

			attr.UDP: protocol(),
			attr.TCP: protocol(),
		},
		Description: "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction, and all protocols and ports are allowed.",
	}
}

func protocol() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Computed: true,
		Default:  objectdefault.StaticValue(defaultProtocolObject()),
		Attributes: map[string]schema.Attribute{
			attr.Policy: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(model.Policies...),
				},
				Default:     stringdefault.StaticString(model.PolicyAllowAll),
				Description: fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", model.PolicyRestricted, model.PolicyAllowAll, model.PolicyDenyAll),
			},
			attr.Ports: schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of port ranges between 1 and 65535 inclusive, in the format `100-200` for a range, or `8080` for a single port",
				PlanModifiers: []planmodifier.Set{
					PortsDiff(),
				},
				Default: setdefault.StaticValue(defaultEmptyPorts()),
			},
		},
	}
}

func groupAccessBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
		Description: "Restrict access to certain group",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attr.GroupID: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Group ID that will have permission to access the Resource.",
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`\w+`), "Group ID can't be empty"),
					},
				},
				attr.SecurityPolicyID: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The ID of a `twingate_security_policy` to use as the access policy for the group IDs in the access block.",
					Validators: []validator.String{
						stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName(attr.GroupID)),
					},
					PlanModifiers: []planmodifier.String{
						UseNullPolicyForGroupAccessWhenValueOmitted(),
					},
				},
				attr.UsageBasedAutolockDurationDays: schema.Int64Attribute{
					Optional:    true,
					Computed:    true,
					Description: "The usage-based auto-lock duration configured on the edge (in days).",
					Validators: []validator.Int64{
						int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName(attr.GroupID)),
					},
					PlanModifiers: []planmodifier.Int64{
						UseNullIntWhenValueOmitted(),
					},
				},
				attr.DLPPolicyID: schema.StringAttribute{
					Optional: true,
					//Computed:      true,
					Description: "The ID of a DLP policy to be used as the DLP policy for the group in this access block.",
					//PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				},
			},
		},
	}
}

func serviceAccessBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		Validators: []validator.Set{
			setvalidator.SizeAtLeast(1),
		},
		Description: "Restrict access to certain service account",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attr.ServiceAccountID: schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The ID of the service account that should have access to this Resource.",
					Validators: []validator.String{
						stringvalidator.RegexMatches(regexp.MustCompile(`\w+`), "ServiceAccount ID can't be empty"),
					},
				},
			},
		},
	}
}

func PortsDiff() planmodifier.Set {
	return portsDiff{}
}

type portsDiff struct{}

// Description returns a human-readable description of the plan modifier.
func (m portsDiff) Description(_ context.Context) string {
	return "Handles ports difference."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m portsDiff) MarkdownDescription(_ context.Context) string {
	return "Handles ports difference."
}

// PlanModifySet implements the plan modification logic.
func (m portsDiff) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.StateValue.IsNull() {
		return
	}

	if equalPorts(req.StateValue, req.PlanValue) {
		resp.PlanValue = req.StateValue
	}
}

func equalPorts(one, another types.Set) bool {
	oldPortsRange, err := convertPorts(one)
	if err != nil {
		return false
	}

	newPortsRange, err := convertPorts(another)
	if err != nil {
		return false
	}

	return portRangeEqual(oldPortsRange, newPortsRange)
}

func portRangeEqual(one, another []*model.PortRange) bool {
	oneMap := convertPortsRangeToMap(one)
	anotherMap := convertPortsRangeToMap(another)

	return reflect.DeepEqual(oneMap, anotherMap)
}

func convertPorts(list types.Set) ([]*model.PortRange, error) {
	items := list.Elements()

	var ports = make([]*model.PortRange, 0, len(items))

	for _, port := range items {
		portRange, err := model.NewPortRange(port.(types.String).ValueString())
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		ports = append(ports, portRange)
	}

	return ports, nil
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

func (r *twingateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input, err := convertResource(&plan)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, TwingateResource)

		return
	}

	resource, err := r.client.CreateResource(ctx, input)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, TwingateResource)

		return
	}

	if err = r.client.AddResourceAccess(ctx, resource.ID, convertResourceAccess(resource.ServiceAccounts, resource.GroupsAccess)); err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, TwingateResource)

		return
	}

	if !input.IsActive {
		if err := r.client.UpdateResourceActiveState(ctx, &model.Resource{
			ID:       resource.ID,
			IsActive: false,
		}); err != nil {
			addErr(&resp.Diagnostics, err, operationCreate, TwingateResource)

			return
		}

		resource.IsActive = false
	}

	r.helper(ctx, resource, &plan, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func convertResourceAccess(serviceAccounts []string, groupsAccess []model.AccessGroup) []client.AccessInput {
	access := make([]client.AccessInput, 0, len(serviceAccounts)+len(groupsAccess))
	for _, account := range serviceAccounts {
		access = append(access, client.AccessInput{PrincipalID: account})
	}

	for _, group := range groupsAccess {
		access = append(access, client.AccessInput{
			PrincipalID:                    group.GroupID,
			SecurityPolicyID:               group.SecurityPolicyID,
			DLPPolicyID:                    group.DLPPolicyID,
			UsageBasedAutolockDurationDays: group.UsageBasedDuration,
		})
	}

	return access
}

func getAccessAttribute(list types.List, attribute string) []string {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil
	}

	obj := list.Elements()[0].(types.Object)
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	val := obj.Attributes()[attribute]
	if val == nil || val.IsNull() || val.IsUnknown() {
		return nil
	}

	return convertIDs(val.(types.Set))
}

//nolint:cyclop
func getGroupAccessAttribute(list types.Set) []model.AccessGroup {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil
	}

	access := make([]model.AccessGroup, 0, len(list.Elements()))

	for _, item := range list.Elements() {
		obj := item.(types.Object)
		if obj.IsNull() || obj.IsUnknown() {
			continue
		}

		groupVal := obj.Attributes()[attr.GroupID]
		accessGroup := model.AccessGroup{
			GroupID: groupVal.(types.String).ValueString(),
		}

		securityPolicyVal := obj.Attributes()[attr.SecurityPolicyID]
		if securityPolicyVal != nil && !securityPolicyVal.IsNull() && !securityPolicyVal.IsUnknown() {
			accessGroup.SecurityPolicyID = securityPolicyVal.(types.String).ValueStringPointer()
		}

		dlpPolicyVal := obj.Attributes()[attr.DLPPolicyID]
		if dlpPolicyVal != nil && !dlpPolicyVal.IsNull() && !dlpPolicyVal.IsUnknown() {
			accessGroup.DLPPolicyID = dlpPolicyVal.(types.String).ValueStringPointer()
		}

		usageBasedDuration := obj.Attributes()[attr.UsageBasedAutolockDurationDays]
		if usageBasedDuration != nil && !usageBasedDuration.IsNull() && !usageBasedDuration.IsUnknown() {
			accessGroup.UsageBasedDuration = usageBasedDuration.(types.Int64).ValueInt64Pointer()
		}

		access = append(access, accessGroup)
	}

	return access
}

func getServiceAccountAccessAttribute(list types.Set) []string {
	if list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return nil
	}

	serviceAccountIDs := make([]string, 0, len(list.Elements()))

	for _, item := range list.Elements() {
		obj := item.(types.Object)
		if obj.IsNull() || obj.IsUnknown() {
			continue
		}

		val := obj.Attributes()[attr.ServiceAccountID]
		if val == nil || val.IsNull() || val.IsUnknown() {
			continue
		}

		serviceAccountIDs = append(serviceAccountIDs, val.(types.String).ValueString())
	}

	return serviceAccountIDs
}

//nolint:cyclop
func convertResource(plan *resourceModel) (*model.Resource, error) {
	protocols, err := convertProtocols(&plan.Protocols)
	if err != nil {
		return nil, err
	}

	accessGroups := getGroupAccessAttribute(plan.GroupAccess)
	serviceAccountIDs := getServiceAccountAccessAttribute(plan.ServiceAccess)

	for _, access := range accessGroups {
		if access.SecurityPolicyID == nil && access.UsageBasedDuration == nil && len(strings.TrimSpace(access.GroupID)) == 0 {
			return nil, ErrInvalidAttributeCombination
		}

		if err := checkGlobalID(access.GroupID); err != nil {
			return nil, err
		}
	}

	for _, id := range serviceAccountIDs {
		if err := checkGlobalID(id); err != nil {
			return nil, err
		}
	}

	isBrowserShortcutEnabled := getOptionalBool(plan.IsBrowserShortcutEnabled)

	if isBrowserShortcutEnabled != nil && *isBrowserShortcutEnabled && isWildcardAddress(plan.Address.ValueString()) {
		return nil, ErrWildcardAddressWithEnabledShortcut
	}

	return &model.Resource{
		Name:                     plan.Name.ValueString(),
		RemoteNetworkID:          plan.RemoteNetworkID.ValueString(),
		Address:                  plan.Address.ValueString(),
		Protocols:                protocols,
		GroupsAccess:             accessGroups,
		ServiceAccounts:          serviceAccountIDs,
		IsActive:                 plan.IsActive.ValueBool(),
		IsAuthoritative:          convertAuthoritativeFlag(plan.IsAuthoritative),
		Alias:                    getOptionalString(plan.Alias),
		IsVisible:                getOptionalBool(plan.IsVisible),
		IsBrowserShortcutEnabled: isBrowserShortcutEnabled,
		SecurityPolicyID:         plan.SecurityPolicyID.ValueStringPointer(),
		DLPPolicyID:              plan.DLPPolicyID.ValueStringPointer(),
	}, nil
}

func checkGlobalID(val string) error {
	const expectedTokens = 2

	data, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return ErrWrongGlobalID
	}

	tokens := strings.Split(string(data), ":")
	if len(tokens) != expectedTokens {
		return ErrWrongGlobalID
	}

	name := tokens[0]

	if name != "Group" && name != "ServiceAccount" {
		return ErrWrongGlobalID
	}

	return nil
}

func getOptionalBool(val types.Bool) *bool {
	if !val.IsUnknown() {
		return val.ValueBoolPointer()
	}

	return nil
}

func getOptionalString(val types.String) *string {
	if !val.IsUnknown() && !val.IsNull() {
		return val.ValueStringPointer()
	}

	return nil
}

func convertIDs(list types.Set) []string {
	return utils.Map(list.Elements(), func(item tfattr.Value) string {
		return item.(types.String).ValueString()
	})
}

func equalProtocolsState(objA, objB *types.Object) bool {
	if objA.IsNull() != objB.IsNull() || objA.IsUnknown() != objB.IsUnknown() {
		return false
	}

	protocolsA, err := convertProtocols(objA)
	if err != nil {
		return false
	}

	protocolsB, err := convertProtocols(objB)
	if err != nil {
		return false
	}

	return equalProtocols(protocolsA, protocolsB)
}

func equalProtocols(one, another *model.Protocols) bool {
	return one.AllowIcmp == another.AllowIcmp && equalProtocol(one.TCP, another.TCP) && equalProtocol(one.UDP, another.UDP)
}

func equalProtocol(one, another *model.Protocol) bool {
	return one.Policy == another.Policy && portRangeEqual(one.Ports, another.Ports)
}

func convertProtocols(protocols *types.Object) (*model.Protocols, error) {
	if protocols == nil || protocols.IsNull() || protocols.IsUnknown() {
		return model.DefaultProtocols(), nil
	}

	udp, err := convertProtocol(protocols.Attributes()[attr.UDP])
	if err != nil {
		return nil, err
	}

	tcp, err := convertProtocol(protocols.Attributes()[attr.TCP])
	if err != nil {
		return nil, err
	}

	return &model.Protocols{
		AllowIcmp: protocols.Attributes()[attr.AllowIcmp].(types.Bool).ValueBool(),
		UDP:       udp,
		TCP:       tcp,
	}, nil
}

func convertProtocol(protocol tfattr.Value) (*model.Protocol, error) {
	obj := convertProtocolObj(protocol)
	if obj.IsNull() {
		return nil, nil //nolint:nilnil
	}

	ports, err := decodePorts(obj)
	if err != nil {
		return nil, err
	}

	policy := obj.Attributes()[attr.Policy].(types.String).ValueString()
	if err := isValidPolicy(policy, ports); err != nil {
		return nil, err
	}

	if policy == model.PolicyDenyAll {
		policy = model.PolicyRestricted
	}

	return model.NewProtocol(policy, ports), nil
}

func convertProtocolObj(protocol tfattr.Value) types.Object {
	if protocol == nil || protocol.IsNull() {
		return types.ObjectNull(nil)
	}

	obj, ok := protocol.(types.Object)
	if !ok || obj.IsNull() {
		return types.ObjectNull(nil)
	}

	return obj
}

func decodePorts(obj types.Object) ([]*model.PortRange, error) {
	portsVal := obj.Attributes()[attr.Ports]
	if portsVal == nil || portsVal.IsNull() {
		return nil, nil
	}

	portsList, ok := portsVal.(types.Set)
	if !ok {
		return nil, nil
	}

	return convertPorts(portsList)
}

func isValidPolicy(policy string, ports []*model.PortRange) error {
	switch policy {
	case model.PolicyAllowAll:
		if len(ports) > 0 {
			return ErrPortsWithPolicyAllowAll
		}

	case model.PolicyDenyAll:
		if len(ports) > 0 {
			return ErrPortsWithPolicyDenyAll
		}

	case model.PolicyRestricted:
		if len(ports) == 0 {
			return ErrPolicyRestrictedWithoutPorts
		}
	}

	return nil
}

func (r *twingateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := r.client.ReadResource(ctx, state.ID.ValueString())
	if resource != nil {
		resource.IsAuthoritative = convertAuthoritativeFlag(state.IsAuthoritative)

		emptyPolicy := ""

		if state.SecurityPolicyID.ValueString() == "" {
			resource.SecurityPolicyID = &emptyPolicy
		}

		if state.DLPPolicyID.ValueString() == "" {
			resource.DLPPolicyID = &emptyPolicy
		}
	}

	r.helper(ctx, resource, &state, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

//nolint:cyclop
func (r *twingateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input, err := convertResource(&plan)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

		return
	}

	planSecurityPolicy := input.SecurityPolicyID
	input.ID = state.ID.ValueString()

	if !plan.GroupAccess.Equal(state.GroupAccess) || !plan.ServiceAccess.Equal(state.ServiceAccess) {
		if err := r.updateResourceAccess(ctx, &plan, &state, input); err != nil {
			addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

			return
		}
	}

	var resource *model.Resource

	if isResourceChanged(&plan, &state) {
		if err := r.setDefaultSecurityPolicy(ctx, input); err != nil {
			addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

			return
		}

		resource, err = r.client.UpdateResource(ctx, input)
	} else {
		resource, err = r.client.ReadResource(ctx, input.ID)
	}

	if resource != nil {
		resource.IsAuthoritative = input.IsAuthoritative

		if planSecurityPolicy != nil && *planSecurityPolicy == "" {
			resource.SecurityPolicyID = planSecurityPolicy
		}
	}

	r.helper(ctx, resource, &state, &plan, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func (r *twingateResource) setDefaultSecurityPolicy(ctx context.Context, resource *model.Resource) error {
	if DefaultSecurityPolicyID == "" {
		policy, _ := r.client.ReadSecurityPolicy(ctx, "", DefaultSecurityPolicyName)
		if policy != nil {
			DefaultSecurityPolicyID = policy.ID
		}
	}

	if DefaultSecurityPolicyID == "" {
		return ErrDefaultPolicyNotSet
	}

	remoteResource, err := r.client.ReadResource(ctx, resource.ID)
	if err != nil {
		return err //nolint:wrapcheck
	}

	if remoteResource.SecurityPolicyID != nil && (resource.SecurityPolicyID == nil || *resource.SecurityPolicyID == "") &&
		*remoteResource.SecurityPolicyID != DefaultSecurityPolicyID {
		resource.SecurityPolicyID = &DefaultSecurityPolicyID
	}

	return nil
}

func isResourceChanged(plan, state *resourceModel) bool {
	return !plan.RemoteNetworkID.Equal(state.RemoteNetworkID) ||
		!plan.Name.Equal(state.Name) ||
		!plan.Address.Equal(state.Address) ||
		!equalProtocolsState(&plan.Protocols, &state.Protocols) ||
		!plan.IsActive.Equal(state.IsActive) ||
		!plan.IsVisible.Equal(state.IsVisible) ||
		!plan.IsBrowserShortcutEnabled.Equal(state.IsBrowserShortcutEnabled) ||
		!plan.Alias.Equal(state.Alias) ||
		!plan.SecurityPolicyID.Equal(state.SecurityPolicyID)
}

func (r *twingateResource) updateResourceAccess(ctx context.Context, plan, state *resourceModel, input *model.Resource) error {
	idsToDelete, serviceAccountsToAdd, groupsToAdd, err := r.getChangedAccessIDs(ctx, plan, state, input)
	if err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	if err := r.client.RemoveResourceAccess(ctx, input.ID, idsToDelete); err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	if err := r.client.SetResourceAccess(ctx, input.ID, convertResourceAccess(serviceAccountsToAdd, groupsToAdd)); err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	return nil
}

func (r *twingateResource) getChangedAccessIDs(ctx context.Context, plan, state *resourceModel, resource *model.Resource) ([]string, []string, []model.AccessGroup, error) {
	remote, err := r.client.ReadResource(ctx, resource.ID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get changedIDs: %w", err)
	}

	var (
		oldServiceAccounts []string
		oldGroups          []model.AccessGroup
	)

	if resource.IsAuthoritative {
		oldGroups, oldServiceAccounts = remote.GroupsAccess, remote.ServiceAccounts
	} else {
		oldGroups = getOldIDsNonAuthoritativeGroupAccess(plan, state)
		oldServiceAccounts = getOldIDsNonAuthoritativeServiceAccountAccess(plan, state)
	}

	// ids to delete
	groupsToDelete := setDifferenceGroups(oldGroups, resource.GroupsAccess)
	serviceAccountsToDelete := setDifference(oldServiceAccounts, resource.ServiceAccounts)

	// ids to add
	groupsToAdd := setDifferenceGroupAccess(resource.GroupsAccess, remote.GroupsAccess)
	serviceAccountsToAdd := setDifference(resource.ServiceAccounts, remote.ServiceAccounts)

	return append(groupsToDelete, serviceAccountsToDelete...), serviceAccountsToAdd, groupsToAdd, nil
}

func getOldIDsNonAuthoritativeServiceAccountAccess(plan, state *resourceModel) []string {
	if !plan.ServiceAccess.Equal(state.ServiceAccess) {
		return getServiceAccountAccessAttribute(state.ServiceAccess)
	}

	return nil
}

func getOldIDsNonAuthoritativeGroupAccess(plan, state *resourceModel) []model.AccessGroup {
	if !plan.GroupAccess.Equal(state.GroupAccess) {
		return getGroupAccessAttribute(state.GroupAccess)
	}

	return nil
}

func (r *twingateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResource(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationDelete, TwingateResource)
}

func (r *twingateResource) helper(ctx context.Context, resource *model.Resource, state, reference *resourceModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			respState.RemoveResource(ctx)

			return
		}

		addErr(diagnostics, err, operation, TwingateResource)

		return
	}

	if resource.Protocols == nil {
		resource.Protocols = model.DefaultProtocols()
	}

	if !resource.IsAuthoritative {
		resource.GroupsAccess = setIntersectionGroupAccess(getGroupAccessAttribute(reference.GroupAccess), resource.GroupsAccess)
		resource.ServiceAccounts = setIntersection(getServiceAccountAccessAttribute(reference.ServiceAccess), resource.ServiceAccounts)

		serviceAccessIDs := utils.MakeLookupMap(getServiceAccountAccessAttribute(reference.ServiceAccess))

		var filteredServiceAccess []string

		for _, serviceAccessID := range resource.ServiceAccounts {
			if serviceAccessIDs[serviceAccessID] {
				filteredServiceAccess = append(filteredServiceAccess, serviceAccessID)
			}
		}

		resource.ServiceAccounts = filteredServiceAccess
	}

	setState(ctx, state, reference, resource, diagnostics)

	if diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diagnostics.Append(respState.Set(ctx, state)...)
}

func setState(ctx context.Context, state, reference *resourceModel, resource *model.Resource, diagnostics *diag.Diagnostics) { //nolint:cyclop
	state.ID = types.StringValue(resource.ID)
	state.Name = types.StringValue(resource.Name)
	state.RemoteNetworkID = types.StringValue(resource.RemoteNetworkID)
	state.Address = types.StringValue(resource.Address)
	state.IsActive = types.BoolValue(resource.IsActive)
	state.IsAuthoritative = types.BoolValue(resource.IsAuthoritative)
	state.SecurityPolicyID = types.StringPointerValue(resource.SecurityPolicyID)

	if !state.IsVisible.IsNull() || !reference.IsVisible.IsUnknown() {
		state.IsVisible = types.BoolPointerValue(resource.IsVisible)
	}

	if !state.IsBrowserShortcutEnabled.IsNull() || !reference.IsBrowserShortcutEnabled.IsUnknown() {
		state.IsBrowserShortcutEnabled = types.BoolPointerValue(resource.IsBrowserShortcutEnabled)
	}

	if !state.Alias.IsNull() || !reference.Alias.IsUnknown() {
		state.Alias = reference.Alias
	}

	if !state.DLPPolicyID.IsNull() || !reference.DLPPolicyID.IsUnknown() {
		state.DLPPolicyID = reference.DLPPolicyID
	}

	if !state.Protocols.IsNull() || !reference.Protocols.IsUnknown() {
		protocols, diags := convertProtocolsToTerraform(resource.Protocols, &reference.Protocols)
		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		if !equalProtocolsState(&state.Protocols, &protocols) {
			state.Protocols = protocols
		}
	}

	groupAccess, diags := convertGroupsAccessToTerraform(ctx, resource.GroupsAccess)
	diagnostics.Append(diags...)
	serviceAccess, diags := convertServiceAccessToTerraform(ctx, resource.ServiceAccounts)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return
	}

	state.GroupAccess = groupAccess
	state.ServiceAccess = serviceAccess
}

func convertProtocolsToTerraform(protocols *model.Protocols, reference *types.Object) (types.Object, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if protocols == nil || reference != nil && (reference.IsUnknown() || reference.IsNull()) {
		return defaultProtocolsModelToTerraform()
	}

	var referenceTCP, referenceUDP tfattr.Value
	if reference != nil {
		referenceTCP = reference.Attributes()[attr.TCP]
		referenceUDP = reference.Attributes()[attr.UDP]
	}

	tcp, diags := convertProtocolModelToTerraform(protocols.TCP, referenceTCP)
	diagnostics.Append(diags...)

	udp, diags := convertProtocolModelToTerraform(protocols.UDP, referenceUDP)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return types.ObjectNull(protocolsAttributeTypes()), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.AllowIcmp: types.BoolValue(protocols.AllowIcmp),
		attr.TCP:       tcp,
		attr.UDP:       udp,
	}

	obj := types.ObjectValueMust(protocolsAttributeTypes(), attributes)

	return obj, diagnostics
}

func convertPortsToTerraform(ports []*model.PortRange) types.Set {
	if len(ports) == 0 {
		return defaultEmptyPorts()
	}

	elements := make([]tfattr.Value, 0, len(ports))
	for _, port := range ports {
		elements = append(elements, types.StringValue(port.String()))
	}

	return types.SetValueMust(types.StringType, elements)
}

func convertProtocolModelToTerraform(protocol *model.Protocol, _ tfattr.Value) (types.Object, diag.Diagnostics) {
	if protocol == nil {
		return types.ObjectNull(protocolAttributeTypes()), nil
	}

	ports := convertPortsToTerraform(protocol.Ports)

	policy := protocol.Policy
	if policy == model.PolicyRestricted && len(ports.Elements()) == 0 {
		policy = model.PolicyDenyAll
	}

	attributes := map[string]tfattr.Value{
		attr.Policy: types.StringValue(policy),
		attr.Ports:  ports,
	}

	return types.ObjectValue(protocolAttributeTypes(), attributes)
}

func defaultProtocolsModelToTerraform() (types.Object, diag.Diagnostics) {
	attributeTypes := protocolsAttributeTypes()

	var diagnostics diag.Diagnostics

	defaultPorts, diags := defaultProtocolModelToTerraform()
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeNullObject(attributeTypes), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.AllowIcmp: types.BoolValue(true),
		attr.TCP:       defaultPorts,
		attr.UDP:       defaultPorts,
	}

	obj, diags := types.ObjectValue(attributeTypes, attributes)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeNullObject(attributeTypes), diagnostics
	}

	return obj, diagnostics
}

func defaultProtocolsObject() types.Object {
	attributeTypes := protocolsAttributeTypes()

	var diagnostics diag.Diagnostics

	defaultPorts, diags := defaultProtocolModelToTerraform()
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeNullObject(attributeTypes)
	}

	attributes := map[string]tfattr.Value{
		attr.AllowIcmp: types.BoolValue(true),
		attr.TCP:       defaultPorts,
		attr.UDP:       defaultPorts,
	}

	obj, diags := types.ObjectValue(attributeTypes, attributes)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeNullObject(attributeTypes)
	}

	return obj
}

func defaultEmptyPorts() types.Set {
	return types.SetValueMust(types.StringType, []tfattr.Value{})
}

func defaultProtocolModelToTerraform() (basetypes.ObjectValue, diag.Diagnostics) {
	attributes := map[string]tfattr.Value{
		attr.Policy: types.StringValue(model.PolicyAllowAll),
		attr.Ports:  defaultEmptyPorts(),
	}

	return types.ObjectValue(protocolAttributeTypes(), attributes)
}

func defaultProtocolObject() basetypes.ObjectValue {
	obj, _ := defaultProtocolModelToTerraform()

	return obj
}

func protocolsAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.AllowIcmp: types.BoolType,
		attr.TCP: types.ObjectType{
			AttrTypes: protocolAttributeTypes(),
		},
		attr.UDP: types.ObjectType{
			AttrTypes: protocolAttributeTypes(),
		},
	}
}

func protocolAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.Policy: types.StringType,
		attr.Ports: types.SetType{
			ElemType: types.StringType,
		},
	}
}

func convertServiceAccessToTerraform(ctx context.Context, serviceAccounts []string) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(serviceAccounts) == 0 {
		return makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()), diagnostics
	}

	objects := make([]types.Object, 0, len(serviceAccounts))

	for _, account := range serviceAccounts {
		attributes := map[string]tfattr.Value{
			attr.ServiceAccountID: types.StringValue(account),
		}

		obj, diags := types.ObjectValue(accessServiceAccountAttributeTypes(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessServiceAccountAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}

func convertGroupsAccessToTerraform(ctx context.Context, groupAccess []model.AccessGroup) (types.Set, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(groupAccess) == 0 {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	objects := make([]types.Object, 0, len(groupAccess))

	for _, access := range groupAccess {
		attributes := map[string]tfattr.Value{
			attr.GroupID:                        types.StringValue(access.GroupID),
			attr.SecurityPolicyID:               types.StringPointerValue(access.SecurityPolicyID),
			attr.DLPPolicyID:                    types.StringPointerValue(access.DLPPolicyID),
			attr.UsageBasedAutolockDurationDays: types.Int64PointerValue(access.UsageBasedDuration),
		}

		obj, diags := types.ObjectValue(accessGroupAttributeTypes(), attributes)
		diagnostics.Append(diags...)

		objects = append(objects, obj)
	}

	if diagnostics.HasError() {
		return makeObjectsSetNull(ctx, accessGroupAttributeTypes()), diagnostics
	}

	return makeObjectsSet(ctx, objects...)
}

func CaseInsensitiveDiff() planmodifier.String {
	return caseInsensitiveDiffModifier{
		description: "Handles case insensitive strings",
	}
}

type caseInsensitiveDiffModifier struct {
	description string
}

func (m caseInsensitiveDiffModifier) Description(_ context.Context) string {
	return m.description
}

func (m caseInsensitiveDiffModifier) MarkdownDescription(_ context.Context) string {
	return m.description
}

func (m caseInsensitiveDiffModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do not replace on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do not replace on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	if !req.PlanValue.IsUnknown() && req.StateValue.IsNull() {
		return
	}

	if strings.EqualFold(strings.ToLower(req.PlanValue.ValueString()), strings.ToLower(req.StateValue.ValueString())) {
		resp.PlanValue = req.StateValue
	}
}

var cidrRgxp = regexp.MustCompile(`(\d{1,3}\.){3}\d{1,3}(/\d+)`)

func isWildcardAddress(address string) bool {
	return strings.ContainsAny(address, "*?") || cidrRgxp.MatchString(address)
}

func UseDefaultPolicyForUnknownModifier() planmodifier.String {
	return useDefaultPolicyForUnknownModifier{}
}

// useDefaultPolicyForUnknownModifier implements the plan modifier.
type useDefaultPolicyForUnknownModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m useDefaultPolicyForUnknownModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute will fallback to Default Policy on unset."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useDefaultPolicyForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute will fallback to Default Policy on unset."
}

// PlanModifyString implements the plan modification logic.
func (m useDefaultPolicyForUnknownModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() && req.ConfigValue.IsNull() {
		resp.PlanValue = types.StringPointerValue(nil)

		return
	}

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	// Do nothing if there is a known planned value.
	if req.ConfigValue.ValueString() != "" {
		return
	}

	if req.StateValue.ValueString() == "" && req.PlanValue.ValueString() == DefaultSecurityPolicyID {
		resp.PlanValue = types.StringValue("")
	} else if req.StateValue.ValueString() == DefaultSecurityPolicyID && req.PlanValue.ValueString() == "" {
		resp.PlanValue = types.StringValue(DefaultSecurityPolicyID)
	}
}

func accessGroupAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.GroupID:                        types.StringType,
		attr.SecurityPolicyID:               types.StringType,
		attr.DLPPolicyID:                    types.StringType,
		attr.UsageBasedAutolockDurationDays: types.Int64Type,
	}
}

func accessServiceAccountAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.ServiceAccountID: types.StringType,
	}
}

func UseNullPolicyForGroupAccessWhenValueOmitted() planmodifier.String {
	return useNullPolicyForGroupAccessWhenValueOmitted{}
}

type useNullPolicyForGroupAccessWhenValueOmitted struct{}

func (m useNullPolicyForGroupAccessWhenValueOmitted) Description(_ context.Context) string {
	return ""
}

func (m useNullPolicyForGroupAccessWhenValueOmitted) MarkdownDescription(_ context.Context) string {
	return ""
}

func (m useNullPolicyForGroupAccessWhenValueOmitted) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() && req.ConfigValue.IsNull() {
		resp.PlanValue = types.StringNull()

		return
	}

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.IsNull() && !req.PlanValue.IsNull() {
		resp.PlanValue = types.StringNull()
	}
}

func UseNullIntWhenValueOmitted() planmodifier.Int64 {
	return useNullIntWhenValueOmitted{}
}

type useNullIntWhenValueOmitted struct{}

func (m useNullIntWhenValueOmitted) Description(_ context.Context) string {
	return ""
}

func (m useNullIntWhenValueOmitted) MarkdownDescription(_ context.Context) string {
	return ""
}

func (m useNullIntWhenValueOmitted) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.StateValue.IsNull() && req.ConfigValue.IsNull() {
		resp.PlanValue = types.Int64Null()

		return
	}

	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.IsNull() && !req.PlanValue.IsNull() {
		resp.PlanValue = types.Int64Null()
	}
}
