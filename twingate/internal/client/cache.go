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

const cacheKey = "cache"

type CacheOptions struct {
	ResourceEnabled bool
	GroupsEnabled   bool
}

var cache = &clientCache{} //nolint:gochecknoglobals

type clientCache struct {
	once     sync.Once
	handlers map[string]resourceHandler
}

type ReadClient interface {
	ReadFullResources(ctx context.Context) ([]*model.Resource, error)
	ReadFullGroups(ctx context.Context) ([]*model.Group, error)
}

func (c *clientCache) setClient(client ReadClient, opts CacheOptions) {
	c.once.Do(func() {
		c.handlers = map[string]resourceHandler{
			reflect.TypeOf(&model.Resource{}).String(): &handler[*model.Resource]{
				enabled:       opts.ResourceEnabled,
				readResources: client.ReadFullResources,
			},
			reflect.TypeOf(&model.Group{}).String(): &handler[*model.Group]{
				enabled:       opts.GroupsEnabled,
				readResources: client.ReadFullGroups,
			},
		}

		group := errgroup.Group{}

		for handlerType, handler := range c.handlers {
			group.Go(func() error {
				if handler.isEnabled() {
					return handler.init()
				}

				log.Printf("[DEBUG] cache init for type %v: skipped.", handlerType)

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			log.Printf("[ERR] cache init failed: %s", err.Error())
		}
	})
}

type resourceHandler interface {
	isEnabled() bool
	init() error
	getResource(resourceID string) (any, bool)
	setResource(resource identifiable)
	invalidateResource(resourceID string)
	matchResources(filter model.ResourceFilter) []any
}

type identifiable interface {
	GetID() string
	Match(filter model.ResourceFilter) bool
}

type readResourcesFunc[T identifiable] func(ctx context.Context) ([]T, error)

type handler[T identifiable] struct {
	once    sync.Once
	enabled bool

	resources     sync.Map
	readResources readResourcesFunc[T]
}

func (h *handler[T]) isEnabled() bool {
	return h.enabled
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

func (h *handler[T]) matchResources(filter model.ResourceFilter) []any {
	var matched []any

	h.resources.Range(func(key, value any) bool {
		obj := value.(T)
		if obj.Match(filter) {
			matched = append(matched, obj)
		}

		return true
	})

	return matched
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
	var initErr error

	h.once.Do(func() {
		var (
			res T
		)

		log.Printf("[DEBUG] cache init for type %T: started.", res)

		resources, err := h.readResources(WithCallerCtx(context.Background(), cacheKey))
		if err != nil {
			log.Printf("[ERR] cache init for type %T failed: %s", res, err.Error())

			initErr = err

			return
		}

		h.setResources(resources)

		log.Printf("[DEBUG] cache init for type %T: finished.", res)
	})

	return initErr
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

func matchResources[T any](filter model.ResourceFilter) []T {
	var (
		res     T
		matched []T
	)

	handle(res, func(handler resourceHandler) {
		resources := handler.matchResources(filter)
		for _, resource := range resources {
			obj, ok := resource.(T)
			if !ok {
				log.Printf("[ERR] matchResources failed: expected type %T, got %T", res, resource)

				return
			}

			matched = append(matched, obj)
		}
	})

	return matched
}

func lazyLoadResources[T any]() {
	var (
		res T
	)

	handle(res, func(handler resourceHandler) {
		if err := handler.init(); err != nil {
			log.Printf("[ERR] lazyLoadResources for type %T failed: %s", res, err.Error())
		}
	})
}
