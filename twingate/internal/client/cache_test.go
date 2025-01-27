package client

import (
	"context"
	"reflect"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
}

func (m mockClient) ReadFullResources(ctx context.Context) ([]*model.Resource, error) {
	return []*model.Resource{}, nil
}

func (m mockClient) ReadFullGroups(ctx context.Context) ([]*model.Group, error) {
	return []*model.Group{}, nil
}

func TestClientCache_SetClient(t *testing.T) {
	cache := &clientCache{}
	cache.setClient(&mockClient{})

	assert.NotNil(t, cache.handlers)
	assert.Contains(t, cache.handlers, reflect.TypeOf(&model.Resource{}).String())
	assert.Contains(t, cache.handlers, reflect.TypeOf(&model.Group{}).String())
}

func TestHandler_SetAndGetResource(t *testing.T) {
	mockResource := &model.Resource{ID: "resource1"}
	handler := &handler[*model.Resource]{
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
	handler := &handler[*model.Resource]{
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

	handler := &handler[*model.Resource]{
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
