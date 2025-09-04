package client

import (
	"context"
	"reflect"
	"testing"

	attrs "github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

var skipCache = CacheOptions{}

type mockClient struct {
}

func (m mockClient) ReadFullResourcesByName(ctx context.Context, filter *model.ResourcesFilter) ([]*model.Resource, error) {
	return []*model.Resource{}, nil
}

func (m mockClient) ReadFullGroupsByName(ctx context.Context, filter *model.GroupsFilter) ([]*model.Group, error) {
	return []*model.Group{}, nil
}

func (m mockClient) ReadFullResources(ctx context.Context) ([]*model.Resource, error) {
	return []*model.Resource{}, nil
}

func (m mockClient) ReadFullGroups(ctx context.Context) ([]*model.Group, error) {
	return []*model.Group{}, nil
}

func TestClientCache_SetClient(t *testing.T) {
	cache := &clientCache{}
	cache.setClient(&mockClient{}, skipCache)

	assert.NotNil(t, cache.handlers)
	assert.Contains(t, cache.handlers, reflect.TypeOf(&model.Resource{}).String())
	assert.Contains(t, cache.handlers, reflect.TypeOf(&model.Group{}).String())
}

func TestHandler_SetAndGetResource(t *testing.T) {
	mockResource := &model.Resource{ID: "resource1"}
	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return []*model.Resource{
				mockResource,
			}, nil
		},
	}

	handler.setResource(mockResource)

	result, exists := handler.getResource("resource1")
	assert.True(t, exists)

	retrievedResource, ok := result.(*model.Resource)
	assert.True(t, ok)
	assert.Equal(t, "resource1", retrievedResource.ID)
}

func TestHandler_InvalidateResource(t *testing.T) {
	mockResource := &model.Resource{ID: "resource2"}
	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return []*model.Resource{
				mockResource,
			}, nil
		},
	}

	handler.setResource(mockResource)
	_, existsBefore := handler.getResource("resource2")
	assert.True(t, existsBefore)

	handler.invalidateResource("resource2")
	_, existsAfter := handler.getResource("resource2")
	assert.False(t, existsAfter)
}

func TestHandler_Init(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "resource1"},
		{ID: "resource2"},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	_, exists1 := handler.getResource("resource1")
	_, exists2 := handler.getResource("resource2")
	assert.True(t, exists1)
	assert.True(t, exists2)
}

func TestHandler_MatchGroupsByName(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test", IsActive: true, Type: model.GroupTypeSystem},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.GroupsFilter{
		Name: optionalString("test"),
	})

	assert.Len(t, matched, 1)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.Equal(t, "group2", group.ID)
}

func TestHandler_MatchGroupsByNameAndIsActive(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test", IsActive: true, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test", IsActive: false, Type: model.GroupTypeSystem},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	isActive := false

	matched := handler.matchResources(&model.GroupsFilter{
		Name:     optionalString("test"),
		IsActive: &isActive,
	})

	assert.Len(t, matched, 1)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.Equal(t, "group3", group.ID)
}

func TestHandler_MatchGroupsByNameAndType(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test", IsActive: true, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test", IsActive: false, Type: model.GroupTypeSystem},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.GroupsFilter{
		Name:  optionalString("test"),
		Types: []string{model.GroupTypeSystem},
	})

	assert.Len(t, matched, 2)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.True(t, group.ID == "group2" || group.ID == "group3")
}

func TestHandler_MatchGroupsByNameAndTypeAndIsActive(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test", IsActive: true, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test", IsActive: false, Type: model.GroupTypeSystem},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	isActive := true
	matched := handler.matchResources(&model.GroupsFilter{
		Name:     optionalString("test"),
		Types:    []string{model.GroupTypeSystem},
		IsActive: &isActive,
	})

	assert.Len(t, matched, 1)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.Equal(t, "group2", group.ID)
}

func TestHandler_MatchGroupsByTypeIsActiveNamePrefix(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test_a", IsActive: false, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test_b", IsActive: true, Type: model.GroupTypeSystem},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	isActive := true
	matched := handler.matchResources(&model.GroupsFilter{
		Name:       optionalString("test"),
		NameFilter: attrs.FilterByPrefix,
		Types:      []string{model.GroupTypeSystem},
		IsActive:   &isActive,
	})

	assert.Len(t, matched, 1)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.Equal(t, "group3", group.ID)
}

func TestHandler_MatchGroupsByTypeIsActiveNameSuffix(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod_ok", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test_ok", IsActive: false, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test_b", IsActive: true, Type: model.GroupTypeSystem},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	isActive := true
	matched := handler.matchResources(&model.GroupsFilter{
		Name:       optionalString("ok"),
		NameFilter: attrs.FilterBySuffix,
		Types:      []string{model.GroupTypeSystem, model.GroupTypeManual},
		IsActive:   &isActive,
	})

	assert.Len(t, matched, 1)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.Equal(t, "group1", group.ID)
}

