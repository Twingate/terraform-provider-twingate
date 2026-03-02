package query

import (
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/hasura/go-graphql-client"
)

type ReadGateway struct {
	Gateway *gqlGateway `graphql:"gateway(id: $id)"`
}

func (q ReadGateway) IsEmpty() bool {
	return q.Gateway == nil
}

func (q ReadGateway) ToModel() *model.Gateway {
	if q.Gateway == nil {
		return nil
	}

	return q.Gateway.ToModel()
}

type gqlGateway struct {
	ID            graphql.ID
	Address       string
	RemoteNetwork struct {
		ID graphql.ID
	}
	X509CA struct {
		ID graphql.ID
	} `graphql:"x509CA"`
	SSHCA *struct {
		ID graphql.ID
	} `graphql:"sshCA"`
}

func (g gqlGateway) ToModel() *model.Gateway {
	gateway := &model.Gateway{
		ID:              string(g.ID),
		Address:         g.Address,
		RemoteNetworkID: string(g.RemoteNetwork.ID),
		X509CAID:        string(g.X509CA.ID),
	}

	if g.SSHCA != nil {
		gateway.SSHCAID = string(g.SSHCA.ID)
	}

	return gateway
}
