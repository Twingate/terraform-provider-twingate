package resource

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

var (
	ErrPortsWithPolicyAllowAll      = errors.New(model.PolicyAllowAll + " policy does not allow specifying ports.")
	ErrPortsWithPolicyDenyAll       = errors.New(model.PolicyDenyAll + " policy does not allow specifying ports.")
	ErrPolicyRestrictedWithoutPorts = errors.New(model.PolicyRestricted + " policy requires specifying ports.")
	ErrInvalidAttributeCombination  = errors.New("invalid attribute combination")
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
	Protocols                types.List   `tfsdk:"protocols"`
	Access                   types.List   `tfsdk:"access"`
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

	res, err := r.client.ReadResource(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("failed to import state", err.Error())

		return
	}

	resp.State.SetAttribute()

	if res.Protocols != nil {
		protocols, diags := convertProtocolsToTerraform(ctx, res.Protocols, reference.Protocols)
		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		state.Protocols = protocols
	}

	if !state.Access.IsNull() {
		access, diags := convertAccessBlockToTerraform(ctx, resource,
			state.Access.Elements()[0].(types.Object).Attributes()[attr.GroupIDs],
			state.Access.Elements()[0].(types.Object).Attributes()[attr.ServiceAccountIDs])

		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		state.Access = access
	}

}

