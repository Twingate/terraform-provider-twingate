package datasources

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &securityPolicy{}

func NewSecurityPolicyDatasource() datasource.DataSource {
	return &securityPolicy{}
}

type securityPolicy struct {
	client *client.Client
}

type securityPolicyModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *securityPolicy) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateSecurityPolicy
}

func (d *securityPolicy) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *securityPolicy) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Security Policies are defined in the Twingate Admin Console and determine user and device authentication requirements for Resources.",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Optional:    true,
				Description: "Return a Security Policy by its ID. The ID for the Security Policy can be obtained from the Admin API or the URL string in the Admin Console.",
			},
			attr.Name: schema.StringAttribute{
				Optional:    true,
				Description: "Return a Security Policy that exactly matches this name.",
			},
		},
	}
}

func (d *securityPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data securityPolicyModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() && !data.Name.IsNull() {
		addErr(&resp.Diagnostics, ErrArgumentsInvalidCombination, operationRead, TwingateSecurityPolicy)
		return
	}

	policy, err := d.client.ReadSecurityPolicy(ctx, data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, TwingateSecurityPolicy)
		return
	}

	data.ID = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
