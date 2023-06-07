package resources

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var ErrInvalidAttributeCombination = errors.New("invalid attribute combination")

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
	Access                   types.Object `tfsdk:"access"`
	IsVisible                types.Bool   `tfsdk:"is_visible"`
	IsBrowserShortcutEnabled types.Bool   `tfsdk:"is_browser_shortcut_enabled"`
	Alias                    types.String `tfsdk:"alias"`
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
}

func (r *twingateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	protocolSchema := schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			attr.Policy: schema.StringAttribute{
				Computed: true,
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(model.Policies...),
				},
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
			},
		},
	}

	resp.Schema = schema.Schema{
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
			attr.IsAuthoritative: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to `false`, assignments made outside of Terraform will be ignored.",
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknownBool(),
				},
			},

			attr.Alias: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Set a DNS alias address for the Resource. Must be a DNS-valid name string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// computed
			attr.IsVisible: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Controls whether this Resource will be visible in the main Resource list in the Twingate Client.",
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknownBool(),
				},
			},
			attr.IsBrowserShortcutEnabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: `Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.`,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknownBool(),
				},
			},

			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated ID of the Resource, encoded in base64",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},

		Blocks: map[string]schema.Block{
			attr.Access: schema.SingleNestedBlock{
				Description: "Restrict access to certain groups or service accounts",
				Attributes: map[string]schema.Attribute{
					attr.GroupIDs: schema.SetAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "List of Group IDs that will have permission to access the Resource.",
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							//setvalidator.AtLeastOneOf(path.Expressions{
							//	path.MatchRelative().AtParent().AtName(attr.ServiceAccountIDs),
							//}...),
						},
					},
					attr.ServiceAccountIDs: schema.SetAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "List of Service Account IDs that will have permission to access the Resource.",
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							//setvalidator.AtLeastOneOf(path.Expressions{
							//	path.MatchRelative().AtParent().AtName(attr.GroupIDs),
							//}...),
						},
					},
				},
			},

			attr.Protocols: schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					attr.AllowIcmp: schema.BoolAttribute{
						Computed:    true,
						Optional:    true,
						Description: "Whether to allow ICMP (ping) traffic",
					},
				},
				Blocks: map[string]schema.Block{
					attr.TCP: protocolSchema,
					attr.UDP: protocolSchema,
				},
				Description: "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction, and all protocols and ports are allowed.",
			},
		},
	}
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

	if err = r.client.AddResourceServiceAccountIDs(ctx, resource.ID, resource.ServiceAccounts); err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, TwingateResource)

		return
	}

	r.helper(ctx, resource, &plan, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func getAccessAttribute(obj *types.Object, attribute string) []string {
	if obj == nil || obj.IsUnknown() {
		return nil
	}

	val := obj.Attributes()[attribute]
	if val == nil || val.IsNull() || val.IsUnknown() {
		return nil
	}

	return convertIDs(val.(types.Set))
}

func convertResource(plan *resourceModel) (*model.Resource, error) {
	protocols, err := convertProtocols(&plan.Protocols)
	if err != nil {
		return nil, err
	}

	groupIDs := getAccessAttribute(&plan.Access, attr.GroupIDs)
	serviceAccountIDs := getAccessAttribute(&plan.Access, attr.ServiceAccountIDs)

	if !plan.Access.IsNull() && groupIDs == nil && serviceAccountIDs == nil {
		return nil, ErrInvalidAttributeCombination
	}

	var isVisible, isBrowserShortcutEnabled *bool
	if !plan.IsVisible.IsUnknown() {
		isVisible = plan.IsVisible.ValueBoolPointer()
	}

	if !plan.IsBrowserShortcutEnabled.IsUnknown() {
		isBrowserShortcutEnabled = plan.IsBrowserShortcutEnabled.ValueBoolPointer()
	}

	var alias *string
	if !plan.Alias.IsUnknown() && !plan.Alias.IsNull() {
		alias = plan.Alias.ValueStringPointer()
	}

	return &model.Resource{
		Name:                     plan.Name.ValueString(),
		RemoteNetworkID:          plan.RemoteNetworkID.ValueString(),
		Address:                  plan.Address.ValueString(),
		Protocols:                protocols,
		Groups:                   groupIDs,
		ServiceAccounts:          serviceAccountIDs,
		IsAuthoritative:          convertAuthoritativeFlag(plan.IsAuthoritative),
		Alias:                    alias,
		IsVisible:                isVisible,
		IsBrowserShortcutEnabled: isBrowserShortcutEnabled,
	}, nil
}

func convertIDs(list types.Set) []string {
	return utils.Map(list.Elements(), func(item tfattr.Value) string {
		return item.(types.String).ValueString()
	})
}

func convertProtocols(protocols *types.Object) (*model.Protocols, error) {
	if protocols == nil || protocols.IsNull() {
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
	if protocol == nil || protocol.IsNull() {
		return nil, nil //nolint:nilnil
	}

	obj, ok := protocol.(types.Object)
	if !ok || obj.IsNull() {
		return nil, nil
	}

	ports, err := decodePorts(obj)
	if err != nil {
		return nil, err
	}

	policy := obj.Attributes()[attr.Policy].(types.String).ValueString()

	return model.NewProtocol(policy, ports), nil
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

	if err = r.deleteResourceGroupIDs(ctx, &state, input); err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

		return
	}

	if err = r.deleteResourceServiceAccountIDs(ctx, &state, input); err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

		return
	}

	resource, err := r.client.UpdateResource(ctx, input)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

		return
	}

	addServiceAccountIDs := setDifference(input.ServiceAccounts, resource.ServiceAccounts)
	if err = r.client.AddResourceServiceAccountIDs(ctx, resource.ID, addServiceAccountIDs); err != nil {
		addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

		return
	}

	resource.ServiceAccounts = setJoin(resource.ServiceAccounts, input.ServiceAccounts)

	r.helper(ctx, resource, &state, &plan, &resp.State, &resp.Diagnostics, nil, operationUpdate)
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

