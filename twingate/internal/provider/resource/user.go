package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrAllowedToChangeOnlyManualUsers = fmt.Errorf("only users of type %s may be modified", model.UserTypeManual)

// Ensure the implementation satisfies the desired interfaces.
var _ resource.Resource = &user{}

func NewUserResource() resource.Resource {
	return &user{}
}

type user struct {
	client *client.Client
}

type userModel struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	FirstName  types.String `tfsdk:"first_name"`
	LastName   types.String `tfsdk:"last_name"`
	SendInvite types.Bool   `tfsdk:"send_invite"`
	IsActive   types.Bool   `tfsdk:"is_active"`
	Role       types.String `tfsdk:"role"`
	Type       types.String `tfsdk:"type"`
}

func (r *user) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = TwingateUser
}

func (r *user) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*client.Client)
}

func (r *user) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Users provides different levels of write capabilities across the Twingate Admin Console. For more information, see Twingate's [documentation](https://www.twingate.com/docs/users).",
		Attributes: map[string]schema.Attribute{
			attr.Email: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The User's email address",
			},
			// optional
			attr.FirstName: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The User's first name",
			},
			attr.LastName: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The User's last name",
			},
			attr.SendInvite: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Determines whether to send an email invitation to the User. True by default.",
			},
			attr.IsActive: schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Determines whether the User is active or not. Inactive users will be not able to sign in.",
			},
			attr.Role: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("Determines the User's role. Either %s.", utils.DocList(model.UserRoles)),
				Validators: []validator.String{
					stringvalidator.OneOf(model.UserRoles...),
				},
			},
			// computed
			attr.Type: schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Indicates the User's type. Either %s.", utils.DocList(model.UserTypes)),
			},
			attr.ID: schema.StringAttribute{
				Computed:      true,
				Description:   "Autogenerated ID of the User, encoded in base64.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *user) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.CreateUser(ctx, &model.User{
		Email:      plan.Email.ValueString(),
		FirstName:  plan.FirstName.ValueString(),
		LastName:   plan.LastName.ValueString(),
		SendInvite: convertSendInviteFlag(plan.SendInvite),
		Role:       withDefaultValue(plan.Role.ValueString(), model.UserRoleMember),
		IsActive:   convertIsActiveFlag(plan.IsActive),
	})

	r.helper(ctx, user, &plan, &resp.State, &resp.Diagnostics, err, operationCreate)
}

func convertSendInviteFlag(val types.Bool) bool {
	if !val.IsUnknown() {
		return val.ValueBool()
	}

	// default value
	return true
}

func convertIsActiveFlag(val types.Bool) bool {
	if !val.IsUnknown() {
		return val.ValueBool()
	}

	// default value
	return true
}

func (r *user) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.ReadUser(ctx, state.ID.ValueString())

	r.helper(ctx, user, &state, &resp.State, &resp.Diagnostics, err, operationRead)
}

func (r *user) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	addErr(&resp.Diagnostics, isAllowedToChangeUser(&state), operationUpdate, TwingateUser)

	if resp.Diagnostics.HasError() {
		return
	}

	userUpdateReq := &model.UserUpdate{
		ID: state.ID.ValueString(),
	}

	if plan.FirstName.ValueString() != "" && state.FirstName != plan.FirstName {
		userUpdateReq.FirstName = plan.FirstName.ValueStringPointer()
	}

	if plan.LastName.ValueString() != "" && state.LastName != plan.LastName {
		userUpdateReq.LastName = plan.LastName.ValueStringPointer()
	}

	if plan.Role.ValueString() != "" && state.Role != plan.Role {
		userUpdateReq.Role = plan.Role.ValueStringPointer()
	}

	isActive := convertIsActiveFlag(plan.IsActive)
	if state.IsActive.ValueBool() != isActive {
		userUpdateReq.IsActive = &isActive
	}

	user, err := r.client.UpdateUser(ctx, userUpdateReq)

	r.helper(ctx, user, &state, &resp.State, &resp.Diagnostics, err, operationUpdate)
}

func isAllowedToChangeUser(state *userModel) error {
	if state.Type.ValueString() != model.UserTypeManual {
		return ErrAllowedToChangeOnlyManualUsers
	}

	return nil
}

func (r *user) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	addErr(&resp.Diagnostics, isAllowedToChangeUser(&state), operationDelete, TwingateUser)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(ctx, state.ID.ValueString())
	addErr(&resp.Diagnostics, err, operationDelete, TwingateUser)
}

func (r *user) helper(ctx context.Context, user *model.User, state *userModel, respState *tfsdk.State, diagnostics *diag.Diagnostics, err error, operation string) {
	if err != nil {
		if errors.Is(err, client.ErrGraphqlResultIsEmpty) {
			// clear state
			respState.RemoveResource(ctx)

			return
		}

		addErr(diagnostics, err, operation, TwingateUser)

		return
	}

	state.ID = types.StringValue(user.ID)
	state.FirstName = types.StringValue(user.FirstName)
	state.LastName = types.StringValue(user.LastName)
	state.Role = types.StringValue(user.Role)
	state.Type = types.StringValue(user.Type)
	state.IsActive = types.BoolValue(user.IsActive)

	// Set refreshed state
	diags := respState.Set(ctx, state)
	diagnostics.Append(diags...)
}