func TestHandler_MatchGroupsByTypeIsActiveNameContains(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod_new_ok", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test_new_ok", IsActive: false, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test_b", IsActive: true, Type: model.GroupTypeSystem},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	isActive := false
	matched := handler.matchResources(&model.GroupsFilter{
		Name:       optionalString("_new_"),
		NameFilter: attrs.FilterByContains,
		IsActive:   &isActive,
	})

	assert.Len(t, matched, 1)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.Equal(t, "group2", group.ID)
}

func TestHandler_MatchGroupsByTypeNameExclude(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod_new_ok", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test_new_ok", IsActive: false, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test_b", IsActive: true, Type: model.GroupTypeSystem},
		{ID: "group4", Name: "test_c", IsActive: true, Type: model.GroupTypeManual},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.GroupsFilter{
		Name:       optionalString("_new_"),
		NameFilter: attrs.FilterByExclude,
		Types:      []string{model.GroupTypeManual},
	})

	assert.Len(t, matched, 1)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.Equal(t, "group4", group.ID)
}

func TestHandler_MatchGroupsByTypeNameRegexp(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod_new_ok", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test_new_ok", IsActive: false, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test_b", IsActive: true, Type: model.GroupTypeSystem},
		{ID: "group4", Name: "test_c", IsActive: true, Type: model.GroupTypeManual},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.GroupsFilter{
		Name:       optionalString("test_*"),
		NameFilter: attrs.FilterByRegexp,
		Types:      []string{model.GroupTypeSystem},
	})

	assert.Len(t, matched, 2)

	group, ok := matched[0].(*model.Group)
	assert.True(t, ok)
	assert.True(t, group.ID == "group2" || group.ID == "group3")
}

func TestHandler_MatchGroupsByTypeNameInvalidRegexp(t *testing.T) {
	mockResources := []*model.Group{
		{ID: "group1", Name: "prod_new_ok", IsActive: true, Type: model.GroupTypeManual},
		{ID: "group2", Name: "test_new_ok", IsActive: false, Type: model.GroupTypeSystem},
		{ID: "group3", Name: "test_b", IsActive: true, Type: model.GroupTypeSystem},
		{ID: "group4", Name: "test_c", IsActive: true, Type: model.GroupTypeManual},
	}

	handler := &handler[*model.Group, *model.GroupsFilter]{
		readResources: func(ctx context.Context) ([]*model.Group, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.GroupsFilter{
		Name:       optionalString("test_{*}"),
		NameFilter: attrs.FilterByRegexp,
		Types:      []string{model.GroupTypeSystem},
	})

	assert.Len(t, matched, 0)
}

func TestHandler_MatchResourcesByName(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod"},
		{ID: "res2", Name: "test"},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Name: optionalString("test"),
	})

	assert.Len(t, matched, 1)

	res, ok := matched[0].(*model.Resource)
	assert.True(t, ok)
	assert.Equal(t, "res2", res.ID)
}

func TestHandler_MatchResourcesByNameAndTag(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod", Tags: map[string]string{"env": "prod"}},
		{ID: "res2", Name: "test", Tags: map[string]string{"env": "stage"}},
		{ID: "res3", Name: "test", Tags: map[string]string{"env": "dev"}},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Name: optionalString("test"),
		Tags: map[string]string{"env": "stage"},
	})

	assert.Len(t, matched, 1)

	res, ok := matched[0].(*model.Resource)
	assert.True(t, ok)
	assert.Equal(t, "res2", res.ID)
}

func TestHandler_MatchResourceByNamePrefix(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod", Tags: map[string]string{"env": "prod", "app": "app"}},
		{ID: "res2", Name: "test_a", Tags: map[string]string{"env": "stage", "app": "app"}},
		{ID: "res3", Name: "test_b", Tags: map[string]string{"env": "dev", "app": "app"}},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Name:       optionalString("test"),
		NameFilter: attrs.FilterByPrefix,
		Tags:       map[string]string{"env": "dev"},
	})

	assert.Len(t, matched, 1)

	res, ok := matched[0].(*model.Resource)
	assert.True(t, ok)
	assert.Equal(t, "res3", res.ID)
}

func TestHandler_MatchResourceByNameSuffix(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod_ok", Tags: map[string]string{"env": "prod", "app": "app"}},
		{ID: "res2", Name: "test_ok", Tags: map[string]string{"env": "stage", "app": "app"}},
		{ID: "res3", Name: "test_b", Tags: map[string]string{"env": "dev", "app": "app"}},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Name:       optionalString("ok"),
		NameFilter: attrs.FilterBySuffix,
		Tags:       map[string]string{"env": "prod", "app": "app"},
	})

	assert.Len(t, matched, 1)

	res, ok := matched[0].(*model.Resource)
	assert.True(t, ok)
	assert.Equal(t, "res1", res.ID)
}

