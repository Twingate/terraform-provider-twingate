package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &dlpPolicy{}

func NewDLPPolicyDatasource() datasource.DataSource {
	return &dlpPolicy{}
}

type dlpPolicy struct {
	client *client.Client
}

type dlpPolicyModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *dlpPolicy) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateDLPPolicy
}

func (d *dlpPolicy) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dlpPolicy) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "DLP policies are currently in early access. For more information, reach out to your account manager.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The DLP policy's ID. Returns a DLP policy that has this ID.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Root(attr.ID).Expression(), path.Root(attr.Name).Expression()),
				},
			},
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The DLP policy's name. Returns a DLP policy that exactly matches this name.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Root(attr.ID).Expression(), path.Root(attr.Name).Expression()),
				},
			},
		},
	}
}

func (d *dlpPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dlpPolicyModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.ReadDLPPolicy(client.WithCallerCtx(ctx, datasourceKey), &model.DLPPolicy{
		ID:   data.ID.ValueString(),
		Name: data.Name.ValueString(),
	})
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateDLPPolicy)

		return
	}

	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