func (r *twingateResource) helper(ctx context.Context, resource *model.Resource, state, reference *resourceModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) { //nolint:cyclop
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

	if !resource.IsActive {
		// fix set active state for the resource on `terraform apply`
		err = r.client.UpdateResourceActiveState(ctx, &model.Resource{
			ID:       resource.ID,
			IsActive: true,
		})

		if err != nil {
			addErr(diagnostics, err, operationUpdate, TwingateResource)

			return
		}
	}

	remoteServiceAccounts, err := r.client.ReadResourceServiceAccounts(ctx, resource.ID)
	if err != nil {
		addErr(diagnostics, err, operationRead, TwingateServiceAccount)

		return
	}

	resource.ServiceAccounts = remoteServiceAccounts

	if !resource.IsAuthoritative {
		resource.Groups = setIntersection(getAccessAttribute(&reference.Access, attr.GroupIDs), resource.Groups)
		resource.ServiceAccounts = setIntersection(getAccessAttribute(&reference.Access, attr.ServiceAccountIDs), resource.ServiceAccounts)
	}

	state.ID = types.StringValue(resource.ID)
	state.Name = types.StringValue(resource.Name)
	state.RemoteNetworkID = types.StringValue(resource.RemoteNetworkID)
	state.Address = types.StringValue(resource.Address)
	state.IsAuthoritative = types.BoolValue(resource.IsAuthoritative)

	if !state.IsVisible.IsNull() || !reference.IsVisible.IsUnknown() {
		state.IsVisible = types.BoolPointerValue(resource.IsVisible)
	}

	if !state.IsBrowserShortcutEnabled.IsNull() || !reference.IsBrowserShortcutEnabled.IsUnknown() {
		state.IsBrowserShortcutEnabled = types.BoolPointerValue(resource.IsBrowserShortcutEnabled)
	}

	if resource.Alias != nil {
		state.Alias = types.StringPointerValue(resource.Alias)
	}

	if state.Alias.IsUnknown() {
		state.Alias = types.StringNull()
	}

	if !state.Protocols.IsNull() {
		protocols, diags := convertProtocolsToTerraform(resource.Protocols, &reference.Protocols)
		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		state.Protocols = protocols
	}

	if !state.Access.IsNull() {
		access, diags := convertAccessBlockToTerraform(resource,
			state.Access.Attributes()[attr.GroupIDs],
			state.Access.Attributes()[attr.ServiceAccountIDs])

		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		state.Access = access
	}

	// Set refreshed state
	diagnostics.Append(respState.Set(ctx, state)...)
}

