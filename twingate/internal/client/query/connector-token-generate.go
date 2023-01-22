package query

import (
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type GenerateConnectorTokens struct {
	ConnectorTokensResponse `graphql:"connectorGenerateTokens(connectorId: $connectorId)"`
}

type ConnectorTokensResponse struct {
	ConnectorTokens gqlConnectorTokens
	OkError
}

type gqlConnectorTokens struct {
	AccessToken  graphql.String
	RefreshToken graphql.String
}

func (q gqlConnectorTokens) ToModel() *model.ConnectorTokens {
	return &model.ConnectorTokens{
		AccessToken:  string(q.AccessToken),
		RefreshToken: string(q.RefreshToken),
	}
}

func (q GenerateConnectorTokens) ToModel() *model.ConnectorTokens {
	return q.ConnectorTokens.ToModel()
}
