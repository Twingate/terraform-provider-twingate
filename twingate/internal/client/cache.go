package client

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

var closedChan chan struct{} //nolint:gochecknoglobals

const (
	minBulkSize        = 10
	requestsBufferSize = 1000
	collectTime        = 70 * time.Millisecond
	shortWaitTime      = 5 * time.Millisecond
)

var cache = &clientCache{ //nolint:gochecknoglobals
	requestedResources: make(chan string, requestsBufferSize),
	handlers: map[ResourceWithID]*handler[ResourceWithID]{
		*model.Resource: handler[*model.Resource]{
			readResources:
		},
	},
}

func init() { //nolint:gochecknoinits
	closedChan = make(chan struct{})
	close(closedChan)

	cache.requestDone.Store(false)

	go cache.run()
}

type clientCache struct {
	lock               sync.RWMutex
	resources          sync.Map
	requestDone        atomic.Bool
	requestedResources chan string

	handlers map[reflect.Type]*handler[ResourceWithID]

	client *Client
}

func (c *clientCache) setClient (client *Client) {
	c.client = client

	c.handlers = map[reflect.Type]*handler[ResourceWithID]{
		type(*model.Resource): &handler[*model.Resource]{
			readResources: client.readResources,
		},
	}
}

// --

type ResourceWithID interface {
	GetID() string
}

type ReadResources[T ResourceWithID] func(ctx context.Context) ([]T, error)

type handler[T ResourceWithID] struct {
	resources          sync.Map
	requestDone        atomic.Bool
	requestedResources chan string

	readResources ReadResources[T]

	//client *Client
}

func (h *handler[T]) getResource(resourceID string) (T, bool) {
	if h.readResources == nil {
		var emptyObj T
		return emptyObj, false
	}

	res, exists := h.resources.Load(resourceID)

	if exists {
		return res.(T), exists
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

	return res.(T), exists
}

func (h *handler[T]) setResource(resource T) {
	h.resources.Store(resource.GetID(), resource)
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

// --

func (c *clientCache) done() <-chan struct{} {
	//c.lock.RLock()
	//defer c.lock.RUnlock()

	if !c.requestDone.Load() {
		return nil
	}

	return closedChan
}

func (c *clientCache) run() { //nolint
	var collectTimer *time.Timer

	resourcesToRequest := make(map[string]bool)

	for {
		select {
		case id := <-c.requestedResources:
			resourcesToRequest[id] = true

			//c.lock.RLock()
			//isDone := c.requestDone
			//c.lock.RUnlock()

			if c.requestDone.Load() {
				//c.lock.Lock()
				//c.requestDone = false
				//c.lock.Unlock()
				c.requestDone.Store(false)
			}

			if collectTimer == nil {
				collectTimer = time.NewTimer(collectTime)

				continue
			} else {
				select {
				case <-collectTimer.C:
					collectTimer = nil

					c.fetchResources(resourcesToRequest)
					resourcesToRequest = make(map[string]bool)

				default: // no op
				}
			}

		default:
			if collectTimer != nil {
				select {
				case <-collectTimer.C:
					collectTimer = nil

					c.fetchResources(resourcesToRequest)
					resourcesToRequest = make(map[string]bool)

				default: // no op
				}
			}

			time.Sleep(shortWaitTime)
		}
	}
}

func (c *clientCache) fetchResources(resourcesToRequest map[string]bool) {
	if len(resourcesToRequest) >= minBulkSize && c.client != nil {
		resources, err := c.client.ReadFullResources(context.Background())
		if err == nil {
			c.setResources(resources)
		}
	}

	// notify
	//c.lock.Lock()
	//c.requestDone = true
	//c.lock.Unlock()
	c.requestDone.Store(true)
}

func (c *clientCache) getResource(resourceID string) (*model.Resource, bool) {
	c.lock.RLock()

	if c.client == nil {
		c.lock.RUnlock()

		return nil, false
	}

	c.lock.RUnlock()

	//c.lock.RLock()
	//res, exists := c.resources[resourceID]
	//c.lock.RUnlock()
	res, exists := c.resources.Load(resourceID)

	if exists {
		return res.(*model.Resource), exists
	}

	c.requestedResources <- resourceID
	// wait for fetching
LOOP:
	for {
		select {
		case <-c.done():
			break LOOP

		default:
			time.Sleep(shortWaitTime)
		}
	}

	//c.lock.RLock()
	//res, exists = c.resources[resourceID]
	//c.lock.RUnlock()

	res, exists = c.resources.Load(resourceID)

	return res.(*model.Resource), exists
}

func (c *clientCache) setResource(resource *model.Resource) {
	//c.lock.Lock()
	//defer c.lock.Unlock()
	//
	//c.resources[resource.ID] = resource

	c.resources.Store(resource.ID, resource)
}

func (c *clientCache) setResources(resources []*model.Resource) {
	//c.lock.Lock()
	//defer c.lock.Unlock()

	for _, resource := range resources {
		//c.resources[resource.ID] = resource
		c.resources.Store(resource.ID, resource)
	}
}

func (c *clientCache) invalidateResource(id string) {
	//c.lock.Lock()
	//defer c.lock.Unlock()
	//
	//delete(c.resources, id)

	c.resources.Delete(id)
}