func convertAccessBlockToTerraform(resource *model.Resource, stateGroupIDs, stateServiceAccounts tfattr.Value) (types.Object, diag.Diagnostics) {
	var diagnostics, diags diag.Diagnostics

	groupIDs, serviceAccountIDs := types.SetNull(types.StringType), types.SetNull(types.StringType)

	if len(resource.Groups) > 0 {
		groupIDs, diags = makeSet(resource.Groups)
		diagnostics.Append(diags...)
	}

	if len(resource.ServiceAccounts) > 0 {
		serviceAccountIDs, diags = makeSet(resource.ServiceAccounts)
		diagnostics.Append(diags...)
	}

	attributeTypes := map[string]tfattr.Type{
		attr.GroupIDs: types.SetType{
			ElemType: types.StringType,
		},
		attr.ServiceAccountIDs: types.SetType{
			ElemType: types.StringType,
		},
	}

	if diagnostics.HasError() {
		return types.ObjectNull(attributeTypes), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.GroupIDs:          stateGroupIDs,
		attr.ServiceAccountIDs: stateServiceAccounts,
	}

	if !groupIDs.IsNull() {
		attributes[attr.GroupIDs] = groupIDs
	}

	if !serviceAccountIDs.IsNull() {
		attributes[attr.ServiceAccountIDs] = serviceAccountIDs
	}

	return types.ObjectValue(attributeTypes, attributes)
}

func makeSet(list []string) (basetypes.SetValue, diag.Diagnostics) {
	return types.SetValue(types.StringType, stringsToTerraformValue(list))
}

func stringsToTerraformValue(list []string) []tfattr.Value {
	if len(list) == 0 {
		return nil
	}

	out := make([]tfattr.Value, 0, len(list))
	for _, item := range list {
		out = append(out, types.StringValue(item))
	}

	return out
}

func (r *twingateResource) deleteResourceGroupIDs(ctx context.Context, state *resourceModel, input *model.Resource) error {
	groupIDs := r.getIDsToDelete(ctx, state, input.Groups, attr.GroupIDs, input)

	return r.client.DeleteResourceGroups(ctx, input.ID, groupIDs) //nolint
}

func (r *twingateResource) getIDsToDelete(ctx context.Context, state *resourceModel, currentIDs []string, attribute string, input *model.Resource) []string {
	oldIDs := r.getOldIDs(ctx, state, attribute, input)
	if len(oldIDs) == 0 {
		return nil
	}

	return setDifference(oldIDs, currentIDs)
}

func (r *twingateResource) getOldIDs(ctx context.Context, state *resourceModel, attribute string, input *model.Resource) []string {
	if input.IsAuthoritative {
		return r.getOldIDsAuthoritative(ctx, input, attribute)
	}

	return r.getOldIDsNonAuthoritative(state, attribute)
}

func (r *twingateResource) getOldIDsAuthoritative(ctx context.Context, input *model.Resource, attribute string) []string {
	switch attribute {
	case attr.ServiceAccountIDs:
		serviceAccounts, err := r.client.ReadResourceServiceAccounts(ctx, input.ID)
		if err != nil {
			return nil
		}

		return serviceAccounts

	case attr.GroupIDs:
		res, err := r.client.ReadResource(ctx, input.ID)
		if err != nil {
			return nil
		}

		return res.Groups
	}

	return nil
}

func (r *twingateResource) getOldIDsNonAuthoritative(state *resourceModel, attribute string) []string {
	switch attribute {
	case attr.GroupIDs, attr.ServiceAccountIDs:
		val := state.Access.Attributes()[attribute]
		if val == nil {
			return nil
		}

		values, ok := val.(types.Set)
		if !ok {
			return nil
		}

		return convertIDs(values)
	}

	return nil
}

