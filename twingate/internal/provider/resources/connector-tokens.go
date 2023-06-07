package resources

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewConnectorTokensResource() resource.Resource {
	return &connectorTokens{}
}

type connectorTokens struct {
	client *client.Client
}

type connectorTokensModel struct {
	ID           types.String `tfsdk:"id"`
	ConnectorID  types.String `tfsdk:"connector_id"`
	AccessToken  types.String `tfsdk:"access_token"`
	RefreshToken types.String `tfsdk:"refresh_token"`
	Keepers      types.Map    `tfsdk:"keepers"`
}

func (r *connectorTokens) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateConnectorTokens
}

func (r *connectorTokens) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

func (r *connectorTokens) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource type will generate tokens for a Connector, which are needed to successfully provision one on your network. The Connector itself has its own resource type and must be created before you can provision tokens.",
		Attributes: map[string]schema.Attribute{
			attr.ConnectorID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The ID of the parent Connector",
			},
			// optional
			attr.Keepers: schema.MapAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Map{
					RequiresMapReplace("If the value of this attribute changes, Terraform will destroy and recreate the resource."),
				},
				ElementType: types.StringType,
			},
			// computed
			attr.AccessToken: schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The Access Token of the parent Connector",
			},
			attr.RefreshToken: schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The Refresh Token of the parent Connector",
			},
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "ID of the Connector Tokens.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *connectorTokens) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectorTokensModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tokens, err := r.client.GenerateConnectorTokens(ctx, plan.ConnectorID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, TwingateConnectorTokens)

		return
	}

	plan.ID = plan.ConnectorID
	plan.AccessToken = types.StringValue(tokens.AccessToken)
	plan.RefreshToken = types.StringValue(tokens.RefreshToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.helper(ctx, r.client, plan.ID.ValueString(), tokens.AccessToken, tokens.RefreshToken, &resp.State, resp.Diagnostics)
}

func (r *connectorTokens) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state connectorTokensModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.helper(ctx, r.client, state.ID.ValueString(), state.AccessToken.ValueString(), state.RefreshToken.ValueString(), &resp.State, resp.Diagnostics)
}

func (r *connectorTokens) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// on update, we should re-create the resource
}

func (r *connectorTokens) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectorTokensModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Just calling generate new tokens for the connector so the old ones are invalidated
	_, err := r.client.GenerateConnectorTokens(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationDelete, TwingateConnectorTokens)
}

func (r *connectorTokens) helper(ctx context.Context, client *client.Client, connectorID, accessToken, refreshToken string, state *tfsdk.State, diagnostics diag.Diagnostics) {
	err := client.VerifyConnectorTokens(ctx, refreshToken, accessToken)
	if err != nil {
		state.RemoveResource(ctx)

		diagnostics.AddWarning(
			"can't to verify connector tokens",
			fmt.Sprintf("can't verify connector %s tokens, assuming not valid and needs to be recreated", connectorID),
		)
	}
}

func RequiresMapReplace(description string) *requiresMapReplace {
	return &requiresMapReplace{description: description}
}

type requiresMapReplace struct {
	description string
}

// Description returns a human-readable description of the plan modifier.
func (m requiresMapReplace) Description(_ context.Context) string {
	return m.description
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m requiresMapReplace) MarkdownDescription(_ context.Context) string {
	return m.description
}

// PlanModifyMap implements the plan modification logic.
func (m requiresMapReplace) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// Do not replace on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do not replace on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	// Do not replace if the plan and state values are equal.
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.RequiresReplace = true
}
