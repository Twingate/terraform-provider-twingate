package client

import (
	"context"
	"log"
	"reflect"
	"sync"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/mitchellh/copystructure"
	"golang.org/x/sync/errgroup"
)

var cache = &clientCache{} //nolint:gochecknoglobals

type clientCache struct {
	once     sync.Once
	handlers map[string]resourceHandler
}

func (c *clientCache) setClient(client *Client) {
	c.once.Do(func() {
		c.handlers = map[string]resourceHandler{
			reflect.TypeOf(&model.Resource{}).String(): &handler[*model.Resource]{
				readResources: client.ReadFullResources,
			},
			reflect.TypeOf(&model.Group{}).String(): &handler[*model.Group]{
				readResources: client.ReadFullGroups,
			},
		}

		group := errgroup.Group{}

		for _, handler := range c.handlers {
			group.Go(func() error {
				return handler.init()
			})
		}

		if err := group.Wait(); err != nil {
			log.Printf("[ERR] cache init failed: %s", err.Error())
		}
	})
}

type resourceHandler interface {
	init() error
	getResource(resourceID string) (any, bool)
	setResource(resource identifiable)
	invalidateResource(resourceID string)
}

type identifiable interface {
	GetID() string
}

type readResourcesFunc[T identifiable] func(ctx context.Context) ([]T, error)

type handler[T identifiable] struct {
	resources     sync.Map
	readResources readResourcesFunc[T]
}

func (h *handler[T]) getResource(resourceID string) (any, bool) {
	var emptyObj T

	if h.readResources == nil {
		return emptyObj, false
	}

	res, exists := h.resources.Load(resourceID)

	if !exists {
		return emptyObj, false
	}

	obj, err := copystructure.Copy(res)

	if err != nil {
		log.Printf("[ERR] %T failed copy object from cache: %s", emptyObj, err.Error())

		return emptyObj, false
	}

	return obj, exists
}

func (h *handler[T]) setResource(resource identifiable) {
	if resource == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERR] setResource failed: %v", r)
		}
	}()

	obj, err := copystructure.Copy(resource)

	if err != nil {
		log.Printf("[ERR] %T failed store object to cache: %s", resource, err.Error())

		return
	}

	h.resources.Store(resource.GetID(), obj)
}

func (h *handler[T]) setResources(resources []T) {
	for _, resource := range resources {
		h.setResource(resource)
	}
}

func (h *handler[T]) invalidateResource(id string) {
	h.resources.Delete(id)
}

func (h *handler[T]) init() error {
	resources, err := h.readResources(context.Background())
	if err != nil {
		return err
	}

	h.setResources(resources)

	return nil
}

func getResource[T any](resourceID string) (T, bool) {
	var (
		res    T
		exists bool
	)

	handle(res, func(handler resourceHandler) {
		resource, found := handler.getResource(resourceID)
		if !found || resource == nil {
			return
		}

		obj, ok := resource.(T)
		if !ok {
			log.Printf("[ERR] getResource failed: expected type %T, got %T", res, resource)

			return
		}

		res = obj
		exists = found
	})

	return res, exists
}

func setResource(resource identifiable) {
	handle(resource, func(handler resourceHandler) {
		handler.setResource(resource)
	})
}

func invalidateResource[T any](resourceID string) {
	var res T

	handle(res, func(handler resourceHandler) {
		handler.invalidateResource(resourceID)
	})
}

func handle(handlerType any, apply func(handler resourceHandler)) {
	if handler, ok := cache.handlers[handlerKey(handlerType)]; ok {
		apply(handler)
	}
}

func handlerKey(handlerType any) string {
	return reflect.TypeOf(handlerType).String()
}
