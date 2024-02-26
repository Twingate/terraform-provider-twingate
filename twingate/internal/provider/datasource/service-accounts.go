package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrServiceAccountsDatasourceShouldSetOneOptionalNameAttribute = errors.New("Only one of name, name_regex, name_contains, name_exclude, name_prefix or name_suffix must be set.")

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &serviceAccounts{}

func NewServiceAccountsDatasource() datasource.DataSource {
	return &serviceAccounts{}
}

type serviceAccounts struct {
	client *client.Client
}

type serviceAccountsModel struct {
	ID              types.String          `tfsdk:"id"`
	Name            types.String          `tfsdk:"name"`
	NameRegexp      types.String          `tfsdk:"name_regexp"`
	NameContains    types.String          `tfsdk:"name_contains"`
	NameExclude     types.String          `tfsdk:"name_exclude"`
	NamePrefix      types.String          `tfsdk:"name_prefix"`
	NameSuffix      types.String          `tfsdk:"name_suffix"`
	ServiceAccounts []serviceAccountModel `tfsdk:"service_accounts"`
}

type serviceAccountModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	ResourceIDs []types.String `tfsdk:"resource_ids"`
	KeyIDs      []types.String `tfsdk:"key_ids"`
}

func (d *serviceAccounts) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateServiceAccounts
}

func (d *serviceAccounts) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serviceAccounts) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Service Accounts offer a way to provide programmatic, centrally-controlled, and consistent access controls.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: computedDatasourceIDDescription,
			},
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only service accounts that exactly match this name. If no options are passed it will return all service accounts. Only one option can be used at a time.",
			},
			attr.Name + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the name of the service account.",
			},
			attr.Name + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exist in the name of the service account.",
			},
			attr.Name + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the exact value does not exist in the name of the service account.",
			},
			attr.Name + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the service account must start with the value.",
			},
			attr.Name + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the service account must end with the value.",
			},
			attr.ServiceAccounts: schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Service Accounts",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "ID of the Service Account resource",
						},
						attr.Name: schema.StringAttribute{
							Computed:    true,
							Description: "Name of the Service Account",
						},
						attr.ResourceIDs: schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of twingate_resource IDs that the Service Account is assigned to.",
						},
						attr.KeyIDs: schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of twingate_service_account_key IDs that are assigned to the Service Account.",
						},
					},
				},
			},
		},
	}
}

func (d *serviceAccounts) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceAccountsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name, filter := getNameFilter(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix)

	if countOptionalAttributes(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrServiceAccountsDatasourceShouldSetOneOptionalNameAttribute, TwingateResources)

		return
	}

	accounts, err := d.client.ReadServiceAccounts(ctx, name, filter)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateServiceAccounts)

		return
	}

	data.ID = types.StringValue(terraformServicesDatasourceID(data.Name.ValueString()))
	data.ServiceAccounts = convertServicesToTerraform(accounts)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func terraformServicesDatasourceID(name string) string {
	id := "all-services"
	if name != "" {
		id = "service-by-name-" + name
	}

	return id
}
