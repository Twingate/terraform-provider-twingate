package client

import (
	"context"
	"sync"
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
	resources:          map[string]*model.Resource{},
	requestedResources: make(chan string, requestsBufferSize),
}

func init() { //nolint:gochecknoinits
	closedChan = make(chan struct{})
	close(closedChan)

	go cache.run()
}

type clientCache struct {
	lock      sync.RWMutex
	resources map[string]*model.Resource

	requestDone        bool
	requestedResources chan string

	client *Client
}

func (c *clientCache) done() <-chan struct{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if !c.requestDone {
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

			c.lock.RLock()
			isDone := c.requestDone
			c.lock.RUnlock()

			if isDone {
				c.lock.Lock()
				c.requestDone = false
				c.lock.Unlock()
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
	c.lock.Lock()
	c.requestDone = true
	c.lock.Unlock()
}

func (c *clientCache) getResource(resourceID string) (*model.Resource, bool) {
	c.lock.RLock()

	if c.client == nil {
		c.lock.RUnlock()

		return nil, false
	}

	c.lock.RUnlock()

	c.lock.RLock()
	res, exists := c.resources[resourceID]
	c.lock.RUnlock()

	if exists {
		return res, exists
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

	c.lock.RLock()
	res, exists = c.resources[resourceID]
	c.lock.RUnlock()

	return res, exists
}

func (c *clientCache) setResource(resource *model.Resource) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.resources[resource.ID] = resource
}

func (c *clientCache) setResources(resources []*model.Resource) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, resource := range resources {
		c.resources[resource.ID] = resource
	}
}

func (c *clientCache) invalidateResource(id string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.resources, id)
}
