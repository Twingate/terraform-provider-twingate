package resource

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var _ ephemeral.EphemeralResource = &ephemeralConnectorTokens{}

func NewEphemeralConnectorTokens() ephemeral.EphemeralResourceWithConfigure {
	return &ephemeralConnectorTokens{}
}

type ephemeralConnectorTokens struct {
	client *client.Client
}

type ephemeralConnectorTokensModel struct {
	ConnectorID  types.String `tfsdk:"connector_id"`
	AccessToken  types.String `tfsdk:"access_token"`
	RefreshToken types.String `tfsdk:"refresh_token"`
}

func (r *ephemeralConnectorTokens) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = TwingateConnectorTokens
}

func (r *ephemeralConnectorTokens) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ephemeralConnectorTokens) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "This resource type will generate tokens for a Connector, which are needed to successfully provision one on your network. The Connector itself has its own resource type and must be created before you can provision tokens.",
		MarkdownDescription: "This resource type will generate tokens for a Connector, which are needed to successfully provision one on your network. The Connector itself has its own resource type and must be created before you can provision tokens.\n\n~> **Warning:** When existing connectors are converted to ephemeral mode, Terraform generates a new token during plan or apply, preventing the connectors from reconnecting until they are updated with the new token.\nRather than converting existing connectors, we recommend creating new connectors with ephemeral resource tokens and deleting the old ones after migration.",
		Attributes: map[string]schema.Attribute{
			attr.ConnectorID: schema.StringAttribute{
				Required:    true,
				Description: "The ID of the parent Connector",
			},
			// computed
			attr.AccessToken: schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The Access Token of the parent Connector",
			},
			attr.RefreshToken: schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The Refresh Token of the parent Connector",
			},
		},
	}
}

func (r *ephemeralConnectorTokens) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client not configured",
			"The provider client is nil. Please report this issue to the provider developers.",
		)

		return
	}

	// Read configuration from the request
	var plan ephemeralConnectorTokensModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tokens, err := r.client.GenerateConnectorTokens(ctx, plan.ConnectorID.ValueString())
	if err != nil {
		addErr(&resp.Diagnostics, err, operationCreate, TwingateConnectorTokens)

		return
	}

	plan.AccessToken = types.StringValue(tokens.AccessToken)
	plan.RefreshToken = types.StringValue(tokens.RefreshToken)

	resp.Diagnostics.Append(resp.Result.Set(ctx, plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
