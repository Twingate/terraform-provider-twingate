package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	ErrUsersDatasourceShouldSetOneOptionalEmailAttribute     = errors.New("Only one of email, email_regex, email_contains, email_exclude, email_prefix or email_suffix must be set.")
	ErrUsersDatasourceShouldSetOneOptionalFirstNameAttribute = errors.New("Only one of first_name, first_name_regex, first_name_contains, first_name_exclude, first_name_prefix or first_name_suffix must be set.")
	ErrUsersDatasourceShouldSetOneOptionalLastNameAttribute  = errors.New("Only one of last_name, last_name_regex, last_name_contains, last_name_exclude, last_name_prefix or last_name_suffix must be set.")
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &users{}

func NewUsersDatasource() datasource.DataSource {
	return &users{}
}

type users struct {
	client *client.Client
}

type usersModel struct {
	ID                types.String `tfsdk:"id"`
	Email             types.String `tfsdk:"email"`
	EmailRegexp       types.String `tfsdk:"email_regexp"`
	EmailContains     types.String `tfsdk:"email_contains"`
	EmailExclude      types.String `tfsdk:"email_exclude"`
	EmailPrefix       types.String `tfsdk:"email_prefix"`
	EmailSuffix       types.String `tfsdk:"email_suffix"`
	FirstName         types.String `tfsdk:"first_name"`
	FirstNameRegexp   types.String `tfsdk:"first_name_regexp"`
	FirstNameContains types.String `tfsdk:"first_name_contains"`
	FirstNameExclude  types.String `tfsdk:"first_name_exclude"`
	FirstNamePrefix   types.String `tfsdk:"first_name_prefix"`
	FirstNameSuffix   types.String `tfsdk:"first_name_suffix"`
	LastName          types.String `tfsdk:"last_name"`
	LastNameRegexp    types.String `tfsdk:"last_name_regexp"`
	LastNameContains  types.String `tfsdk:"last_name_contains"`
	LastNameExclude   types.String `tfsdk:"last_name_exclude"`
	LastNamePrefix    types.String `tfsdk:"last_name_prefix"`
	LastNameSuffix    types.String `tfsdk:"last_name_suffix"`
	Roles             types.Set    `tfsdk:"roles"`
	Users             []userModel  `tfsdk:"users"`
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

//nolint:funlen
func (d *users) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: userDescription,
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: computedDatasourceIDDescription,
			},

			// email

			attr.Email: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only users that exactly match this email.",
			},
			attr.Email + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the email of the user.",
			},
			attr.Email + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exist in the email of the user.",
			},
			attr.Email + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value does not exist in the email of the user.",
			},
			attr.Email + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The email of the user must start with the value.",
			},
			attr.Email + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The email of the user must end with the value.",
			},

			// first name

			attr.FirstName: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only users that exactly match the first name.",
			},
			attr.FirstName + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the first name of the user.",
			},
			attr.FirstName + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exist in the first name of the user.",
			},
			attr.FirstName + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value does not exist in the first name of the user.",
			},
			attr.FirstName + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The first name of the user must start with the value.",
			},
			attr.FirstName + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The first name of the user must end with the value.",
			},

			// last name

			attr.LastName: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only users that exactly match the last name.",
			},
			attr.LastName + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the last name of the user.",
			},
			attr.LastName + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exist in the last name of the user.",
			},
			attr.LastName + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value does not exist in the last name of the user.",
			},
			attr.LastName + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The last name of the user must start with the value.",
			},
			attr.LastName + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The last name of the user must end with the value.",
			},

			attr.Roles: schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Returns users that match a list of roles. Valid roles: `ADMIN`, `DEVOPS`, `SUPPORT`, `MEMBER`.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf(model.UserRoles...)),
				},
			},

			attr.Users: schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the User",
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
						attr.Role: schema.StringAttribute{
							Computed:    true,
							Description: fmt.Sprintf("Indicates the User's role. Either %s, %s, %s, or %s.", model.UserRoleAdmin, model.UserRoleDevops, model.UserRoleSupport, model.UserRoleMember),
						},
						attr.Type: schema.StringAttribute{
							Computed:    true,
							Description: fmt.Sprintf("Indicates the User's type. Either %s.", utils.DocList(model.UserTypes)),
						},
					},
				},
			},
		},
	}
}

