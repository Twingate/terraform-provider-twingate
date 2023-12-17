package datasource

import (
    "context"
    "errors"
    "fmt"
    "github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
    "github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
    "github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var ErrResourcesDatasourceNameAttributes = errors.New("Only one of name, name_regex, name_contains, name_exclude, name_prefix or name_suffix must be set.")

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &resources{}

func NewResourcesDatasource() datasource.DataSource {
    return &resources{}
}

type resources struct {
    client *client.Client
}

type resourcesModel struct {
    ID           types.String    `tfsdk:"id"`
    Name         types.String    `tfsdk:"name"`
    NameRegexp   types.String    `tfsdk:"name_regexp"`
    NameContains types.String    `tfsdk:"name_contains"`
    NameExclude  types.String    `tfsdk:"name_exclude"`
    NamePrefix   types.String    `tfsdk:"name_prefix"`
    NameSuffix   types.String    `tfsdk:"name_suffix"`
    Resources    []resourceModel `tfsdk:"resources"`
}

func (d *resources) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = TwingateResources
}

func (d *resources) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func protocolSchema() schema.SingleNestedAttribute {
    return schema.SingleNestedAttribute{
        Computed: true,
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
}

func (d *resources) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Resources in Twingate represent servers on the private network that clients can connect to. Resources can be defined by IP, CIDR range, FQDN, or DNS zone. For more information, see the Twingate [documentation](https://docs.twingate.com/docs/resources-and-access-nodes).",
        Attributes: map[string]schema.Attribute{
            attr.ID: schema.StringAttribute{
                Computed:    true,
                Description: computedDatasourceIDDescription,
            },
            attr.Name: schema.StringAttribute{
                Optional:    true,
                Description: "Returns only resources that exactly match this name.",
            },
            attr.Name + attr.FilterByRegexp: schema.StringAttribute{
                Optional:    true,
                Description: "The regular expression match of the name of the resource.",
            },
            attr.Name + attr.FilterByContains: schema.StringAttribute{
                Optional:    true,
                Description: "Match when the value exist in the name of the resource.",
            },
            attr.Name + attr.FilterByExclude: schema.StringAttribute{
                Optional:    true,
                Description: "Match when the value does not exist in the name of the resource.",
            },
            attr.Name + attr.FilterByPrefix: schema.StringAttribute{
                Optional:    true,
                Description: "The name of the resource must start with the value.",
            },
            attr.Name + attr.FilterBySuffix: schema.StringAttribute{
                Optional:    true,
                Description: "The name of the resource must end with the value.",
            },
            // computed
            attr.Resources: schema.ListNestedAttribute{
                Computed:    true,
                Description: "List of Resources",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        attr.ID: schema.StringAttribute{
                            Computed:    true,
                            Description: "The id of the Resource",
                        },
                        attr.Name: schema.StringAttribute{
                            Computed:    true,
                            Description: "The name of the Resource",
                        },
                        attr.Address: schema.StringAttribute{
                            Computed:    true,
                            Description: "The Resource's IP/CIDR or FQDN/DNS zone",
                        },
                        attr.RemoteNetworkID: schema.StringAttribute{
                            Computed:    true,
                            Description: "Remote Network ID where the Resource lives",
                        },
                        attr.Protocols: schema.SingleNestedAttribute{
                            Description: "Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no restriction, and all protocols and ports are allowed.",
                            Computed:    true,
                            Attributes: map[string]schema.Attribute{
                                attr.AllowIcmp: schema.BoolAttribute{
                                    Computed:    true,
                                    Description: "Whether to allow ICMP (ping) traffic",
                                },
                                attr.TCP: protocolSchema(),
                                attr.UDP: protocolSchema(),
                            },
                        },
                    },
                },
            },
        },
    }
}

func (d *resources) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data resourcesModel

    // Read Terraform configuration data into the model
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    if resp.Diagnostics.HasError() {
        return
    }

    resources, err := d.client.ReadResourcesByName(ctx, data.Name.ValueString())
    if err != nil && !errors.Is(err, client.ErrGraphqlResultIsEmpty) {
        addErr(&resp.Diagnostics, err, TwingateResources)

        return
    }

    data.ID = types.StringValue("query resources by name: " + data.Name.ValueString())
    data.Resources = convertResourcesToTerraform(resources)

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
