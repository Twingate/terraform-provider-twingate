package datasource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
	tfattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &resource{}

func NewResourceDatasource() datasource.DataSource {
	return &resource{}
}

type resource struct {
	client *client.Client
}

type resourceModel struct {
	ID              types.String    `tfsdk:"id"`
	Name            types.String    `tfsdk:"name"`
	Address         types.String    `tfsdk:"address"`
	RemoteNetworkID types.String    `tfsdk:"remote_network_id"`
	Tags            types.Map       `tfsdk:"tags"`
	Protocols       *protocolsModel `tfsdk:"protocols"`
}

type protocolsModel struct {
	AllowIcmp types.Bool     `tfsdk:"allow_icmp"`
	TCP       *protocolModel `tfsdk:"tcp"`
	UDP       *protocolModel `tfsdk:"udp"`
}

type protocolModel struct {
	Policy types.String   `tfsdk:"policy"`
	Ports  []types.String `tfsdk:"ports"`
}

func (d *resource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = TwingateResource
}

func (d *resource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *resource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	protocolSchema := schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			attr.Policy: schema.StringAttribute{
				Computed:    true,
				Description: fmt.Sprintf("Whether to allow or deny all ports, or restrict protocol access within certain port ranges: Can be `%s` (only listed ports are allowed), `%s`, or `%s`", model.PolicyRestricted, model.PolicyAllowAll, model.PolicyDenyAll),
			},
			attr.Ports: schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of port ranges between 1 and 65535 inclusive, in the format `100-200` for a range, or `8080` for a single port",
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Resources in Twingate represent any network destination address that you wish to provide private access to for users authorized via the Twingate Client application. Resources can be defined by either IP or DNS address, and all private DNS addresses will be automatically resolved with no client configuration changes. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
		Attributes: map[string]schema.Attribute{
			attr.ID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Resource. The ID for the Resource can be obtained from the Admin API or the URL string in the Admin Console.",
			},
			// computed
			attr.Name: schema.StringAttribute{
				Computed:    true,
				Description: "The name of the Resource",
			},
			attr.Address: schema.StringAttribute{
				Computed:    true,
				Description: "The Resource's address, which may be an IP address, CIDR range, or DNS address",
			},
			attr.RemoteNetworkID: schema.StringAttribute{
				Computed:    true,
				Description: "The Remote Network ID that the Resource is associated with. Resources may only be associated with a single Remote Network.",
			},
			attr.Tags: schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The `tags` attribute consists of a key-value pairs that correspond with tags to be set on the resource.",
			},
		},
		Blocks: map[string]schema.Block{
			attr.Protocols: schema.SingleNestedBlock{
				Description: "By default (when this argument is not defined) no restriction is applied, and all protocols and ports are allowed.",
				Attributes: map[string]schema.Attribute{
					attr.AllowIcmp: schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to allow ICMP (ping) traffic",
					},
				},
				Blocks: map[string]schema.Block{
					attr.TCP: &protocolSchema,
					attr.UDP: &protocolSchema,
				},
			},
		},
	}
}

func (d *resource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data resourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := d.client.ReadResource(client.WithCallerCtx(ctx, datasourceKey), data.ID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, TwingateResource)

		return
	}

	data.Name = types.StringValue(resource.Name)
	data.Address = types.StringValue(resource.Address)
	data.RemoteNetworkID = types.StringValue(resource.RemoteNetworkID)
	data.Protocols = convertProtocolsToTerraform(resource.Protocols)
	tags, diags := convertTagsToTerraform(resource.Tags)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Tags = tags

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func convertProtocolsToTerraform(protocols *model.Protocols) *protocolsModel {
	if protocols == nil {
		return nil
	}

	return &protocolsModel{
		AllowIcmp: types.BoolValue(protocols.AllowIcmp),
		TCP:       convertProtocolToTerraform(protocols.TCP),
		UDP:       convertProtocolToTerraform(protocols.UDP),
	}
}

func convertProtocolToTerraform(protocol *model.Protocol) *protocolModel {
	return &protocolModel{
		Policy: types.StringValue(protocol.Policy),
		Ports: utils.Map(protocol.Ports, func(port *model.PortRange) types.String {
			return types.StringValue(port.String())
		}),
	}
}

func convertTagsToTerraform(tags map[string]string) (types.Map, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if len(tags) == 0 {
		return types.MapNull(types.StringType), diagnostics
	}

	elements := make(map[string]tfattr.Value, len(tags))
	for key, val := range tags {
		elements[key] = types.StringValue(val)
	}

	return types.MapValue(types.StringType, elements)
}