func (d *users) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) { //nolint
	var data usersModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var email, emailFilter, firstName, firstNameFilter, lastName, lastNameFilter string

	// email

	if data.Email.ValueString() != "" {
		email = data.Email.ValueString()
	}

	if data.EmailRegexp.ValueString() != "" {
		email = data.EmailRegexp.ValueString()
		emailFilter = attr.FilterByRegexp
	}

	if data.EmailContains.ValueString() != "" {
		email = data.EmailContains.ValueString()
		emailFilter = attr.FilterByContains
	}

	if data.EmailExclude.ValueString() != "" {
		email = data.EmailExclude.ValueString()
		emailFilter = attr.FilterByExclude
	}

	if data.EmailPrefix.ValueString() != "" {
		email = data.EmailPrefix.ValueString()
		emailFilter = attr.FilterByPrefix
	}

	if data.EmailSuffix.ValueString() != "" {
		email = data.EmailSuffix.ValueString()
		emailFilter = attr.FilterBySuffix
	}

	// first name

	if data.FirstName.ValueString() != "" {
		firstName = data.FirstName.ValueString()
	}

	if data.FirstNameRegexp.ValueString() != "" {
		firstName = data.FirstNameRegexp.ValueString()
		firstNameFilter = attr.FilterByRegexp
	}

	if data.FirstNameContains.ValueString() != "" {
		firstName = data.FirstNameContains.ValueString()
		firstNameFilter = attr.FilterByContains
	}

	if data.FirstNameExclude.ValueString() != "" {
		firstName = data.FirstNameExclude.ValueString()
		firstNameFilter = attr.FilterByExclude
	}

	if data.FirstNamePrefix.ValueString() != "" {
		firstName = data.FirstNamePrefix.ValueString()
		firstNameFilter = attr.FilterByPrefix
	}

	if data.FirstNameSuffix.ValueString() != "" {
		firstName = data.FirstNameSuffix.ValueString()
		firstNameFilter = attr.FilterBySuffix
	}

	// last name

	if data.LastName.ValueString() != "" {
		lastName = data.LastName.ValueString()
	}

	if data.LastNameRegexp.ValueString() != "" {
		lastName = data.LastNameRegexp.ValueString()
		lastNameFilter = attr.FilterByRegexp
	}

	if data.LastNameContains.ValueString() != "" {
		lastName = data.LastNameContains.ValueString()
		lastNameFilter = attr.FilterByContains
	}

	if data.LastNameExclude.ValueString() != "" {
		lastName = data.LastNameExclude.ValueString()
		lastNameFilter = attr.FilterByExclude
	}

	if data.LastNamePrefix.ValueString() != "" {
		lastName = data.LastNamePrefix.ValueString()
		lastNameFilter = attr.FilterByPrefix
	}

	if data.LastNameSuffix.ValueString() != "" {
		lastName = data.LastNameSuffix.ValueString()
		lastNameFilter = attr.FilterBySuffix
	}

	if countOptionalAttributes(data.Email, data.EmailRegexp, data.EmailContains, data.EmailExclude, data.EmailPrefix, data.EmailSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrUsersDatasourceShouldSetOneOptionalEmailAttribute, TwingateResources)

		return
	}

	if countOptionalAttributes(data.FirstName, data.FirstNameRegexp, data.FirstNameContains, data.FirstNameExclude, data.FirstNamePrefix, data.FirstNameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrUsersDatasourceShouldSetOneOptionalFirstNameAttribute, TwingateResources)

		return
	}

	if countOptionalAttributes(data.LastName, data.LastNameRegexp, data.LastNameContains, data.LastNameExclude, data.LastNamePrefix, data.LastNameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrUsersDatasourceShouldSetOneOptionalLastNameAttribute, TwingateResources)

		return
	}

	var filter *client.UsersFilter

	if email != "" || firstName != "" || lastName != "" || len(data.Roles.Elements()) > 0 {
		filter = &client.UsersFilter{}
	}

	if email != "" {
		filter.Email = &client.StringFilter{Name: email, Filter: emailFilter}
	}

	if firstName != "" {
		filter.FirstName = &client.StringFilter{Name: firstName, Filter: firstNameFilter}
	}

	if lastName != "" {
		filter.LastName = &client.StringFilter{Name: lastName, Filter: lastNameFilter}
	}

	if len(data.Roles.Elements()) > 0 {
		filter.Roles = utils.Map(data.Roles.Elements(), func(item tfattr.Value) string {
			return item.(types.String).ValueString()
		})
	}

	users, err := d.client.ReadUsers(ctx, filter)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateUsers)

		return
	}

	data.ID = types.StringValue("users-all")
	data.Users = convertUsersToTerraform(users)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