func TestHandler_MatchResourceByNameContains(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod_new_ok", Tags: map[string]string{"env": "prod", "app": "app"}},
		{ID: "res2", Name: "test_new_ok", Tags: map[string]string{"env": "stage", "app": "app"}},
		{ID: "res3", Name: "test_b", Tags: map[string]string{"env": "dev", "app": "app"}},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Name:       optionalString("_new_"),
		NameFilter: attrs.FilterByContains,
		Tags:       map[string]string{"env": "stage"},
	})

	assert.Len(t, matched, 1)

	res, ok := matched[0].(*model.Resource)
	assert.True(t, ok)
	assert.Equal(t, "res2", res.ID)
}

func TestHandler_MatchResourceByNameExclude(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod_new_ok", Tags: map[string]string{"env": "prod", "app": "app"}},
		{ID: "res2", Name: "test_new_ok", Tags: map[string]string{"env": "stage", "app": "app"}},
		{ID: "res3", Name: "test_b", Tags: map[string]string{"env": "dev", "app": "app"}},
		{ID: "res4", Name: "test_c", Tags: map[string]string{"env": "test", "app": "app"}},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Name:       optionalString("_new_"),
		NameFilter: attrs.FilterByExclude,
		Tags:       map[string]string{"env": "test"},
	})

	assert.Len(t, matched, 1)

	res, ok := matched[0].(*model.Resource)
	assert.True(t, ok)
	assert.Equal(t, "res4", res.ID)
}

func TestHandler_MatchResourceByNameRegexp(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod_new_ok", Tags: map[string]string{"env": "prod"}},
		{ID: "res2", Name: "test_new_ok", Tags: map[string]string{"env": "stage"}},
		{ID: "res3", Name: "test_b", Tags: map[string]string{"env": "dev", "app": "app"}},
		{ID: "res4", Name: "test_c", Tags: map[string]string{"env": "test", "app": "app"}},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Name:       optionalString("test_*"),
		NameFilter: attrs.FilterByRegexp,
		Tags:       map[string]string{"app": "app"},
	})

	assert.Len(t, matched, 2)

	res, ok := matched[0].(*model.Resource)
	assert.True(t, ok)
	assert.True(t, res.ID == "res3" || res.ID == "res4")
}

func TestHandler_MatchResourceByNameInvalidRegexp(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod_new_ok", Tags: map[string]string{"env": "prod"}},
		{ID: "res2", Name: "test_new_ok", Tags: map[string]string{"env": "stage"}},
		{ID: "res3", Name: "test_b", Tags: map[string]string{"env": "dev", "app": "app"}},
		{ID: "res4", Name: "test_c", Tags: map[string]string{"env": "test", "app": "app"}},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Name:       optionalString("test_{*}"),
		NameFilter: attrs.FilterByRegexp,
		Tags:       map[string]string{"app": "app"},
	})

	assert.Len(t, matched, 0)
}

func TestHandler_MatchResourceByNameWithFilter(t *testing.T) {
	mockResources := []*model.Resource{
		{ID: "res1", Name: "prod_new_ok", Tags: map[string]string{"env": "prod"}},
		{ID: "res2", Name: "test_new_ok", Tags: map[string]string{"env": "stage"}},
		{ID: "res3", Name: "test_b", Tags: map[string]string{"env": "dev", "app": "app"}},
		{ID: "res4", Name: "test_c", Tags: map[string]string{"env": "test", "app": "test_app"}},
	}

	mockFilteredResources := []*model.Resource{
		{ID: "res3", Name: "test_b", Tags: map[string]string{"env": "dev", "app": "app"}},
		{ID: "res4", Name: "test_c", Tags: map[string]string{"env": "test", "app": "app"}},
	}

	handler := &handler[*model.Resource, *model.ResourcesFilter]{
		readResources: func(ctx context.Context) ([]*model.Resource, error) {
			return mockResources, nil
		},
		filterResources: func(ctx context.Context, filter *model.ResourcesFilter) ([]*model.Resource, error) {
			return mockFilteredResources, nil
		},
		filter: &model.ResourcesFilter{
			Tags: map[string]string{"app": "app"},
		},
	}

	err := handler.init()
	assert.NoError(t, err)

	matched := handler.matchResources(&model.ResourcesFilter{
		Tags: map[string]string{"app": "app"},
	})

	assert.Len(t, matched, 2)
}
