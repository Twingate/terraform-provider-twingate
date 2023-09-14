package resource

import (
	"context"
	"errors"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrInvalidExpirationTime = errors.New("Invalid key expiration time. A value from 0-365 is required.")

// Ensure the implementation satisfies the desired interfaces.
var _ resource.Resource = &serviceKey{}

func NewServiceKeyResource() resource.Resource {
	return &serviceKey{}
}

type serviceKey struct {
	client *client.Client
}

type serviceKeyModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	Name             types.String `tfsdk:"name"`
	Token            types.String `tfsdk:"token"`
	IsActive         types.Bool   `tfsdk:"is_active"`
	ExpirationTime   types.Int64  `tfsdk:"expiration_time"`
}

func (r *serviceKey) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateServiceAccountKey
}

func (r *serviceKey) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

func (r *serviceKey) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Service Key authorizes access to all Resources assigned to a Service Account.",
		Attributes: map[string]schema.Attribute{
			attr.ServiceAccountID: schema.StringAttribute{
				Required:    true,
				Description: "The id of the Service Account",
			},
			// optional
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the Service Key",
			},
			attr.ExpirationTime: schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Specifies how many days until a Service Account Key expires. This should be an integer between 0 and 365 representing the number of days until the Service Account Key will expire. Defaults to 0, meaning the key will never expire.",
				Default:     int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			// computed
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "Autogenerated Service Key ID",
			},
			attr.Token: schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Autogenerated Service Key token. Used to configure a Twingate Client running in headless mode.",
			},
			attr.IsActive: schema.BoolAttribute{
				Computed:    true,
				Description: "If the value of this attribute changes to false, Terraform will destroy and recreate the resource.",
			},
		},
	}
}

func (r *serviceKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	expirationTime := int(plan.ExpirationTime.ValueInt64())
	if expirationTime > 365 || expirationTime < 0 {
		addErr(&resp.Diagnostics, ErrInvalidExpirationTime, operationCreate, TwingateServiceAccountKey)

		return
	}

	serviceKey, err := r.client.CreateServiceKey(ctx, &model.ServiceKey{
		Service:        plan.ServiceAccountID.ValueString(),
		Name:           plan.Name.ValueString(),
		ExpirationTime: expirationTime,
	})

	if err == nil && serviceKey != nil {
		plan.Token = types.StringValue(serviceKey.Token)
	}

	r.helper(ctx, serviceKey, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func (r *serviceKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceKey, err := r.client.ReadServiceKey(ctx, state.ID.ValueString())

	r.helper(ctx, serviceKey, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

func (r *serviceKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceKey, err := r.client.UpdateServiceKey(ctx,
		&model.ServiceKey{
			ID:   state.ID.ValueString(),
			Name: plan.Name.ValueString(),
		},
	)

	r.helper(ctx, serviceKey, &state, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func (r *serviceKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceKey, err := r.client.ReadServiceKey(ctx, state.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationDelete, TwingateServiceAccountKey)

		return
	}

	if serviceKey.IsActive() {
		if err = r.client.RevokeServiceKey(ctx, state.ID.ValueString()); err != nil {
			addErr(&resp.Diagnostics, err, operationDelete, TwingateServiceAccountKey)

			return
		}
	}

	err = r.client.DeleteServiceKey(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationDelete, TwingateServiceAccountKey)
}

func (r *serviceKey) helper(ctx context.Context, serviceKey *model.ServiceKey, state *serviceKeyModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			respState.RemoveResource(ctx)

			return
		}

		addErr(diagnostics, err, operation, TwingateServiceAccountKey)

		return
	}

	state.ID = types.StringValue(serviceKey.ID)
	state.Name = types.StringValue(serviceKey.Name)
	state.ServiceAccountID = types.StringValue(serviceKey.Service)
	state.IsActive = types.BoolValue(serviceKey.IsActive())

	// Set refreshed state
	diags := respState.Set(ctx, state)
	diagnostics.Append(diags...)
}
