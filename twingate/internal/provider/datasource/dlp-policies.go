package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrDLPPoliciesDatasourceShouldSetOneOptionalNameAttribute = errors.New("Only one of name, name_regex, name_contains, name_exclude, name_prefix or name_suffix must be set.")

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &dlpPolicies{}

func NewDLPPoliciesDatasource() datasource.DataSource {
	return &dlpPolicies{}
}

type dlpPolicies struct {
	client *client.Client
}

type dlpPoliciesModel struct {
	ID           types.String      `tfsdk:"id"`
	Name         types.String      `tfsdk:"name"`
	NameRegexp   types.String      `tfsdk:"name_regexp"`
	NameContains types.String      `tfsdk:"name_contains"`
	NameExclude  types.String      `tfsdk:"name_exclude"`
	NamePrefix   types.String      `tfsdk:"name_prefix"`
	NameSuffix   types.String      `tfsdk:"name_suffix"`
	DLPPolicies  []dlpPolicyModel  `tfsdk:"dlp_policies"`
}

func (d *dlpPolicies) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateDLPPolicies
}

func (d *dlpPolicies) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dlpPolicies) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "DLP Policies are defined in the Twingate Admin Console and determine data loss prevention settings for Resources.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: computedDatasourceIDDescription,
			},
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only DLP policies that exactly match this name. If no options are passed it will return all DLP policies. Only one option can be used at a time.",
			},
			attr.Name + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "The regular expression match of the name of the DLP policy.",
			},
			attr.Name + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the value exists in the name of the DLP policy.",
			},
			attr.Name + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Match when the exact value does not exist in the name of the DLP policy.",
			},
			attr.Name + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the DLP policy must start with the value.",
			},
			attr.Name + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "The name of the DLP policy must end with the value.",
			},
			attr.DLPPolicies: schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "Return a matching DLP Policy by its ID. The ID for the DLP Policy can be obtained from the Admin API or the URL string in the Admin Console.",
						},
						attr.Name: schema.StringAttribute{
							Computed:    true,
							Description: "Return a DLP Policy that exactly matches this name.",
						},
					},
				},
			},
		},
	}
}

func (d *dlpPolicies) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dlpPoliciesModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name, filter := getNameFilter(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix)

	if countOptionalAttributes(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrDLPPoliciesDatasourceShouldSetOneOptionalNameAttribute, TwingateDLPPolicies)

		return
	}

	policies, err := d.client.ReadDLPPolicies(ctx, name, filter)
	if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
		addErr(&resp.Diagnostics, err, TwingateDLPPolicy)

		return
	}

	data.ID = types.StringValue("dlp-policies-all")
	data.DLPPolicies = convertDLPPoliciesToTerraform(policies)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func convertDLPPoliciesToTerraform(policies []*model.DLPPolicy) []dlpPolicyModel {
	result := make([]dlpPolicyModel, 0, len(policies))

	for _, policy := range policies {
		result = append(result, dlpPolicyModel{
			ID:   types.StringValue(policy.ID),
			Name: types.StringValue(policy.Name),
		})
	}

	return result
}