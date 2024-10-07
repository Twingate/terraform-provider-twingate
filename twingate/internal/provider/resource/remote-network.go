package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"


	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ resource.Resource = &remoteNetwork{}

func NewRemoteNetworkResource() resource.Resource {
	return &remoteNetwork{}
}

type remoteNetwork struct {
	client *client.Client
}

type remoteNetworkModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Location types.String `tfsdk:"location"`
	ExitNode types.Bool   `tfsdk:"exit_node"`
}

func (r *remoteNetwork) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateRemoteNetwork
}

func (r *remoteNetwork) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

func (r *remoteNetwork) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(attr.ID), req, resp)
}

func (r *remoteNetwork) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Remote Network represents a single private network in Twingate that can have one or more Connectors and Resources assigned to it. You must create a Remote Network before creating Resources and Connectors that belong to it. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/remote-networks).",
		Attributes: map[string]schema.Attribute{
			attr.Name: schema.StringAttribute{
				Required:    true,
				Description: "The name of the Remote Network",
			},
			attr.Location: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("The location of the Remote Network. Must be one of the following: %s.", strings.Join(model.Locations, ", ")),
				Validators: []validator.String{
					stringvalidator.OneOf(model.Locations...),
				},
			},
			attr.ExitNode: schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "TODO",
				Default:       booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			// computed
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Remote Network",
			},
		},
	}
}

func (r *remoteNetwork) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan remoteNetworkModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// init with default value
	location := model.LocationOther
	if !plan.Location.IsUnknown() {
		location = plan.Location.ValueString()
	}

	network, err := r.client.CreateRemoteNetwork(ctx, &model.RemoteNetwork{
		Name:     plan.Name.ValueString(),
		Location: location,
		ExitNode: plan.ExitNode.ValueBool(),
	})

	r.helper(ctx, network, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func (r *remoteNetwork) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state remoteNetworkModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	network, err := r.client.ReadRemoteNetworkByID(ctx, state.ID.ValueString())

	r.helper(ctx, network, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

func (r *remoteNetwork) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan remoteNetworkModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	network := &model.RemoteNetwork{
		ID:       state.ID.ValueString(),
		Name:     plan.Name.ValueString(),
		Location: plan.Location.ValueString(),
	}

	if plan.Name == state.Name {
		network.Name = ""
	}

	network, err := r.client.UpdateRemoteNetwork(ctx, network)

	r.helper(ctx, network, &plan, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func (r *remoteNetwork) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state remoteNetworkModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRemoteNetwork(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationDelete, TwingateRemoteNetwork)
}

func (r *remoteNetwork) helper(ctx context.Context, network *model.RemoteNetwork, state *remoteNetworkModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			respState.RemoveResource(ctx)

			return
		}

		addErr(diagnostics, err, operation, TwingateRemoteNetwork)

		return
	}

	state.ID = types.StringValue(network.ID)
	state.Name = types.StringValue(network.Name)
	state.Location = types.StringValue(network.Location)

	// Set refreshed state
	diags := respState.Set(ctx, state)
	diagnostics.Append(diags...)
}