func (r *twingateResource) deleteResourceServiceAccountIDs(ctx context.Context, state *resourceModel, input *model.Resource) error {
	idsToDelete := r.getIDsToDelete(ctx, state, input.ServiceAccounts, attr.ServiceAccountIDs, input)

	return r.client.DeleteResourceServiceAccounts(ctx, input.ID, idsToDelete) //nolint
}

func convertProtocolsToTerraform(protocols *model.Protocols, reference *types.Object) (types.Object, diag.Diagnostics) {
	if protocols == nil {
		return defaultProtocolsModelToTerraform()
	}

	var diagnostics diag.Diagnostics

	tcp, diags := convertProtocolModelToTerraform(protocols.TCP, reference.Attributes()[attr.TCP])
	diagnostics.Append(diags...)

	udp, diags := convertProtocolModelToTerraform(protocols.UDP, reference.Attributes()[attr.UDP])
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return types.ObjectNull(protocolsAttributeTypes()), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.AllowIcmp: types.BoolValue(protocols.AllowIcmp),
		attr.TCP:       tcp,
		attr.UDP:       udp,
	}

	return types.ObjectValue(protocolsAttributeTypes(), attributes)
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

func defaultProtocolsModelToTerraform() (types.Object, diag.Diagnostics) {
	attributeTypes := protocolsAttributeTypes()

	var diagnostics diag.Diagnostics

	defaultPorts, diags := defaultProtocolModelToTerraform()
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return types.ObjectNull(attributeTypes), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.AllowIcmp: types.BoolValue(true),
		attr.TCP:       defaultPorts,
		attr.UDP:       defaultPorts,
	}

	return types.ObjectValue(attributeTypes, attributes)
}

func defaultProtocolModelToTerraform() (basetypes.ObjectValue, diag.Diagnostics) {
	attributes := map[string]tfattr.Value{
		attr.Policy: types.StringValue(model.PolicyAllowAll),
		attr.Ports:  types.ListNull(types.StringType),
	}

	return types.ObjectValue(protocolAttributeTypes(), attributes)
}

func convertProtocolModelToTerraform(protocol *model.Protocol, reference tfattr.Value) (types.Object, diag.Diagnostics) {
	if protocol == nil {
		return types.ObjectNull(protocolAttributeTypes()), nil
	}

	policy := protocol.Policy

	statePolicy := reference.(types.Object).Attributes()[attr.Policy].(types.String).ValueString()
	if statePolicy == model.PolicyDenyAll && policy == model.PolicyRestricted {
		policy = model.PolicyDenyAll
	}

	var statePorts = types.Set{}

	statePortsVal := reference.(types.Object).Attributes()[attr.Ports]

	if statePortsVal != nil && !statePortsVal.IsUnknown() {
		statePortsSet, ok := statePortsVal.(types.Set)
		if ok {
			statePorts = statePortsSet
		}
	}

	ports := convertPortsToTerraform(protocol.Ports)
	if equalPorts(ports, statePorts) && !statePorts.IsNull() {
		ports = statePorts
	}

	attributes := map[string]tfattr.Value{
		attr.Policy: types.StringValue(policy),
		attr.Ports:  ports,
	}

	return types.ObjectValue(protocolAttributeTypes(), attributes)
}

func convertPortsToTerraform(ports []*model.PortRange) types.Set {
	elements := make([]tfattr.Value, 0, len(ports))
	for _, port := range ports {
		elements = append(elements, types.StringValue(port.String()))
	}

	return types.SetValueMust(types.StringType, elements)
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
	// Do nothing if there is no state value.
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

// UseStateForUnknownBool returns a plan modifier that copies a known prior state
// value into the planned value. Use this when it is known that an unconfigured
// value will remain the same after a resource update.
//
// To prevent Terraform errors, the framework automatically sets unconfigured
// and Computed attributes to an unknown value "(known after apply)" on update.
// Using this plan modifier will instead display the prior state value in the
// plan, unless a prior plan modifier adjusts the value.
func UseStateForUnknownBool() planmodifier.Bool {
	return useStateForUnknownBoolModifier{}
}

// useStateForUnknownModifier implements the plan modifier.
type useStateForUnknownBoolModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m useStateForUnknownBoolModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m useStateForUnknownBoolModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModifyBool implements the plan modification logic.
func (m useStateForUnknownBoolModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}
