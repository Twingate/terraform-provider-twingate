package query

import (
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/utils"
)

const CursorConnectors = "connectorsEndCursor"

type ReadConnectors struct {
	Connectors `graphql:"connectors(filter: $filter, after: $connectorsEndCursor, first: $pageLimit)"`
}

type Connectors struct {
	PaginatedResource[*ConnectorEdge]
}

type ConnectorEdge struct {
	Node *gqlConnector
}

func (q ReadConnectors) IsEmpty() bool {
	return len(q.Edges) == 0
}

func (q ReadConnectors) ToModel() []*model.Connector {
	if len(q.Edges) == 0 {
		return nil
	}

	return q.Connectors.ToModel()
}

func (c Connectors) ToModel() []*model.Connector {
	return utils.Map[*ConnectorEdge, *model.Connector](c.Edges, func(edge *ConnectorEdge) *model.Connector {
		return edge.Node.ToModel()
	})
}

type ConnectorFilterInput struct {
	Name *StringFilterOperationInput `json:"name"`
}

func NewConnectorFilterInput(name, filter string) *ConnectorFilterInput {
	return &ConnectorFilterInput{
		Name: NewStringFilterOperationInput(name, filter),
	}
}
