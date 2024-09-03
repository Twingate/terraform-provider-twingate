package datasource

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &dlpPolicies{}

func NewDLPPoliciesDatasource() datasource.DataSource {
	return &dlpPolicies{}
}

var invalidNameRegex = regexp.MustCompile(`\W+`)

type dlpPolicies struct {
	client *client.Client
}

type dlpPoliciesModel struct {
	ID           types.String     `tfsdk:"id"`
	Name         types.String     `tfsdk:"name"`
	NameRegexp   types.String     `tfsdk:"name_regexp"`
	NameContains types.String     `tfsdk:"name_contains"`
	NameExclude  types.String     `tfsdk:"name_exclude"`
	NamePrefix   types.String     `tfsdk:"name_prefix"`
	NameSuffix   types.String     `tfsdk:"name_suffix"`
	Policies     []dlpPolicyModel `tfsdk:"dlp_policies"`
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
		Description: "DLP policies are currently in early access. For more information, reach out to your account manager.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this data source.",
			},
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only DLP policies that exactly match this name. If no options are passed, returns all DLP policies.",
			},
			attr.Name + attr.FilterByContains: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only DLP policies that contain this string.",
			},
			attr.Name + attr.FilterByExclude: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only DLP policies that do not include this string.",
			},
			attr.Name + attr.FilterByPrefix: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only DLP policies that start in this string.",
			},
			attr.Name + attr.FilterByRegexp: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only DLP policies that satisfy this regex.",
			},
			attr.Name + attr.FilterBySuffix: schema.StringAttribute{
				Optional:    true,
				Description: "Returns only DLP policies that end in this string.",
			},

			attr.DLPPolicies: schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of DLP policies",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						attr.ID: schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the DLP policy",
						},
						attr.Name: schema.StringAttribute{
							Computed:    true,
							Description: "The name of the DLP policy",
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

	if countOptionalAttributes(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix) > 1 {
		addErr(&resp.Diagnostics, ErrGroupsDatasourceShouldSetOneOptionalNameAttribute, TwingateDLPPolicies)

		return
	}

	name, filter := getNameFilter(data.Name, data.NameRegexp, data.NameContains, data.NameExclude, data.NamePrefix, data.NameSuffix)
	policies, err := d.client.ReadDLPPolicies(client.WithCallerCtx(ctx, datasourceKey), name, filter)

	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateDLPPolicy)

		return
	}

	data.ID = types.StringValue("policies-by-name-" + sanitizeName(name))
	data.Policies = convertPoliciesToTerraform(policies)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
