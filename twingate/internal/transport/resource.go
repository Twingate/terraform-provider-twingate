package transport

import (
	"context"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

const (
	resourceResourceName        = "resource"
	readResourceQueryGroupsSize = 50
)

type gqlResource struct {
	IDName
	Address struct {
		Value graphql.String
	}
	RemoteNetwork struct {
		ID graphql.ID
	}
	Protocols *Protocols
}

type Protocols struct {
	UDP       *Protocol       `json:"udp"`
	TCP       *Protocol       `json:"tcp"`
	AllowIcmp graphql.Boolean `json:"allowIcmp"`
}

type Protocol struct {
	Ports  []*PortRange   `json:"ports"`
	Policy graphql.String `json:"policy"`
}

type PortRange struct {
	Start graphql.Int `json:"start"`
	End   graphql.Int `json:"end"`
}

type Resource struct {
	ID              graphql.ID
	RemoteNetworkID graphql.ID
	Address         graphql.String
	Name            graphql.String
	GroupsIds       []*graphql.ID
	Protocols       *Protocols
	IsActive        graphql.Boolean
}

type createResourceQuery struct {
	ResourceCreate struct {
		OkError
		Entity struct {
			ID graphql.ID
		}
	} `graphql:"resourceCreate(name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

type ProtocolsInput struct {
	UDP       *ProtocolInput  `json:"udp"`
	TCP       *ProtocolInput  `json:"tcp"`
	AllowIcmp graphql.Boolean `json:"allowIcmp"`
}

type ProtocolInput struct {
	Ports  []*PortRangeInput `json:"ports"`
	Policy graphql.String    `json:"policy"`
}

type PortRangeInput struct {
	Start graphql.Int `json:"start"`
	End   graphql.Int `json:"end"`
}

func (client *Client) CreateResource(ctx context.Context, resource *model.Resource) (string, error) {
	protocols := &ProtocolsInput{
		AllowIcmp: true,
		TCP: &ProtocolInput{
			Policy: model.PolicyRestricted,
			Ports: []*PortRangeInput{
				{Start: 80, End: 83},
				{Start: 85, End: 85},
			},
		},
		UDP: &ProtocolInput{
			Policy: model.PolicyAllowAll,
		},
	}

	variables := newVars(
		gqlID(resource.RemoteNetworkID, "remoteNetworkId"),
		gqlIDs(resource.Groups, "groupIds"),
		gqlField(resource.Name, "name"),
		gqlField(resource.Address, "address"),
		//gqlField(newProtocolsInput(resource.Protocols), "protocols"),
	)
	variables["protocols"] = protocols
	//variables["protocols"] = newProtocolsInput(resource.Protocols)

	response := createResourceQuery{}
	err := client.GraphqlClient.NamedMutate(ctx, "createResource", &response, variables)

	if err != nil {
		return "", NewAPIError(err, "create", resourceResourceName)
	}

	if !response.ResourceCreate.Ok {
		return "", NewAPIError(NewMutationError(response.ResourceCreate.Error), "create", resourceResourceName)
	}

	return idToString(response.ResourceCreate.Entity.ID), nil
}

type readResourceQuery struct {
	Resource *struct {
		//IDName
		//Address struct {
		//	//Type  graphql.String
		//	Value graphql.String
		//}
		//RemoteNetwork struct {
		//	ID graphql.ID
		//}
		//Protocols *Protocols
		gqlResource
		Groups struct {
			PageInfo struct {
				HasNextPage graphql.Boolean
			}
			Edges []*Edges
		} `graphql:"groups(first: $first)"`
		IsActive graphql.Boolean
	} `graphql:"resource(id: $id)"`
}

func (client *Client) ReadResource(ctx context.Context, resourceID string) (*model.Resource, error) {
	if resourceID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", resourceResourceName)
	}

	response := readResourceQuery{}
	variables := newVars(
		gqlID(resourceID),
		gqlField(readResourceQueryGroupsSize, "first"),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, resourceID)
	}

	if response.Resource.Groups.PageInfo.HasNextPage {
		return nil, NewAPIErrorWithID(ErrTooManyGroupsError, "read", resourceResourceName, resourceID)
	}

	return response.ToModel(), nil
}

type readResourcesQuery struct { //nolint
	Resources struct {
		Edges []*Edges
	}
}

func (client *Client) ReadResources(ctx context.Context) ([]*model.Resource, error) { //nolint
	response := readResourcesQuery{}

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, nil)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	return response.ToModel(), nil
}

type updateResourceQuery struct {
	ResourceUpdate *OkError `graphql:"resourceUpdate(id: $id, name: $name, address: $address, remoteNetworkId: $remoteNetworkId, groupIds: $groupIds, protocols: $protocols)"`
}

func (client *Client) UpdateResource(ctx context.Context, resource *model.Resource) error {
	variables := newVars(
		gqlID(resource.ID),
		gqlID(resource.RemoteNetworkID, "remoteNetworkId"),
		gqlIDs(resource.Groups, "groupIds"),
		gqlField(resource.Name, "name"),
		gqlField(resource.Address, "address"),
		gqlField(newProtocolsInput(resource.Protocols), "protocols"),
	)

	response := updateResourceQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)

	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	if !response.ResourceUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ResourceUpdate.Error), "update", resourceResourceName, resource.ID)
	}

	return nil
}

type deleteResourceQuery struct {
	ResourceDelete *OkError `graphql:"resourceDelete(id: $id)"`
}

func (client *Client) DeleteResource(ctx context.Context, resourceID string) error {
	if resourceID == "" {
		return NewAPIError(ErrGraphqlIDIsEmpty, "delete", resourceResourceName)
	}

	response := deleteResourceQuery{}

	variables := newVars(gqlID(resourceID))

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)
	if err != nil {
		return NewAPIErrorWithID(err, "delete", resourceResourceName, resourceID)
	}

	if !response.ResourceDelete.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ResourceDelete.Error), "delete", resourceResourceName, resourceID)
	}

	return nil
}

type readResourceWithoutGroupsQuery struct {
	Resource *gqlResource `graphql:"resource(id: $id)"`
	//Resource *struct {
	//	IDName
	//	Address struct {
	//		Value graphql.String
	//	}
	//	RemoteNetwork struct {
	//		ID graphql.ID
	//	}
	//	Protocols *Protocols
	//} `graphql:"resource(id: $id)"`
}

func (client *Client) ReadResourceWithoutGroups(ctx context.Context, resourceID string) (*model.Resource, error) {
	if resourceID == "" {
		return nil, NewAPIError(ErrGraphqlIDIsEmpty, "read", resourceResourceName)
	}

	response := readResourceWithoutGroupsQuery{}
	variables := newVars(gqlID(resourceID))

	err := client.GraphqlClient.NamedQuery(ctx, "readResource", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	if response.Resource == nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, resourceID)
	}

	return response.Resource.ToModel(), nil
}

type updateResourceActiveStateQuery struct {
	ResourceUpdate *OkError `graphql:"resourceUpdate(id: $id, isActive: $isActive)"`
}

func (client *Client) UpdateResourceActiveState(ctx context.Context, resource *model.Resource) error {
	variables := map[string]interface{}{
		"id":       resource.ID,
		"isActive": resource.IsActive,
	}

	response := updateResourceActiveStateQuery{}

	err := client.GraphqlClient.NamedMutate(ctx, "updateResource", &response, variables)

	if err != nil {
		return NewAPIErrorWithID(err, "update", resourceResourceName, resource.ID)
	}

	if !response.ResourceUpdate.Ok {
		return NewAPIErrorWithID(NewMutationError(response.ResourceUpdate.Error), "update", resourceResourceName, resource.ID)
	}

	return nil
}

type readResourcesByNameQuery struct {
	Resources struct {
		Edges []*struct {
			Node *gqlResource
			//Node *struct {
			//IDName
			//Address struct {
			//	Value graphql.String
			//}
			//RemoteNetwork struct {
			//	ID graphql.ID
			//}
			//Protocols *Protocols
			//}
		}
	} `graphql:"resources(filter: {name: {eq: $name}})"`
}

func (client *Client) ReadResourcesByName(ctx context.Context, name string) ([]*model.Resource, error) {
	response := readResourcesByNameQuery{}
	variables := newVars(
		gqlField(name, "name"),
	)

	err := client.GraphqlClient.NamedQuery(ctx, "readResources", &response, variables)
	if err != nil {
		return nil, NewAPIErrorWithID(err, "read", resourceResourceName, "All")
	}

	if len(response.Resources.Edges) == 0 {
		return nil, NewAPIErrorWithID(ErrGraphqlResultIsEmpty, "read", resourceResourceName, "All")
	}

	return response.ToModel(), nil
}
