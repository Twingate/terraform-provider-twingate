package resource

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	schemaVersion             int64 = 4
)

var (
	ErrPortsWithPolicyAllowAll            = errors.New(model.PolicyAllowAll + " policy does not allow specifying ports.")
	ErrPortsWithPolicyDenyAll             = errors.New(model.PolicyDenyAll + " policy does not allow specifying ports.")
	ErrPolicyRestrictedWithoutPorts       = errors.New(model.PolicyRestricted + " policy requires specifying ports.")
	ErrInvalidAttributeCombination        = errors.New("invalid attribute combination")
	ErrWildcardAddressWithEnabledShortcut = errors.New("Resources with a CIDR range or wildcard can't have the browser shortcut enabled.")
	ErrWrongGlobalID                      = errors.New("Unable to parse global ID")
	ErrBypassRoutingWithWildcardAddress   = errors.New("Bypass Resources cannot have a wildcard address")
	ErrBypassRoutingWithSecurityPolicy    = errors.New("Bypass Resources cannot have a security policy")
	ErrBypassRoutingWithPortRestriction   = errors.New("Bypass Resources cannot have port restrictions")
)

// Ensure the implementation satisfies the desired interfaces.
var _ resource.Resource = &twingateResource{}

func NewResourceResource() resource.Resource {
	return &twingateResource{}
}

type twingateResource struct {
	baseResource
}

type resourceModel struct {
	baseResourceModel

	Protocols                types.Object `tfsdk:"protocols"`
	IsActive                 types.Bool   `tfsdk:"is_active"`
	IsBrowserShortcutEnabled types.Bool   `tfsdk:"is_browser_shortcut_enabled"`
	RoutingMode              types.String `tfsdk:"routing_mode"`
}

func (r *twingateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateResource
}

func (r *twingateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(attr.ID), req, resp)

	res, err := r.client.ReadResource(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("failed to import state", err.Error())

		return
	}

	resp.State.SetAttribute(ctx, path.Root(attr.RoutingMode), types.StringPointerValue(res.RoutingMode))

	if res.Protocols != nil {
		protocols, diags := convertProtocolsToTerraform(res.Protocols, nil)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.State.SetAttribute(ctx, path.Root(attr.Protocols), protocols)
	}

	setImportedBaseResource(ctx, &resp.State, &resp.Diagnostics, res.BaseResource, r.defaultTags)
}

func (r *twingateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     schemaVersion,
		Description: "Resources in Twingate represent servers on the private network that clients can connect to. Resources can be defined by IP, CIDR range, FQDN, or DNS zone. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
		Attributes: withBaseAttributes(map[string]schema.Attribute{
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
			attr.Protocols: protocols(),
			attr.IsBrowserShortcutEnabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Controls whether an \"Open in Browser\" shortcut will be shown for this Resource in the Twingate Client. Default is `false`.",
				Default:     booldefault.StaticBool(false),
			},
			attr.RoutingMode: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Controls whether traffic to this Resource is routed through Twingate or bypassed. Valid values are `" + model.RoutingModeThroughTwingate + "` (default) and `" + model.RoutingModeBypassTwingate + "`. `" + model.RoutingModeBypassTwingate + "` requires a Resource with no security policy, a non-wildcard address and cannot have port restrictions.",
				Default:     stringdefault.StaticString(model.RoutingModeThroughTwingate),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(model.RoutingModes...),
				},
			},
			attr.ID: schema.StringAttribute{
				Computed:      true,
				Description:   "Autogenerated ID of the Resource, encoded in base64",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		}),

		Blocks: accessBlocks(),
	}
}

func (r *twingateResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from schema v0 to latest
		0: upgradeResourceStateV0(),
		// State upgrade implementation from schema v1 to latest
		1: upgradeResourceStateV1(),
		// State upgrade implementation from schema v2 to latest
		2: upgradeResourceStateV2(),
		// State upgrade implementation from schema v3 to latest
		3: upgradeResourceStateV3(),
	}
}

