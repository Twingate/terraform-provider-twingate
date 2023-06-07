package resources

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	operationRead   = "read"
	operationCreate = "create"
	operationUpdate = "update"
	operationDelete = "delete"
)

var ErrNotAllowChangeRemoteNetworkID = errors.New("connectors cannot be moved between Remote Networks: you must either create a new Connector or destroy and recreate the existing one")

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
			},
			attr.StatusUpdatesEnabled: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Determines whether status notifications are enabled for the Connector.",
			},
			// computed
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated ID of the Connector, encoded in base64.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
		Name:      plan.Name.ValueString(),
		NetworkID: plan.RemoteNetworkID.ValueString(),
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
		Name:                 plan.Name.ValueString(),
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

	state.ID = types.StringValue(conn.ID)
	state.Name = types.StringValue(conn.Name)
	state.RemoteNetworkID = types.StringValue(conn.NetworkID)
	state.StatusUpdatesEnabled = types.BoolPointerValue(conn.StatusUpdatesEnabled)

	// Set refreshed state
	diags := respState.Set(ctx, state)
	diagnostics.Append(diags...)
}

func addErr(diagnostics *diag.Diagnostics, err error, operation, resource string) {
	if err == nil {
		return
	}

	diagnostics.AddError(
		fmt.Sprintf("failed to %s %s", operation, resource),
		err.Error(),
	)
}
