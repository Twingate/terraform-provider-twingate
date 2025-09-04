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
	ResourcesFilter *model.ResourcesFilter
	GroupsFilter    *model.GroupsFilter
}

var cache = &clientCache{} //nolint:gochecknoglobals

type clientCache struct {
	once     sync.Once
	handlers map[string]resourceHandler
}

type ReadClient interface {
	ReadFullResources(ctx context.Context) ([]*model.Resource, error)
	ReadFullGroups(ctx context.Context) ([]*model.Group, error)

	ReadFullResourcesByName(ctx context.Context, filter *model.ResourcesFilter) ([]*model.Resource, error)
	ReadFullGroupsByName(ctx context.Context, filter *model.GroupsFilter) ([]*model.Group, error)
}

func (c *clientCache) setClient(client ReadClient, opts CacheOptions) {
	c.once.Do(func() {
		c.handlers = map[string]resourceHandler{
			reflect.TypeOf(&model.Resource{}).String(): &handler[*model.Resource, *model.ResourcesFilter]{
				enabled:         opts.ResourceEnabled,
				readResources:   client.ReadFullResources,
				filter:          opts.ResourcesFilter,
				filterResources: client.ReadFullResourcesByName,
			},
			reflect.TypeOf(&model.Group{}).String(): &handler[*model.Group, *model.GroupsFilter]{
				enabled:         opts.GroupsEnabled,
				readResources:   client.ReadFullGroups,
				filter:          opts.GroupsFilter,
				filterResources: client.ReadFullGroupsByName,
			},
		}

		group := errgroup.Group{}

		for handlerType, handler := range c.handlers {
			group.Go(func() error {
				if handler.isEnabled() {
					return handler.init()
				}

				log.Printf("[TWINGATE_LOG] cache init for type %v: skipped.", handlerType)

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			log.Printf("[TWINGATE_LOG] [ERR] cache init failed: %s", err.Error())
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

type handler[T identifiable, F any] struct {
	once    sync.Once
	enabled bool
	filter  F

	resources       sync.Map
	readResources   readResourcesFunc[T]
	filterResources func(ctx context.Context, filter F) ([]T, error)
}

func (h *handler[T, F]) isEnabled() bool {
	return h.enabled
}

func (h *handler[T, F]) getResource(resourceID string) (any, bool) {
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
		log.Printf("[TWINGATE_LOG] [ERR] %T failed copy object from cache: %s", emptyObj, err.Error())

		return emptyObj, false
	}

	return obj, exists
}

func (h *handler[T, F]) matchResources(filter model.ResourceFilter) []any {
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

func (h *handler[T, F]) setResource(resource identifiable) {
	if resource == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[TWINGATE_LOG] [ERR] setResource failed: %v", r)
		}
	}()

	obj, err := copystructure.Copy(resource)

	if err != nil {
		log.Printf("[TWINGATE_LOG] [ERR] %T failed store object to cache: %s", resource, err.Error())

		return
	}

	h.resources.Store(resource.GetID(), obj)
}

func (h *handler[T, F]) setResources(resources []T) {
	for _, resource := range resources {
		h.setResource(resource)
	}
}

func (h *handler[T, F]) invalidateResource(id string) {
	h.resources.Delete(id)
}

func (h *handler[T, F]) init() error {
	var initErr error

	h.once.Do(func() {
		var (
			res       T
			err       error
			resources []T
		)

		log.Printf("[TWINGATE_LOG] cache init for type %T: started. Filter set: %v.", res, !isNil(h.filter))

		if isNil(h.filter) {
			// read all resources
			resources, err = h.readResources(WithCallerCtx(context.Background(), cacheKey))
		} else {
			log.Printf("[TWINGATE_LOG] cache init for type %T: started. Applying filter: %v.", res, h.filter)

			// read filtered resources
			resources, err = h.filterResources(WithCallerCtx(context.Background(), cacheKey), h.filter)
		}

		if err != nil {
			log.Printf("[TWINGATE_LOG] [ERR] cache init for type %T failed: %s", res, err.Error())

			initErr = err

			return
		}

		h.setResources(resources)

		log.Printf("[TWINGATE_LOG] cache init for type %T: finished.", res)
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
			log.Printf("[TWINGATE_LOG] [ERR] getResource failed: expected type %T, got %T", res, resource)

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
				log.Printf("[TWINGATE_LOG] [ERR] matchResources failed: expected type %T, got %T", res, resource)

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
			log.Printf("[TWINGATE_LOG] [ERR] lazyLoadResources for type %T failed: %s", res, err.Error())
		}
	})
}

func isNil(obj any) bool {
	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}
