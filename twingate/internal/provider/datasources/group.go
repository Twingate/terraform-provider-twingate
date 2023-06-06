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
var _ datasource.DataSource = &group{}

func NewGroupDatasource() datasource.DataSource {
	return &group{}
}

type group struct {
	client *client.Client
}

type groupModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	SecurityPolicyID types.String `tfsdk:"security_policy_id"`
	IsActive         types.Bool   `tfsdk:"is_active"`
}

func (d *group) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateGroup
}

func (d *group) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *group) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Group. The ID for the Group can be obtained from the Admin API or the URL string in the Admin Console.",
			},
			// computed
			attr.Name: schema.StringAttribute{
				Computed:    true,
				Description: "The name of the Group",
			},
			attr.IsActive: schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates if the Group is active",
			},
			attr.Type: schema.StringAttribute{
				Computed:    true,
				Description: "The type of the Group",
			},
			attr.SecurityPolicyID: schema.StringAttribute{
				Computed:    true,
				Description: "The Security Policy assigned to the Group.",
			},
		},
	}
}

func (d *group) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data groupModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := d.client.ReadGroup(ctx, data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationRead, TwingateGroup)
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Type = types.StringValue(group.Type)
	data.IsActive = types.BoolValue(group.IsActive)
	data.SecurityPolicyID = types.StringValue(group.SecurityPolicyID)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
