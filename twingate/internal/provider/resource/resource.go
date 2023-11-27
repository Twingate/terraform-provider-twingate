package resource

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

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

var (
	ErrPortsWithPolicyAllowAll            = errors.New(model.PolicyAllowAll + " policy does not allow specifying ports.")
	ErrPortsWithPolicyDenyAll             = errors.New(model.PolicyDenyAll + " policy does not allow specifying ports.")
	ErrPolicyRestrictedWithoutPorts       = errors.New(model.PolicyRestricted + " policy requires specifying ports.")
	ErrInvalidAttributeCombination        = errors.New("invalid attribute combination")
	ErrWildcardAddressWithEnabledShortcut = errors.New("Resources with a CIDR range or wildcard can't have the browser shortcut enabled.")
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
	Access                   types.List   `tfsdk:"access"`
	IsVisible                types.Bool   `tfsdk:"is_visible"`
	IsBrowserShortcutEnabled types.Bool   `tfsdk:"is_browser_shortcut_enabled"`
	Alias                    types.String `tfsdk:"alias"`
	SecurityPolicyID         types.String `tfsdk:"security_policy_id"`
}

type resourceModelV0 struct {
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
	SecurityPolicyID         types.String `tfsdk:"security_policy_id"`
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

	if len(res.Groups) > 0 || len(res.ServiceAccounts) > 0 {
		access, diags := convertAccessBlockToTerraform(ctx, res, types.SetNull(types.StringType), types.SetNull(types.StringType))

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.State.SetAttribute(ctx, path.Root(attr.Access), access)
	}
}

func (r *twingateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
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
				Optional:    true,
				Computed:    true,
				Description: "The ID of a `twingate_security_policy` to set as this Resource's Security Policy.",
				Default:     stringdefault.StaticString(""),
			},
			attr.IsVisible: schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Controls whether this Resource will be visible in the main Resource list in the Twingate Client.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			attr.IsBrowserShortcutEnabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: `Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.`,
				Default:     booldefault.StaticBool(false),
			},
			attr.ID: schema.StringAttribute{
				Computed:      true,
				Description:   "Autogenerated ID of the Resource, encoded in base64",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},

		Blocks: map[string]schema.Block{attr.Access: accessBlock()},
	}
}

