package resource

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatorfuncerr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const minLengthConnectorName = 3

var ErrNotAllowChangeRemoteNetworkID = errors.New("connectors cannot be moved between Remote Networks: you must either create a new Connector or destroy and recreate the existing one")

// Ensure the implementation satisfies the desired interfaces.
var _ resource.Resource = &connector{}
var _ resource.ResourceWithImportState = &connector{}

var spacesRgx = regexp.MustCompile(`\s+`)

func NewConnectorResource() resource.Resource {
	return &connector{}
}

type connector struct {
	client *client.Client
}

type connectorModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	RemoteNetworkID      types.String `tfsdk:"remote_network_id"`
	StatusUpdatesEnabled types.Bool   `tfsdk:"status_updates_enabled"`
	State                types.String `tfsdk:"state"`
	Hostname             types.String `tfsdk:"hostname"`
	Version              types.String `tfsdk:"version"`
	PublicIP             types.String `tfsdk:"public_ip"`
	PrivateIPs           types.Set    `tfsdk:"private_ips"`
}

func (r *connector) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateConnector
}

func (r *connector) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

func (r *connector) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(attr.ID), req, resp)

	con, err := r.client.ReadConnector(ctx, req.ID)
	if err != nil {
		addErr(&resp.Diagnostics, err, "import", TwingateConnector)

		return
	}

	resp.State.SetAttribute(ctx, path.Root(attr.Name), types.StringValue(con.Name))
}

func (r *connector) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var oldVal, newVal types.String

	diags := req.State.GetAttribute(ctx, path.Root(attr.RemoteNetworkID), &oldVal)
	resp.Diagnostics.Append(diags...)

	diags = req.Plan.GetAttribute(ctx, path.Root(attr.RemoteNetworkID), &newVal)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !oldVal.IsNull() && !newVal.IsNull() && oldVal.String() != newVal.String() {
		resp.Diagnostics.AddAttributeError(path.Root(attr.RemoteNetworkID), ErrNotAllowChangeRemoteNetworkID.Error(), "not allowed to change")
	}
}

func (r *connector) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Connectors provide connectivity to Remote Networks. This resource type will create the Connector in the Twingate Admin Console, but in order to successfully deploy it, you must also generate Connector tokens that authenticate the Connector with Twingate. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/understanding-access-nodes).",
		Attributes: map[string]schema.Attribute{
			attr.RemoteNetworkID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The ID of the Remote Network the Connector is attached to.",
			},
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the Connector, if not provided one will be generated.",
				Validators: []validator.String{
					SanitizedNameLengthValidator(minLengthConnectorName),
				},
				PlanModifiers: []planmodifier.String{
					SanitizeInsensitiveModifier(),
				},
			},
			attr.StatusUpdatesEnabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Determines whether status notifications are enabled for the Connector. Default is `true`.",
			},
			// computed
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated ID of the Connector, encoded in base64.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attr.State: schema.StringAttribute{
				Computed:    true,
				Description: "The Connector's state. One of `ALIVE`, `DEAD_NO_HEARTBEAT`, `DEAD_HEARTBEAT_TOO_OLD` or `DEAD_NO_RELAYS`.",
			},
			attr.Hostname: schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the machine hosting the Connector.",
			},
			attr.Version: schema.StringAttribute{
				Computed:    true,
				Description: "The Connector's version.",
			},
			attr.PublicIP: schema.StringAttribute{
				Computed:    true,
				Description: "The Connector's public IP address.",
			},
			attr.PrivateIPs: schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The Connector's private IP addresses.",
			},
		},
	}
}

func (r *connector) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectorModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn, err := r.client.CreateConnector(ctx, &model.Connector{
		Name:                 sanitizeName(plan.Name.ValueString()),
		NetworkID:            plan.RemoteNetworkID.ValueString(),
		StatusUpdatesEnabled: getOptionalBool(plan.StatusUpdatesEnabled),
	})

	r.helper(ctx, conn, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func (r *connector) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state connectorModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn, err := r.client.ReadConnector(ctx, state.ID.ValueString())

	r.helper(ctx, conn, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

func (r *connector) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan connectorModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// allowed to change `name` and `status_updates_enabled`
	if plan.Name == state.Name && plan.StatusUpdatesEnabled == state.StatusUpdatesEnabled {
		return
	}

	conn := &model.Connector{
		ID:                   state.ID.ValueString(),
		Name:                 sanitizeName(plan.Name.ValueString()),
		StatusUpdatesEnabled: plan.StatusUpdatesEnabled.ValueBoolPointer(),
	}

	if plan.Name == state.Name {
		conn.Name = ""
	}

	conn, err := r.client.UpdateConnector(ctx, conn)

	r.helper(ctx, conn, &plan, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func (r *connector) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectorModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteConnector(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationDelete, TwingateConnector)
}

func (r *connector) helper(ctx context.Context, conn *model.Connector, state *connectorModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			respState.RemoveResource(ctx)

			return
		}

		addErr(diagnostics, err, operation, TwingateConnector)

		return
	}

	if state.Name.IsUnknown() {
		state.Name = types.StringValue(conn.Name)
	}

	state.ID = types.StringValue(conn.ID)
	state.RemoteNetworkID = types.StringValue(conn.NetworkID)
	state.StatusUpdatesEnabled = types.BoolPointerValue(conn.StatusUpdatesEnabled)
	state.State = types.StringValue(conn.State)
	state.Version = types.StringValue(conn.Version)
	state.Hostname = types.StringValue(conn.Hostname)
	state.PublicIP = types.StringValue(conn.PublicIP)
	state.PrivateIPs = utils.MakeStringSet(conn.PrivateIPs)

	// Set refreshed state
	diags := respState.Set(ctx, state)
	diagnostics.Append(diags...)
}

func SanitizeInsensitiveModifier() planmodifier.String {
	return sanitizeInsensitiveModifier{}
}

type sanitizeInsensitiveModifier struct{}

func (m sanitizeInsensitiveModifier) Description(_ context.Context) string {
	return ""
}

func (m sanitizeInsensitiveModifier) MarkdownDescription(_ context.Context) string {
	return ""
}

func (m sanitizeInsensitiveModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
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

	if strings.EqualFold(sanitizeName(req.PlanValue.ValueString()), sanitizeName(req.StateValue.ValueString())) {
		resp.PlanValue = req.StateValue
	}
}

func sanitizeName(name string) string {
	return strings.TrimSpace(spacesRgx.ReplaceAllString(name, " "))
}

var _ validator.String = sanitizedLengthValidator{}
var _ function.StringParameterValidator = sanitizedLengthValidator{}

type sanitizedLengthValidator struct {
	minLen int
}

func (v sanitizedLengthValidator) Description(_ context.Context) string {
	return fmt.Sprintf("must be at least %d characters long", v.minLen)
}

func (v sanitizedLengthValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v sanitizedLengthValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if len(sanitizeName(value)) < v.minLen {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}

func (v sanitizedLengthValidator) ValidateParameterString(ctx context.Context, request function.StringParameterValidatorRequest, response *function.StringParameterValidatorResponse) {
	if request.Value.IsNull() || request.Value.IsUnknown() {
		return
	}

	value := request.Value.ValueString()

	if len(sanitizeName(value)) < v.minLen {
		response.Error = validatorfuncerr.InvalidParameterValueMatchFuncError(
			request.ArgumentPosition,
			v.Description(ctx),
			value,
		)
	}
}

func SanitizedNameLengthValidator(minLen int) sanitizedLengthValidator {
	return sanitizedLengthValidator{
		minLen: minLen,
	}
}
