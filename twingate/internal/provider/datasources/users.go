package datasources

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

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &user{}

func NewUsersDatasource() datasource.DataSource {
	return &users{}
}

type users struct {
	client *client.Client
}

type usersModel struct {
	ID    types.String `tfsdk:"id"`
	Users []userModel  `tfsdk:"users"`
}

func (d *users) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateUsers
}

func (d *users) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *users) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: userDescription,
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Users datasource.",
			},

			attr.Users: schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Required:    true,
							Description: "The ID of the User. The ID for the User can be obtained from the Admin API or the URL string in the Admin Console.",
						},
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
							Description: fmt.Sprintf("Indicates the User's role. Either %s", utils.DocList(model.UserRoles)),
						},
						attr.Type: schema.StringAttribute{
							Computed:    true,
							Description: fmt.Sprintf("Indicates the User's type. Either %s", utils.DocList(model.UserTypes)),
						},
					},
				},
			},
		},
	}
}

func (d *users) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	users, err := d.client.ReadUsers(ctx)
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, TwingateUsers)
		return
	}

	data := usersModel{
		ID:    types.StringValue("users-all"),
		Users: convertUsersToTerraform(users),
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func convertUsersToTerraform(users []*model.User) []userModel {
	return utils.Map(users, func(user *model.User) userModel {
		return userModel{
			ID:        types.StringValue(user.ID),
			FirstName: types.StringValue(user.FirstName),
			LastName:  types.StringValue(user.LastName),
			Email:     types.StringValue(user.Email),
			IsAdmin:   types.BoolValue(user.IsAdmin()),
			Role:      types.StringValue(user.Role),
			Type:      types.StringValue(user.Type),
		}
	})
}