func (r *twingateResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader { //nolint
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 1 (Schema.Version)
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					attr.ID: schema.StringAttribute{
						Computed: true,
					},
					attr.Name: schema.StringAttribute{
						Required: true,
					},
					attr.Address: schema.StringAttribute{
						Required: true,
					},
					attr.RemoteNetworkID: schema.StringAttribute{
						Required: true,
					},
					attr.IsAuthoritative: schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
					attr.Alias: schema.StringAttribute{
						Optional: true,
					},
					attr.SecurityPolicyID: schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
					attr.IsVisible: schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
					attr.IsBrowserShortcutEnabled: schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},

				Blocks: map[string]schema.Block{
					attr.Access: schema.ListNestedBlock{
						Validators: []validator.List{
							listvalidator.SizeAtMost(1),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								attr.GroupIDs: schema.SetAttribute{
									Optional:    true,
									ElementType: types.StringType,
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
								},
								attr.ServiceAccountIDs: schema.SetAttribute{
									Optional:    true,
									ElementType: types.StringType,
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
									Optional: true,
									Computed: true,
								},
							},
							Blocks: map[string]schema.Block{
								attr.UDP: schema.ListNestedBlock{
									Validators: []validator.List{
										listvalidator.SizeAtMost(1),
									},
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											attr.Policy: schema.StringAttribute{
												Optional: true,
												Computed: true,
											},
											attr.Ports: schema.SetAttribute{
												Optional:    true,
												Computed:    true,
												ElementType: types.StringType,
											},
										},
									},
								},
								attr.TCP: schema.ListNestedBlock{
									Validators: []validator.List{
										listvalidator.SizeAtMost(1),
									},
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											attr.Policy: schema.StringAttribute{
												Optional: true,
												Computed: true,
											},
											attr.Ports: schema.SetAttribute{
												Optional:    true,
												Computed:    true,
												ElementType: types.StringType,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorState resourceModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)

				if resp.Diagnostics.HasError() {
					return
				}

				protocols, err := convertProtocolsV0(priorState.Protocols)
				if err != nil {
					resp.Diagnostics.AddError(
						"failed to convert protocols for prior state version 0",
						err.Error(),
					)

					return
				}

				protocolsState, diags := convertProtocolsToTerraform(protocols, nil)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedState := resourceModel{
					ID:              priorState.ID,
					Name:            priorState.Name,
					Address:         priorState.Address,
					RemoteNetworkID: priorState.RemoteNetworkID,
					Protocols:       protocolsState,
					Access:          priorState.Access,
				}

				if !priorState.IsAuthoritative.IsNull() {
					upgradedState.IsAuthoritative = priorState.IsAuthoritative
				}

				if !priorState.IsVisible.IsNull() {
					upgradedState.IsVisible = priorState.IsVisible
				}

				if !priorState.IsBrowserShortcutEnabled.IsNull() {
					upgradedState.IsBrowserShortcutEnabled = priorState.IsBrowserShortcutEnabled
				}

				if !priorState.Alias.IsNull() && priorState.Alias.ValueString() != "" {
					upgradedState.Alias = priorState.Alias
				}

				if !priorState.SecurityPolicyID.IsNull() && priorState.SecurityPolicyID.ValueString() != "" {
					upgradedState.SecurityPolicyID = priorState.SecurityPolicyID
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)

				resp.Diagnostics.AddWarning("Please upgrade protocols sections", "Follow this docs to update protocols from blocks to attributes")
			},
		},
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

func accessBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
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

	if err = r.client.AddResourceAccess(ctx, resource.ID, resource.ServiceAccounts); err != nil {
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

	isBrowserShortcutEnabled := getOptionalBool(plan.IsBrowserShortcutEnabled)

	if isBrowserShortcutEnabled != nil && *isBrowserShortcutEnabled && isWildcardAddress(plan.Address.ValueString()) {
		return nil, ErrWildcardAddressWithEnabledShortcut
	}

	return &model.Resource{
		Name:                     plan.Name.ValueString(),
		RemoteNetworkID:          plan.RemoteNetworkID.ValueString(),
		Address:                  plan.Address.ValueString(),
		Protocols:                protocols,
		Groups:                   groupIDs,
		ServiceAccounts:          serviceAccountIDs,
		IsAuthoritative:          convertAuthoritativeFlag(plan.IsAuthoritative),
		Alias:                    getOptionalString(plan.Alias),
		IsVisible:                getOptionalBool(plan.IsVisible),
		IsBrowserShortcutEnabled: isBrowserShortcutEnabled,
		SecurityPolicyID:         plan.SecurityPolicyID.ValueString(),
	}, nil
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

func convertProtocolsV0(protocols types.List) (*model.Protocols, error) {
	if protocols.IsNull() || protocols.IsUnknown() || len(protocols.Elements()) == 0 {
		return model.DefaultProtocols(), nil
	}

	obj := protocols.Elements()[0].(types.Object)
	if obj.IsNull() || obj.IsUnknown() {
		return model.DefaultProtocols(), nil
	}

	udp, err := convertProtocolV0(obj.Attributes()[attr.UDP])
	if err != nil {
		return nil, err
	}

	tcp, err := convertProtocolV0(obj.Attributes()[attr.TCP])
	if err != nil {
		return nil, err
	}

	return &model.Protocols{
		AllowIcmp: obj.Attributes()[attr.AllowIcmp].(types.Bool).ValueBool(),
		UDP:       udp,
		TCP:       tcp,
	}, nil
}

func convertProtocolV0(protocol tfattr.Value) (*model.Protocol, error) {
	obj := convertProtocolObjV0(protocol)
	if obj.IsNull() {
		return nil, nil //nolint:nilnil
	}

	ports, err := decodePortsV0(obj)
	if err != nil {
		return nil, err
	}

	policy := obj.Attributes()[attr.Policy].(types.String).ValueString()
	if err := isValidPolicyV0(policy, ports); err != nil {
		return nil, err
	}

	if policy == model.PolicyDenyAll {
		policy = model.PolicyRestricted
	}

	return model.NewProtocol(policy, ports), nil
}

func convertProtocolObjV0(protocol tfattr.Value) types.Object {
	if protocol == nil || protocol.IsNull() {
		return types.ObjectNull(nil)
	}

	list, ok := protocol.(types.List)
	if !ok || list.IsNull() || list.IsUnknown() || len(list.Elements()) == 0 {
		return types.ObjectNull(nil)
	}

	obj := list.Elements()[0].(types.Object)
	if obj.IsNull() || obj.IsUnknown() {
		return types.ObjectNull(nil)
	}

	return obj
}

func decodePortsV0(obj types.Object) ([]*model.PortRange, error) {
	portsVal := obj.Attributes()[attr.Ports]
	if portsVal == nil || portsVal.IsNull() {
		return nil, nil
	}

	portsList, ok := portsVal.(types.Set)
	if !ok {
		return nil, nil
	}

	return convertPortsV0(portsList)
}

func convertPortsV0(list types.Set) ([]*model.PortRange, error) {
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

func isValidPolicyV0(policy string, ports []*model.PortRange) error {
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

		if state.SecurityPolicyID.ValueString() == "" {
			resource.SecurityPolicyID = ""
		}
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
		if err := r.updateResourceAccess(ctx, &plan, &state, input); err != nil {
			addErr(&resp.Diagnostics, err, operationUpdate, TwingateResource)

			return
		}
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
	return !plan.RemoteNetworkID.Equal(state.RemoteNetworkID) ||
		!plan.Name.Equal(state.Name) ||
		!plan.Address.Equal(state.Address) ||
		!equalProtocolsState(&plan.Protocols, &state.Protocols) ||
		!plan.IsVisible.Equal(state.IsVisible) ||
		!plan.IsBrowserShortcutEnabled.Equal(state.IsBrowserShortcutEnabled) ||
		!plan.Alias.Equal(state.Alias) ||
		!plan.SecurityPolicyID.Equal(state.SecurityPolicyID)
}

func (r *twingateResource) updateResourceAccess(ctx context.Context, plan, state *resourceModel, input *model.Resource) error {
	idsToDelete, idsToAdd, err := r.getChangedAccessIDs(ctx, plan, state, input)
	if err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	if err := r.client.RemoveResourceAccess(ctx, input.ID, idsToDelete); err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	if err := r.client.AddResourceAccess(ctx, input.ID, idsToAdd); err != nil {
		return fmt.Errorf("failed to update resource access: %w", err)
	}

	return nil
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

	if !resource.IsAuthoritative {
		resource.Groups = setIntersection(getAccessAttribute(reference.Access, attr.GroupIDs), resource.Groups)
		resource.ServiceAccounts = setIntersection(getAccessAttribute(reference.Access, attr.ServiceAccountIDs), resource.ServiceAccounts)
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
	state.IsAuthoritative = types.BoolValue(resource.IsAuthoritative)
	state.SecurityPolicyID = types.StringValue(resource.SecurityPolicyID)

	if !state.IsVisible.IsNull() || !reference.IsVisible.IsUnknown() {
		state.IsVisible = types.BoolPointerValue(resource.IsVisible)
	}

	if !state.IsBrowserShortcutEnabled.IsNull() || !reference.IsBrowserShortcutEnabled.IsUnknown() {
		state.IsBrowserShortcutEnabled = types.BoolPointerValue(resource.IsBrowserShortcutEnabled)
	}

	if !state.Alias.IsNull() || !reference.Alias.IsUnknown() {
		state.Alias = reference.Alias
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
	return types.SetNull(types.StringType)
}

func defaultProtocolModelToTerraform() (basetypes.ObjectValue, diag.Diagnostics) {
	attributes := map[string]tfattr.Value{
		attr.Policy: types.StringValue(model.PolicyAllowAll),
		attr.Ports:  types.SetNull(types.StringType),
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

func makeNullObject(attributeTypes map[string]tfattr.Type) types.Object {
	return types.ObjectNull(attributeTypes)
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

var cidrRgxp = regexp.MustCompile(`(\d{1,3}\.){3}\d{1,3}(/\d+)?`)

func isWildcardAddress(address string) bool {
	return strings.ContainsAny(address, "*?") || cidrRgxp.MatchString(address)
}
