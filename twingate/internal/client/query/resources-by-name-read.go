package query

import "github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"

type ReadResourcesByName struct {
	Resources `graphql:"resources(filter: $filter, after: $resourcesEndCursor, first: $pageLimit)"`
}

func (q ReadResourcesByName) IsEmpty() bool {
	return len(q.Edges) == 0
}

type ResourceFilterInput struct {
	Name            *StringFilterOperationInput          `json:"name"`
	Tags            *TagsFilterOperatorInput             `json:"tags"`
	RemoteNetworkID *RemoteNetworkIdFilterOperationInput `json:"remoteNetworkId"`
}

func NewResourceFilterInput(input *model.ResourcesFilter) *ResourceFilterInput {
	if input == nil {
		return nil
	}

	name := NewStringFilterOperationInput(input.GetName(), input.NameFilter)

	if len(input.NameIn) > 0 {
		name = NewStringFilterInOperationInput(input.NameIn)
	}

	return &ResourceFilterInput{
		Name:            name,
		Tags:            NewTagsFilterOperatorInput(input.Tags),
		RemoteNetworkID: NewRemoteNetworkIdFilterOperationInput(input.RemoteNetworkID),
	}
}

func NewTagsFilterOperatorInput(tags map[string]string) *TagsFilterOperatorInput {
	if len(tags) == 0 {
		return nil
	}

	filter := &TagsFilterOperatorInput{
		And: make([]TagKeyValueFilterInput, 0, len(tags)),
	}

	for key, value := range tags {
		filter.And = append(filter.And, TagKeyValueFilterInput{
			Key: key,
			Value: TagValueFilterInput{
				Eq: &value,
			},
		})
	}

	return filter
}

func NewRemoteNetworkIdFilterOperationInput(remoteNetworkId *string) *RemoteNetworkIdFilterOperationInput {
	if remoteNetworkId == nil {
		return nil
	}

	return &RemoteNetworkIdFilterOperationInput{
		Eq: remoteNetworkId,
	}
}

type TagsFilterOperatorInput struct {
	And []TagKeyValueFilterInput `json:"and"`
}

type TagKeyValueFilterInput struct {
	Key   string              `json:"key"`
	Value TagValueFilterInput `json:"value"`
}

type TagValueFilterInput struct {
	Eq *string `json:"eq"`
}

type RemoteNetworkIdFilterOperationInput struct {
	Eq *string  `json:"eq"`
	In []string `json:"in"`
}