func (r *twingateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	protocolSchema := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attr.Policy: schema.StringAttribute{
					Required: true,
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
				Description: "Set a DNS alias address for the Resource. Must be a DNS-valid name string.",
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
			attr.Access: schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "Restrict access to certain groups or service accounts",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						attr.GroupIDs: schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "List of Group IDs that will have permission to access the Resource.",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
						attr.ServiceAccountIDs: schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "List of Service Account IDs that will have permission to access the Resource.",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},

			attr.Protocols: schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
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

	if err = r.client.AddResourceServiceAccountIDs(ctx, resource); err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, TwingateResource)

		return
	}

	r.helper(ctx, resource, &plan, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
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

func convertResource(plan *resourceModel) (*model.Resource, error) {
	protocols, err := convertProtocols(&plan.Protocols)
	if err != nil {
		return nil, err
	}

	groupIDs := getAccessAttribute(plan.Access, attr.GroupIDs)
	serviceAccountIDs := getAccessAttribute(plan.Access, attr.ServiceAccountIDs)

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

func convertProtocols(protocols *types.List) (*model.Protocols, error) {
	if protocols == nil || protocols.IsNull() || len(protocols.Elements()) == 0 {
		return model.DefaultProtocols(), nil
	}

	udp, err := convertProtocol(protocols.Elements()[0].(types.Object).Attributes()[attr.UDP])
	if err != nil {
		return nil, err
	}

	tcp, err := convertProtocol(protocols.Elements()[0].(types.Object).Attributes()[attr.TCP])
	if err != nil {
		return nil, err
	}

	return &model.Protocols{
		AllowIcmp: protocols.Elements()[0].(types.Object).Attributes()[attr.AllowIcmp].(types.Bool).ValueBool(),
		UDP:       udp,
		TCP:       tcp,
	}, nil
}

func convertProtocol(listVal tfattr.Value) (*model.Protocol, error) {
	if listVal == nil || listVal.IsNull() {
		return nil, nil //nolint:nilnil
	}

	list := listVal.(types.List)

	if list.IsNull() || len(list.Elements()) == 0 {
		return nil, nil
	}

	protocol := list.Elements()[0]

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

	switch policy {
	case model.PolicyAllowAll:
		if len(ports) > 0 {
			return nil, ErrPortsWithPolicyAllowAll
		}

	case model.PolicyDenyAll:
		if len(ports) > 0 {
			return nil, ErrPortsWithPolicyDenyAll
		}

	case model.PolicyRestricted:
		if len(ports) == 0 {
			return nil, ErrPolicyRestrictedWithoutPorts
		}
	}

	if policy == model.PolicyDenyAll {
		policy = model.PolicyRestricted
	}

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

	if !plan.Access.Equal(state.Access) {
		idsToDelete, idsToAdd, err := r.getChangedAccessIDs(ctx, &plan, &state, input)
		if err != nil {
			addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

			return
		}

		if err := r.client.RemoveResourceAccess(ctx, input.ID, idsToDelete); err != nil {
			addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

			return
		}

		if err = r.client.AddResourceAccess(ctx, input.ID, idsToAdd); err != nil {
			addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

			return
		}
	}

	var resource *model.Resource
	if !plan.RemoteNetworkID.Equal(state.RemoteNetworkID) ||
		!plan.Name.Equal(state.Name) ||
		!plan.Address.Equal(state.Address) ||
		!plan.Protocols.Equal(state.Protocols) ||
		!plan.IsVisible.Equal(state.IsVisible) ||
		!plan.IsBrowserShortcutEnabled.Equal(state.IsBrowserShortcutEnabled) ||
		!plan.Alias.Equal(state.Alias) {
		resource, err = r.client.UpdateResource(ctx, input)
	} else {
		resource, err = r.client.ReadResource(ctx, input.ID)
	}

	if resource != nil {
		resource.IsAuthoritative = input.IsAuthoritative
	}

	//if err = r.deleteResourceGroupIDs(ctx, &state, input); err != nil {
	//	addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)
	//
	//	return
	//}
	//
	//if err = r.deleteResourceServiceAccountIDs(ctx, &state, input); err != nil {
	//	addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)
	//
	//	return
	//}
	//
	//resource, err := r.client.UpdateResource(ctx, input)
	//if err != nil {
	//	addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)
	//
	//	return
	//}
	//
	//addServiceAccountIDs := setDifference(input.ServiceAccounts, resource.ServiceAccounts)
	//if err = r.client.AddResourceServiceAccountIDs(ctx, resource.ID, addServiceAccountIDs); err != nil {
	//	addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)
	//
	//	return
	//}
	//
	//resource.ServiceAccounts = setJoin(resource.ServiceAccounts, input.ServiceAccounts)

	r.helper(ctx, resource, &state, &plan, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func (r *twingateResource) getChangedAccessIDs(ctx context.Context, plan, state *resourceModel, resource *model.Resource) ([]string, []string, error) {
	remote, err := r.client.ReadResource(ctx, resource.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get changedIDs: %w", err)
	}

	var oldGroups, oldServiceAccounts []string
	if resource.IsAuthoritative {
		oldGroups, oldServiceAccounts = remote.Groups, remote.ServiceAccounts
	} else {
		oldGroups = getOldIDsNonAuthoritative(plan, state, attr.GroupIDs)
		oldServiceAccounts = getOldIDsNonAuthoritative(plan, state, attr.ServiceAccountIDs)
	}

	// ids to delete
	groupsToDelete := setDifference(oldGroups, resource.Groups)
	serviceAccountsToDelete := setDifference(oldServiceAccounts, resource.ServiceAccounts)

	// ids to add
	groupsToAdd := setDifference(resource.Groups, remote.Groups)
	serviceAccountsToAdd := setDifference(resource.ServiceAccounts, remote.ServiceAccounts)

	return append(groupsToDelete, serviceAccountsToDelete...), append(groupsToAdd, serviceAccountsToAdd...), nil
}

func getOldIDsNonAuthoritative(plan, state *resourceModel, attribute string) []string {
	if !plan.Access.Equal(state.Access) {
		return getAccessAttribute(state.Access, attribute)
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
		resource.Groups = setIntersection(getAccessAttribute(reference.Access, attr.GroupIDs), resource.Groups)
		resource.ServiceAccounts = setIntersection(getAccessAttribute(reference.Access, attr.ServiceAccountIDs), resource.ServiceAccounts)
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

	if !state.Alias.IsNull() || !reference.Alias.IsUnknown() {
		state.Alias = reference.Alias
	}

	if !state.Protocols.IsNull() {
		protocols, diags := convertProtocolsToTerraform(ctx, resource.Protocols, reference.Protocols)
		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		state.Protocols = protocols
	}

	if !state.Access.IsNull() {
		access, diags := convertAccessBlockToTerraform(ctx, resource,
			state.Access.Elements()[0].(types.Object).Attributes()[attr.GroupIDs],
			state.Access.Elements()[0].(types.Object).Attributes()[attr.ServiceAccountIDs])

		diagnostics.Append(diags...)

		if diagnostics.HasError() {
			return
		}

		state.Access = access
	}

	// Set refreshed state
	diagnostics.Append(respState.Set(ctx, state)...)
}

func convertAccessBlockToTerraform(ctx context.Context, resource *model.Resource, stateGroupIDs, stateServiceAccounts tfattr.Value) (types.List, diag.Diagnostics) {
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

	if diagnostics.HasError() {
		return makeObjectsListNull(ctx, accessAttributeTypes()), diagnostics
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

	obj, diags := types.ObjectValue(accessAttributeTypes(), attributes)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeObjectsListNull(ctx, accessAttributeTypes()), diagnostics
	}

	return makeObjectsList(ctx, obj)
}

func accessAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.GroupIDs: types.SetType{
			ElemType: types.StringType,
		},
		attr.ServiceAccountIDs: types.SetType{
			ElemType: types.StringType,
		},
	}
}

func makeObjectsListNull(ctx context.Context, attributeTypes map[string]tfattr.Type) types.List {
	return types.ListNull(types.ObjectNull(attributeTypes).Type(ctx))
}

func makeObjectsList(ctx context.Context, objects ...types.Object) (types.List, diag.Diagnostics) {
	obj := objects[0]

	items := utils.Map(objects, func(item types.Object) tfattr.Value {
		return tfattr.Value(item)
	})

	return types.ListValue(obj.Type(ctx), items)
}

func makeSet(list []string) (types.Set, diag.Diagnostics) {
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

//func (r *twingateResource) deleteResourceGroupIDs(ctx context.Context, state *resourceModel, input *model.Resource) error {
//	groupIDs := r.getIDsToDelete(ctx, state, input.Groups, attr.GroupIDs, input)
//
//	return r.client.DeleteResourceGroups(ctx, input.ID, groupIDs) //nolint
//}

//func (r *twingateResource) getIDsToDelete(ctx context.Context, state *resourceModel, currentIDs []string, attribute string, input *model.Resource) []string {
//	oldIDs := r.getOldIDs(ctx, state, attribute, input)
//	if len(oldIDs) == 0 {
//		return nil
//	}
//
//	return setDifference(oldIDs, currentIDs)
//}

//func (r *twingateResource) getOldIDs(ctx context.Context, state *resourceModel, attribute string, input *model.Resource) []string {
//	if input.IsAuthoritative {
//		return r.getOldIDsAuthoritative(ctx, input, attribute)
//	}
//
//	return r.getOldIDsNonAuthoritative(state, attribute)
//}

//func (r *twingateResource) getOldIDsAuthoritative(ctx context.Context, input *model.Resource, attribute string) []string {
//	switch attribute {
//	case attr.ServiceAccountIDs:
//		serviceAccounts, err := r.client.ReadResourceServiceAccounts(ctx, input.ID)
//		if err != nil {
//			return nil
//		}
//
//		return serviceAccounts
//
//	case attr.GroupIDs:
//		res, err := r.client.ReadResource(ctx, input.ID)
//		if err != nil {
//			return nil
//		}
//
//		return res.Groups
//	}
//
//	return nil
//}

//func (r *twingateResource) getOldIDsNonAuthoritative(state *resourceModel, attribute string) []string {
//	switch attribute {
//	case attr.GroupIDs, attr.ServiceAccountIDs:
//		val := state.Access.Elements()[0].(types.Object).Attributes()[attribute]
//		if val == nil {
//			return nil
//		}
//
//		values, ok := val.(types.Set)
//		if !ok {
//			return nil
//		}
//
//		return convertIDs(values)
//	}
//
//	return nil
//}

//func (r *twingateResource) deleteResourceServiceAccountIDs(ctx context.Context, state *resourceModel, input *model.Resource) error {
//	idsToDelete := r.getIDsToDelete(ctx, state, input.ServiceAccounts, attr.ServiceAccountIDs, input)
//
//	return r.client.DeleteResourceServiceAccounts(ctx, input.ID, idsToDelete) //nolint
//}

func convertProtocolsToTerraform(ctx context.Context, protocols *model.Protocols, reference types.List) (types.List, diag.Diagnostics) {
	if protocols == nil {
		return defaultProtocolsModelToTerraform(ctx)
	}

	var diagnostics diag.Diagnostics

	tcp, diags := convertProtocolModelToTerraform(protocols.TCP, reference.Elements()[0].(types.Object).Attributes()[attr.TCP])
	diagnostics.Append(diags...)

	udp, diags := convertProtocolModelToTerraform(protocols.UDP, reference.Elements()[0].(types.Object).Attributes()[attr.UDP])
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return types.ListNull(types.ObjectNull(protocolsAttributeTypes()).Type(ctx)), diagnostics
	}

	attributes := map[string]tfattr.Value{
		attr.AllowIcmp: types.BoolValue(protocols.AllowIcmp),
		attr.TCP:       types.ListValueMust(tcp.Type(ctx), []tfattr.Value{tcp}),
		attr.UDP:       types.ListValueMust(udp.Type(ctx), []tfattr.Value{udp}),
	}

	obj := types.ObjectValueMust(protocolsAttributeTypes(), attributes)
	list := []tfattr.Value{obj}

	return types.ListValue(obj.Type(ctx), list)
}

func protocolsAttributeTypes() map[string]tfattr.Type {
	return map[string]tfattr.Type{
		attr.AllowIcmp: types.BoolType,
		attr.TCP: types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: protocolAttributeTypes(),
			},
		},
		attr.UDP: types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: protocolAttributeTypes(),
			},
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

func defaultProtocolsModelToTerraform(ctx context.Context) (types.List, diag.Diagnostics) {
	attributeTypes := protocolsAttributeTypes()

	var diagnostics diag.Diagnostics

	defaultPorts, diags := defaultProtocolModelToTerraform()
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeObjectsListNull(ctx, attributeTypes), diagnostics
	}

	tcp, diags := makeObjectsList(ctx, defaultPorts)
	diagnostics.Append(diags...)

	udp, diags := makeObjectsList(ctx, defaultPorts)
	diagnostics.Append(diags...)

	attributes := map[string]tfattr.Value{
		attr.AllowIcmp: types.BoolValue(true),
		attr.TCP:       tcp,
		attr.UDP:       udp,
	}

	obj, diags := types.ObjectValue(attributeTypes, attributes)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return makeObjectsListNull(ctx, attributeTypes), diagnostics
	}

	return makeObjectsList(ctx, obj)
}