func protocols(optionalDeprecationMessage ...string) schema.SingleNestedAttribute {
	var deprecationMessage string
	if len(optionalDeprecationMessage) > 0 {
		deprecationMessage = optionalDeprecationMessage[0]
	}

	return schema.SingleNestedAttribute{
		Optional:           true,
		Computed:           true,
		Default:            objectdefault.StaticValue(defaultProtocolsObject()),
		DeprecationMessage: deprecationMessage,
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

	if err = r.client.AddResourceAccess(ctx, resource.ID, client.NewAccessInputs(resource.ServiceAccounts, resource.GroupsAccess)); err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, TwingateResource)

		return
	}

	if !input.IsActive {
		if err := r.client.UpdateResourceActiveState(ctx, &model.Resource{
			BaseResource: model.BaseResource{
				ID: resource.ID,
			},
			IsActive: false,
		}); err != nil {
			addErr(&resp.Diagnostics, err, operationCreate, TwingateResource)

			return
		}

		resource.IsActive = false
	}

	r.helper(ctx, resource, &plan, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func convertResource(plan *resourceModel) (*model.Resource, error) {
	protocols, err := convertProtocols(&plan.Protocols)
	if err != nil {
		return nil, err
	}

	base, err := convertBaseResource(&plan.baseResourceModel)
	if err != nil {
		return nil, err
	}

	isBrowserShortcutEnabled := getOptionalBool(plan.IsBrowserShortcutEnabled)

	if err := validateResourceConstraints(plan, base.GroupsAccess, protocols, isBrowserShortcutEnabled); err != nil {
		return nil, err
	}

	return &model.Resource{
		BaseResource:             base,
		Protocols:                protocols,
		IsActive:                 plan.IsActive.ValueBool(),
		IsAuthoritative:          convertAuthoritativeFlag(plan.IsAuthoritative),
		IsBrowserShortcutEnabled: isBrowserShortcutEnabled,
		RoutingMode:              plan.RoutingMode.ValueStringPointer(),
	}, nil
}

func getKeyValueMap(typesMap types.Map) map[string]string {
	if typesMap.IsNull() || typesMap.IsUnknown() || len(typesMap.Elements()) == 0 {
		return nil
	}

	keyValueMap := make(map[string]string, len(typesMap.Elements()))

	for key, val := range typesMap.Elements() {
		keyValueMap[key] = val.(types.String).ValueString()
	}

	return keyValueMap
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
	}

	r.helper(ctx, resource, &state, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

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

	input.ID = state.ID.ValueString()

	if err := updateResourceAccess(ctx, r.client, input.ID, input.IsAuthoritative, plan.accessModel, state.accessModel, baseResourceAccess(input.BaseResource), r.readAccess(input.ID)); err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

		return
	}

	var resource *model.Resource

	if isResourceChanged(&plan, &state) {
		resource, err = r.client.UpdateResource(ctx, input)
	} else {
		resource, err = r.client.ReadResource(ctx, input.ID)
	}

	if resource != nil {
		resource.IsAuthoritative = input.IsAuthoritative
	}

	r.helper(ctx, resource, &state, &plan, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func isResourceChanged(plan, state *resourceModel) bool {
	return isBaseResourceChanged(&plan.baseResourceModel, &state.baseResourceModel) ||
		!equalProtocolsState(&plan.Protocols, &state.Protocols) ||
		!plan.IsActive.Equal(state.IsActive) ||
		!plan.IsBrowserShortcutEnabled.Equal(state.IsBrowserShortcutEnabled) ||
		!plan.RoutingMode.Equal(state.RoutingMode)
}

func (r *twingateResource) readAccess(resourceID string) func(ctx context.Context) (resourceAccess, error) {
	return func(ctx context.Context) (resourceAccess, error) {
		remote, err := r.client.ReadResource(ctx, resourceID)
		if err != nil {
			return resourceAccess{}, err //nolint:wrapcheck
		}

		return baseResourceAccess(remote.BaseResource), nil
	}
}

func (r *twingateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	deleteResource(ctx, req, resp, TwingateResource, r.client.DeleteResource)
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

	setState(ctx, state, reference, resource, diagnostics)

	if diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diagnostics.Append(respState.Set(ctx, state)...)
}

func setState(ctx context.Context, state, reference *resourceModel, resource *model.Resource, diagnostics *diag.Diagnostics) {
	// reference fields are read up front: state and reference may be the same model
	referenceProtocols := reference.Protocols
	referenceIsBrowserShortcutEnabled := reference.IsBrowserShortcutEnabled

	setBaseResourceState(ctx, &state.baseResourceModel, &reference.baseResourceModel, resource.BaseResource, resource.IsAuthoritative, diagnostics)

	if diagnostics.HasError() {
		return
	}

	state.IsActive = types.BoolValue(resource.IsActive)
	state.RoutingMode = types.StringPointerValue(resource.RoutingMode)

	if !state.IsBrowserShortcutEnabled.IsNull() || !referenceIsBrowserShortcutEnabled.IsUnknown() {
		state.IsBrowserShortcutEnabled = types.BoolPointerValue(resource.IsBrowserShortcutEnabled)
	}

	if !state.Protocols.IsNull() || !referenceProtocols.IsUnknown() {
		protocols, diags := convertProtocolsToTerraform(resource.Protocols, &referenceProtocols)
		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		if !equalProtocolsState(&state.Protocols, &protocols) {
			state.Protocols = protocols
		}
	}
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

func convertProtocolModelToTerraform(protocol *model.Protocol, reference tfattr.Value) (types.Object, diag.Diagnostics) {
	if protocol == nil {
		return types.ObjectNull(protocolAttributeTypes()), nil
	}

	refProtocol, err := convertProtocol(reference)
	if err != nil {
		var diags diag.Diagnostics

		diags.AddError("failed to convert protocol", err.Error())

		return types.ObjectNull(protocolAttributeTypes()), diags
	}

	var ports types.Set
	if refProtocol != nil && portRangeEqual(protocol.Ports, refProtocol.Ports) {
		ports = convertPortsToTerraform(refProtocol.Ports)
	} else {
		ports = convertPortsToTerraform(protocol.Ports)
	}

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

var cidrRgxp = regexp.MustCompile(`(\d{1,3}\.){3}\d{1,3}(/\d+)`)

func isWildcardAddress(address string) bool {
	return strings.ContainsAny(address, "*?") || cidrRgxp.MatchString(address)
}

// validateResourceConstraints runs the apply-time cross-attribute checks that the
// schema alone can't express (browser-shortcut vs. wildcard, and the BYPASS_TWINGATE
// routing-mode constraints).
func validateResourceConstraints(plan *resourceModel, accessGroups []model.AccessGroup, protocols *model.Protocols, isBrowserShortcutEnabled *bool) error {
	if isBrowserShortcutEnabled != nil && *isBrowserShortcutEnabled && isWildcardAddress(plan.Address.ValueString()) {
		return ErrWildcardAddressWithEnabledShortcut
	}

	return validateBypassRoutingMode(plan.RoutingMode.ValueStringPointer(), plan.Address.ValueString(), plan.SecurityPolicyID, accessGroups, protocols)
}

func validateBypassRoutingMode(routingMode *string, address string, securityPolicyID types.String, accessGroups []model.AccessGroup, protocols *model.Protocols) error {
	if routingMode == nil || *routingMode != model.RoutingModeBypassTwingate {
		return nil
	}

	if strings.ContainsAny(address, "*?") {
		return ErrBypassRoutingWithWildcardAddress
	}

	if !securityPolicyID.IsNull() && !securityPolicyID.IsUnknown() && securityPolicyID.ValueString() != "" {
		return ErrBypassRoutingWithSecurityPolicy
	}

	for _, group := range accessGroups {
		if group.SecurityPolicyID != nil && *group.SecurityPolicyID != "" {
			return ErrBypassRoutingWithSecurityPolicy
		}
	}

	if hasPortRestriction(protocols) {
		return ErrBypassRoutingWithPortRestriction
	}

	return nil
}

// hasPortRestriction reports whether the resource restricts ports on any protocol.
// BYPASS_TWINGATE routes traffic outside Twingate, so port-level policies (anything
// other than ALLOW_ALL — i.e. RESTRICTED or DENY_ALL) can't be enforced and are rejected.
func hasPortRestriction(protocols *model.Protocols) bool {
	if protocols == nil {
		return false
	}

	return protocolRestricted(protocols.TCP) || protocolRestricted(protocols.UDP)
}

func protocolRestricted(protocol *model.Protocol) bool {
	return protocol != nil && protocol.Policy != "" && protocol.Policy != model.PolicyAllowAll
}
