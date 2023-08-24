package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
)

type GenerateConnectorTokens struct {
	ConnectorTokensResponse `graphql:"connectorGenerateTokens(connectorId: $connectorId)"`
}

type ConnectorTokensResponse struct {
	ConnectorTokens gqlConnectorTokens
	OkError
}

type gqlConnectorTokens struct {
	AccessToken  string
	RefreshToken string
}

func (q gqlConnectorTokens) ToModel() *model.ConnectorTokens {
	return &model.ConnectorTokens{
		AccessToken:  q.AccessToken,
		RefreshToken: q.RefreshToken,
	}
}

func (q GenerateConnectorTokens) ToModel() *model.ConnectorTokens {
	return q.ConnectorTokens.ToModel()
}

func (q GenerateConnectorTokens) IsEmpty() bool {
	return q.ConnectorTokens.RefreshToken == "" && q.ConnectorTokens.AccessToken == ""
}
