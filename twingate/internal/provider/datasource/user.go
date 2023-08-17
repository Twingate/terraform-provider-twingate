package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const userDescription = "Users in Twingate can be given access to Twingate Resources and may either be added manually or automatically synchronized with a 3rd party identity provider. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/users)."

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &user{}

func NewUserDatasource() datasource.DataSource {
	return &user{}
}

type user struct {
	client *client.Client
}

type userModel struct {
	ID        types.String `tfsdk:"id"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Email     types.String `tfsdk:"email"`
	IsAdmin   types.Bool   `tfsdk:"is_admin"`
	Role      types.String `tfsdk:"role"`
	Type      types.String `tfsdk:"type"`
}

func (d *user) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateUser
}

func (d *user) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *user) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: userDescription,
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the User. The ID for the User can be obtained from the Admin API or the URL string in the Admin Console.",
			},
			// computed
			attr.FirstName: schema.StringAttribute{
				Computed:    true,
				Description: "The first name of the User",
			},
			attr.LastName: schema.StringAttribute{
				Computed:    true,
				Description: "The last name of the User",
			},
			attr.Email: schema.StringAttribute{
				Computed:    true,
				Description: "The email address of the User",
			},
			attr.IsAdmin: schema.BoolAttribute{
				Computed:           true,
				Description:        "Indicates whether the User is an admin",
				DeprecationMessage: "This read-only Boolean value will be deprecated in a future release. You may use the `role` value instead.",
			},
			attr.Role: schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Indicates the User's role. Either %s, %s, %s, or %s", model.UserRoleAdmin, model.UserRoleDevops, model.UserRoleSupport, model.UserRoleMember),
			},
			attr.Type: schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Indicates the User's type. Either %s.", utils.DocList(model.UserTypes)),
			},
		},
	}
}

func (d *user) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := d.client.ReadUser(ctx, data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateUser)

		return
	}

	data.ID = types.StringValue(user.ID)
	data.FirstName = types.StringValue(user.FirstName)
	data.LastName = types.StringValue(user.LastName)
	data.Email = types.StringValue(user.Email)
	data.IsAdmin = types.BoolValue(user.IsAdmin())
	data.Role = types.StringValue(user.Role)
	data.Type = types.StringValue(user.Type)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