func defaultProtocolModelToTerraform() (basetypes.ObjectValue, diag.Diagnostics) {
	attributes := map[string]tfattr.Value{
		attr.Policy: types.StringValue(model.PolicyAllowAll),
		attr.Ports:  types.ListNull(types.StringType),
	}

	return types.ObjectValue(protocolAttributeTypes(), attributes)
}

func convertProtocolOnlyModelToTerraform(protocol *model.Protocol) (types.Object, diag.Diagnostics) {
	if protocol == nil {
		return types.ObjectNull(protocolAttributeTypes()), nil
	}

	var statePorts = types.Set{}

	ports := convertPortsToTerraform(protocol.Ports)
	if equalPorts(ports, statePorts) && !statePorts.IsNull() {
		ports = statePorts
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

func convertProtocolModelToTerraform(protocol *model.Protocol, reference tfattr.Value) (types.Object, diag.Diagnostics) {
	if protocol == nil {
		return types.ObjectNull(protocolAttributeTypes()), nil
	}

	var statePorts = types.Set{}

	statePortsVal := reference.(types.List).Elements()[0].(types.Object).Attributes()[attr.Ports]

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

func convertProtocolModelToTerraform(protocol *model.Protocol, reference tfattr.Value) (types.Object, diag.Diagnostics) {
	if protocol == nil {
		return types.ObjectNull(protocolAttributeTypes()), nil
	}

	var statePorts = types.Set{}

	statePortsVal := reference.(types.List).Elements()[0].(types.Object).Attributes()[attr.Ports]

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

//func Resource() *schema.Resource { //nolint:funlen
//	portsSchema := &schema.Resource{
//		Schema: map[string]*schema.Schema{
//			attr.Policy: {
//				Type:         schema.TypeString,
//				Required:     true,
//				ValidateFunc: validation.StringInSlice(model.Policies, false),
//				Description:  fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", model.PolicyRestricted, model.PolicyAllowAll, model.PolicyDenyAll),
//			},
//			attr.Ports: {
//				Type:        schema.TypeList,
//				Optional:    true,
//				Description: "List of port ranges between 1 and 65535 inclusive, in the format `100-200` for a range, or `8080` for a single port",
//				Elem: &schema.Schema{
//					Type: schema.TypeString,
//				},
//				DiffSuppressFunc: portsNotChanged,
//			},
//		},
//	}
//
//	protocolsSchema := &schema.Resource{
//		Schema: map[string]*schema.Schema{
//			attr.AllowIcmp: {
//				Type:        schema.TypeBool,
//				Optional:    true,
//				Default:     true,
//				Description: "Whether to allow ICMP (ping) traffic",
//			},
//			attr.TCP: {
//				Type:     schema.TypeList,
//				Required: true,
//				MaxItems: 1,
//				Elem:     portsSchema,
//			},
//			attr.UDP: {
//				Type:     schema.TypeList,
//				Required: true,
//				MaxItems: 1,
//				Elem:     portsSchema,
//			},
//		},
//	}
//
//	accessSchema := &schema.Resource{
//		Schema: map[string]*schema.Schema{
//			attr.GroupIDs: {
//				Type:         schema.TypeSet,
//				Elem:         &schema.Schema{Type: schema.TypeString},
//				MinItems:     1,
//				Optional:     true,
//				AtLeastOneOf: []string{attr.Path(attr.Access, attr.ServiceAccountIDs)},
//				Description:  "List of Group IDs that will have permission to access the Resource.",
//			},
//			attr.ServiceAccountIDs: {
//				Type:         schema.TypeSet,
//				Elem:         &schema.Schema{Type: schema.TypeString},
//				MinItems:     1,
//				Optional:     true,
//				AtLeastOneOf: []string{attr.Path(attr.Access, attr.GroupIDs)},
//				Description:  "List of Service Account IDs that will have permission to access the Resource.",
//			},
//		},
//	}
//
//	return &schema.Resource{
//		Description:   "Resources in Twingate represent servers on the private network that clients can connect to. Resources can be defined by IP, CIDR range, FQDN, or DNS zone. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
//		CreateContext: resourceCreate,
//		UpdateContext: resourceUpdate,
//		ReadContext:   resourceRead,
//		DeleteContext: resourceDelete,
//
//		Schema: map[string]*schema.Schema{
//			// required
//			attr.Name: {
//				Type:        schema.TypeString,
//				Required:    true,
//				Description: "The name of the Resource",
//			},
//			attr.Address: {
//				Type:        schema.TypeString,
//				Required:    true,
//				Description: "The Resource's IP/CIDR or FQDN/DNS zone",
//			},
//			attr.RemoteNetworkID: {
//				Type:        schema.TypeString,
//				Required:    true,
//				Description: "Remote Network ID where the Resource lives",
//			},
//			// optional
//			attr.IsAuthoritative: {
//				Type:        schema.TypeBool,
//				Optional:    true,
//				Computed:    true,
//				Description: "Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to `false`, assignments made outside of Terraform will be ignored.",
//			},
//			attr.Protocols: {
//				Type:                  schema.TypeList,
//				Optional:              true,
//				MaxItems:              1,
//				Description:           "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction, and all protocols and ports are allowed.",
//				Elem:                  protocolsSchema,
//				DiffSuppressOnRefresh: true,
//				DiffSuppressFunc:      protocolsNotChanged,
//			},
//			attr.Access: {
//				Type:        schema.TypeList,
//				Optional:    true,
//				MaxItems:    1,
//				Description: "Restrict access to certain groups or service accounts",
//				Elem:        accessSchema,
//			},
//			// computed
//			attr.IsVisible: {
//				Type:        schema.TypeBool,
//				Optional:    true,
//				Computed:    true,
//				Description: "Controls whether this Resource will be visible in the main Resource list in the Twingate Client.",
//			},
//			attr.IsBrowserShortcutEnabled: {
//				Type:        schema.TypeBool,
//				Optional:    true,
//				Computed:    true,
//				Description: `Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.`,
//			},
//			attr.Alias: {
//				Type:             schema.TypeString,
//				Optional:         true,
//				Description:      "Set a DNS alias address for the Resource. Must be a DNS-valid name string.",
//				DiffSuppressFunc: aliasDiff,
//			},
//			attr.ID: {
//				Type:        schema.TypeString,
//				Computed:    true,
//				Description: "Autogenerated ID of the Resource, encoded in base64",
//			},
//		},
//		Importer: &schema.ResourceImporter{
//			StateContext: schema.ImportStatePassthroughContext,
//		},
//	}
//}
//
//func resourceCreate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	client := meta.(*client.Client)
//
//	resource, err := convertResource(resourceData)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	resource, err = client.CreateResource(ctx, resource)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	if err = client.AddResourceServiceAccountIDs(ctx, resource); err != nil {
//		return diag.FromErr(err)
//	}
//
//	log.Printf("[INFO] Created resource %s", resource.Name)
//
//	return resourceResourceReadHelper(ctx, client, resourceData, resource, nil)
//}
//
//func resourceUpdate(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	client := meta.(*client.Client)
//
//	resource, err := convertResource(resourceData)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	resource.ID = resourceData.Id()
//
//	if resourceData.HasChange(attr.Access) {
//		idsToDelete, idsToAdd, err := getChangedAccessIDs(ctx, resourceData, resource, client)
//		if err != nil {
//			return diag.FromErr(err)
//		}
//
//		if err := client.RemoveResourceAccess(ctx, resource.ID, idsToDelete); err != nil {
//			return diag.FromErr(err)
//		}
//
//		if err = client.AddResourceAccess(ctx, resource.ID, idsToAdd); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//
//	if resourceData.HasChanges(
//		attr.RemoteNetworkID,
//		attr.Name,
//		attr.Address,
//		attr.Protocols,
//		attr.IsVisible,
//		attr.IsBrowserShortcutEnabled,
//		attr.Alias,
//	) {
//		resource, err = client.UpdateResource(ctx, resource)
//	} else {
//		resource, err = client.ReadResource(ctx, resource.ID)
//	}
//
//	if resource != nil {
//		resource.IsAuthoritative = convertAuthoritativeFlag(resourceData)
//	}
//
//	log.Printf("[INFO] Updated resource %s", resource.Name)
//
//	return resourceResourceReadHelper(ctx, client, resourceData, resource, err)
//}
//
//func resourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	client := meta.(*client.Client)
//
//	resource, err := client.ReadResource(ctx, resourceData.Id())
//	if resource != nil {
//		resource.IsAuthoritative = convertAuthoritativeFlag(resourceData)
//	}
//
//	return resourceResourceReadHelper(ctx, client, resourceData, resource, err)
//}
//
//func resourceDelete(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	c := meta.(*client.Client)
//	resourceID := resourceData.Id()
//
//	err := c.DeleteResource(ctx, resourceID)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	log.Printf("[INFO] Deleted resource id %s", resourceData.Id())
//
//	return nil
//}
//
//func resourceResourceReadHelper(ctx context.Context, resourceClient *client.Client, resourceData *schema.ResourceData, resource *model.Resource, err error) diag.Diagnostics {
//	if err != nil {
//		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
//			// clear state
//			resourceData.SetId("")
//
//			return nil
//		}
//
//		return diag.FromErr(err)
//	}
//
//	if resource.Protocols == nil {
//		resource.Protocols = model.DefaultProtocols()
//	}
//
//	if !resource.IsActive {
//		// fix set active state for the resource on `terraform apply`
//		err = resourceClient.UpdateResourceActiveState(ctx, &model.Resource{
//			ID:       resource.ID,
//			IsActive: true,
//		})
//
//		if err != nil {
//			return diag.FromErr(err)
//		}
//	}
//
//	if !resource.IsAuthoritative {
//		groups, serviceAccounts := convertAccess(resourceData)
//		resource.ServiceAccounts = setIntersection(serviceAccounts, resource.ServiceAccounts)
//		resource.Groups = setIntersection(groups, resource.Groups)
//	}
//
//	resourceData.SetId(resource.ID)
//
//	return readDiagnostics(resourceData, resource)
//}
//
//func readDiagnostics(resourceData *schema.ResourceData, resource *model.Resource) diag.Diagnostics { //nolint:cyclop
//	if err := resourceData.Set(attr.Name, resource.Name); err != nil {
//		return ErrAttributeSet(err, attr.Name)
//	}
//
//	if err := resourceData.Set(attr.RemoteNetworkID, resource.RemoteNetworkID); err != nil {
//		return ErrAttributeSet(err, attr.RemoteNetworkID)
//	}
//
//	if err := resourceData.Set(attr.Address, resource.Address); err != nil {
//		return ErrAttributeSet(err, attr.Address)
//	}
//
//	if err := resourceData.Set(attr.IsAuthoritative, resource.IsAuthoritative); err != nil {
//		return ErrAttributeSet(err, attr.IsAuthoritative)
//	}
//
//	if err := resourceData.Set(attr.Access, resource.AccessToTerraform()); err != nil {
//		return ErrAttributeSet(err, attr.Access)
//	}
//
//	if err := resourceData.Set(attr.Protocols, resource.Protocols.ToTerraform()); err != nil {
//		return ErrAttributeSet(err, attr.Protocols)
//	}
//
//	if resource.IsVisible != nil {
//		if err := resourceData.Set(attr.IsVisible, *resource.IsVisible); err != nil {
//			return ErrAttributeSet(err, attr.IsVisible)
//		}
//	}
//
//	if resource.IsBrowserShortcutEnabled != nil {
//		if err := resourceData.Set(attr.IsBrowserShortcutEnabled, *resource.IsBrowserShortcutEnabled); err != nil {
//			return ErrAttributeSet(err, attr.IsBrowserShortcutEnabled)
//		}
//	}
//
//	var alias interface{}
//	if resource.Alias != nil {
//		alias = *resource.Alias
//	}
//
//	if err := resourceData.Set(attr.Alias, alias); err != nil {
//		return ErrAttributeSet(err, attr.Alias)
//	}
//
//	return nil
//}
//
//func aliasDiff(key, _, _ string, resourceData *schema.ResourceData) bool {
//	oldVal, newVal := castToStrings(resourceData.GetChange(key))
//
//	return oldVal == newVal
//}
//
//func equalPorts(a, b interface{}) bool {
//	oldPorts, newPorts := a.([]interface{}), b.([]interface{})
//
//	oldPortsRange, err := convertPorts(oldPorts)
//	if err != nil {
//		return false
//	}
//
//	newPortsRange, err := convertPorts(newPorts)
//	if err != nil {
//		return false
//	}
//
//	oldPortsMap := convertPortsRangeToMap(oldPortsRange)
//	newPortsMap := convertPortsRangeToMap(newPortsRange)
//
//	return reflect.DeepEqual(oldPortsMap, newPortsMap)
//}
//
//func convertPortsRangeToMap(portsRange []*model.PortRange) map[int]struct{} {
//	out := make(map[int]struct{})
//
//	for _, port := range portsRange {
//		if port.Start == port.End {
//			out[port.Start] = struct{}{}
//
//			continue
//		}
//
//		for i := port.Start; i <= port.End; i++ {
//			out[i] = struct{}{}
//		}
//	}
//
//	return out
//}
//
//func portsNotChanged(attribute, oldValue, newValue string, data *schema.ResourceData) bool {
//	keys := []string{
//		attr.Path(attr.Protocols, attr.TCP, attr.Ports),
//		attr.Path(attr.Protocols, attr.UDP, attr.Ports),
//	}
//
//	if strings.HasSuffix(attribute, "#") && newValue == "0" {
//		return newValue == oldValue
//	}
//
//	for _, key := range keys {
//		if strings.HasPrefix(attribute, key) {
//			return equalPorts(data.GetChange(key))
//		}
//	}
//
//	return false
//}
//
//// protocolsNotChanged - suppress protocols change when uses default value.
//func protocolsNotChanged(attribute, oldValue, newValue string, data *schema.ResourceData) bool {
//	switch attribute {
//	case attr.Len(attr.Protocols):
//		return newValue == "0"
//	case attr.Len(attr.Protocols, attr.TCP), attr.Len(attr.Protocols, attr.UDP):
//		return newValue == "0"
//	case attr.Path(attr.Protocols, attr.TCP, attr.Policy), attr.Path(attr.Protocols, attr.UDP, attr.Policy):
//		return oldValue == model.PolicyAllowAll && newValue == ""
//	}
//
//	return false
//}
//
//func getChangedAccessIDs(ctx context.Context, resourceData *schema.ResourceData, resource *model.Resource, client *client.Client) ([]string, []string, error) {
//	remote, err := client.ReadResource(ctx, resource.ID)
//	if err != nil {
//		return nil, nil, fmt.Errorf("failed to get changedIDs: %w", err)
//	}
//
//	var oldGroups, oldServiceAccounts []string
//	if resource.IsAuthoritative {
//		oldGroups, oldServiceAccounts = remote.Groups, remote.ServiceAccounts
//	} else {
//		oldGroups = getOldIDsNonAuthoritative(resourceData, attr.GroupIDs)
//		oldServiceAccounts = getOldIDsNonAuthoritative(resourceData, attr.ServiceAccountIDs)
//	}
//
//	// ids to delete
//	groupsToDelete := setDifference(oldGroups, resource.Groups)
//	serviceAccountsToDelete := setDifference(oldServiceAccounts, resource.ServiceAccounts)
//
//	// ids to add
//	groupsToAdd := setDifference(resource.Groups, remote.Groups)
//	serviceAccountsToAdd := setDifference(resource.ServiceAccounts, remote.ServiceAccounts)
//
//	return append(groupsToDelete, serviceAccountsToDelete...), append(groupsToAdd, serviceAccountsToAdd...), nil
//}
//
//func getOldIDsNonAuthoritative(resourceData *schema.ResourceData, attribute string) []string {
//	if resourceData.HasChange(attr.Path(attr.Access, attribute)) {
//		old, _ := resourceData.GetChange(attr.Path(attr.Access, attribute))
//
//		return convertIDs(old)
//	}
//
//	return nil
//}
//
//func convertResource(data *schema.ResourceData) (*model.Resource, error) {
//	protocols, err := convertProtocols(data)
//	if err != nil {
//		return nil, err
//	}
//
//	groups, serviceAccounts := convertAccess(data)
//	res := &model.Resource{
//		Name:            data.Get(attr.Name).(string),
//		RemoteNetworkID: data.Get(attr.RemoteNetworkID).(string),
//		Address:         data.Get(attr.Address).(string),
//		Protocols:       protocols,
//		Groups:          groups,
//		ServiceAccounts: serviceAccounts,
//		IsAuthoritative: convertAuthoritativeFlag(data),
//		Alias:           getOptionalString(data, attr.Alias),
//	}
//
//	isVisible, ok := data.GetOkExists(attr.IsVisible) //nolint
//	if val := isVisible.(bool); ok {
//		res.IsVisible = &val
//	}
//
//	isBrowserShortcutEnabled, ok := data.GetOkExists(attr.IsBrowserShortcutEnabled) //nolint:staticcheck
//	if val := isBrowserShortcutEnabled.(bool); ok {
//		res.IsBrowserShortcutEnabled = &val
//	}
//
//	return res, nil
//}
//
//func getOptionalString(data *schema.ResourceData, attr string) *string {
//	var result *string
//
//	cfg := data.GetRawConfig()
//	val := cfg.GetAttr(attr)
//
//	if !val.IsNull() {
//		str := val.AsString()
//		result = &str
//	}
//
//	return result
//}
//
//func convertAccess(data *schema.ResourceData) ([]string, []string) {
//	rawList := data.Get(attr.Access).([]interface{})
//	if len(rawList) == 0 || rawList[0] == nil {
//		return nil, nil
//	}
//
//	rawMap := rawList[0].(map[string]interface{})
//
//	return convertIDs(rawMap[attr.GroupIDs]), convertIDs(rawMap[attr.ServiceAccountIDs])
//}
//
//func convertAuthoritativeFlag(data *schema.ResourceData) bool {
//	flag, hasFlag := data.GetOkExists(attr.IsAuthoritative) //nolint:staticcheck
//
//	if hasFlag {
//		return flag.(bool)
//	}
//
//	// default value
//	return true
//}
//
//func convertProtocols(data *schema.ResourceData) (*model.Protocols, error) {
//	rawList := data.Get(attr.Protocols).([]interface{})
//	if len(rawList) == 0 {
//		return model.DefaultProtocols(), nil
//	}
//
//	rawMap := rawList[0].(map[string]interface{})
//
//	udp, err := convertProtocol(rawMap[attr.UDP].([]interface{}))
//	if err != nil {
//		return nil, err
//	}
//
//	tcp, err := convertProtocol(rawMap[attr.TCP].([]interface{}))
//	if err != nil {
//		return nil, err
//	}
//
//	return &model.Protocols{
//		UDP:       udp,
//		TCP:       tcp,
//		AllowIcmp: rawMap[attr.AllowIcmp].(bool),
//	}, nil
//}
//
//func convertProtocol(rawList []interface{}) (*model.Protocol, error) {
//	if len(rawList) == 0 {
//		return nil, nil //nolint:nilnil
//	}
//
//	rawMap := rawList[0].(map[string]interface{})
//	policy := rawMap[attr.Policy].(string)
//
//	ports, err := convertPorts(rawMap[attr.Ports].([]interface{}))
//	if err != nil {
//		return nil, err
//	}
//
//	switch policy {
//	case model.PolicyAllowAll:
//		if len(ports) > 0 {
//			return nil, ErrPortsWithPolicyAllowAll
//		}
//
//	case model.PolicyDenyAll:
//		if len(ports) > 0 {
//			return nil, ErrPortsWithPolicyDenyAll
//		}
//
//	case model.PolicyRestricted:
//		if len(ports) == 0 {
//			return nil, ErrPolicyRestrictedWithoutPorts
//		}
//	}
//
//	if policy == model.PolicyDenyAll {
//		policy = model.PolicyRestricted
//	}
//
//	return model.NewProtocol(policy, ports), nil
//}
//
//func convertPorts(rawList []interface{}) ([]*model.PortRange, error) {
//	var ports = make([]*model.PortRange, 0, len(rawList))
//
//	for _, port := range rawList {
//		var str string
//		if port != nil {
//			str = port.(string)
//		}
//
//		portRange, err := model.NewPortRange(str)
//		if err != nil {
//			return nil, err //nolint:wrapcheck
//		}
//
//		ports = append(ports, portRange)
//	}
//
//	return ports, nil
//}
