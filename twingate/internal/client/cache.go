package client

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/mitchellh/copystructure"
)

var closedChan chan struct{} //nolint:gochecknoglobals

const (
	minBulkSize        = 10
	requestsBufferSize = 1000
	collectTime        = 2 * time.Second
	shortWaitTime      = 5 * time.Millisecond
)

var cache = &clientCache{} //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	closedChan = make(chan struct{})
	close(closedChan)
}

type clientCache struct {
	lock sync.RWMutex

	handlers map[string]resourceHandler
}

func (c *clientCache) setClient(client *Client) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.handlers != nil {
		return
	}

	c.handlers = map[string]resourceHandler{
		reflect.TypeOf(&model.Resource{}).String(): &handler[*model.Resource]{
			readResources:      client.ReadFullResources,
			requestedResources: make(chan string, requestsBufferSize),
		},
		reflect.TypeOf(&model.Group{}).String(): &handler[*model.Group]{
			readResources:      client.ReadFullGroups,
			requestedResources: make(chan string, requestsBufferSize),
		},
	}

	for _, worker := range c.handlers {
		go worker.run()
	}
}

type resourceHandler interface {
	run()
	getResource(resourceID string) (any, bool)
	setResource(resource identifiable)
	invalidateResource(resourceID string)
}

type identifiable interface {
	GetID() string
}

type readResourcesFunc[T identifiable] func(ctx context.Context) ([]T, error)

type handler[T identifiable] struct {
	resources          sync.Map
	requestDone        atomic.Bool
	requestedResources chan string
	readResources      readResourcesFunc[T]
}

func (h *handler[T]) getResource(resourceID string) (any, bool) {
	var emptyObj T

	if h.readResources == nil {
		return emptyObj, false
	}

	res, exists := h.resources.Load(resourceID)

	if exists {
		obj, _ := copystructure.Copy(res)

		return obj, exists
	}

	h.requestedResources <- resourceID
	// wait for fetching
LOOP:
	for {
		select {
		case <-h.done():
			break LOOP

		default:
			time.Sleep(shortWaitTime)
		}
	}

	res, exists = h.resources.Load(resourceID)

	if exists {
		obj, _ := copystructure.Copy(res)

		return obj, exists
	}

	return emptyObj, false
}

func (h *handler[T]) setResource(resource identifiable) {
	obj, _ := copystructure.Copy(resource)

	h.resources.Store(resource.GetID(), obj)
}

func (h *handler[T]) setResources(resources []T) {
	for _, resource := range resources {
		h.resources.Store(resource.GetID(), resource)
	}
}

func (h *handler[T]) invalidateResource(id string) {
	h.resources.Delete(id)
}

func (h *handler[T]) done() <-chan struct{} {
	if !h.requestDone.Load() {
		return nil
	}

	return closedChan
}

func (h *handler[T]) fetchResources(resourcesToRequest map[string]bool) {
	if len(resourcesToRequest) >= minBulkSize && h.readResources != nil {
		resources, err := h.readResources(context.Background())
		if err == nil {
			h.setResources(resources)
		}
	}

	// notify
	h.requestDone.Store(true)
}

func (h *handler[T]) run() { //nolint
	var collectTimer *time.Timer

	resourcesToRequest := make(map[string]bool)

	for {
		select {
		case id := <-h.requestedResources:
			resourcesToRequest[id] = true

			if h.requestDone.Load() {
				h.requestDone.Store(false)
			}

			if collectTimer == nil {
				collectTimer = time.NewTimer(collectTime)

				continue
			} else {
				select {
				case <-collectTimer.C:
					collectTimer = nil

					h.fetchResources(resourcesToRequest)
					resourcesToRequest = make(map[string]bool)

				default: // no op
				}
			}

		default:
			if collectTimer != nil {
				select {
				case <-collectTimer.C:
					collectTimer = nil

					h.fetchResources(resourcesToRequest)
					resourcesToRequest = make(map[string]bool)

				default: // no op
				}
			}

			time.Sleep(shortWaitTime)
		}
	}
}

func getResource[T any](resourceID string) (T, bool) {
	var (
		res    T
		exists bool
	)

	handle(res, func(handler resourceHandler) {
		resource, found := handler.getResource(resourceID)
		if resource == nil {
			return
		}

		res = resource.(T)
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
