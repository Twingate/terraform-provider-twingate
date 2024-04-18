package client

import (
	"context"
	"sync"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
)

var closedChan chan struct{}

func init() {
	closedChan = make(chan struct{})
	close(closedChan)

	go cache.run()
}

var cache = &clientCache{
	resources:          map[string]*model.Resource{},
	requestedResources: make(chan string, 1000),
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

func (c *clientCache) run() {
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
				collectTimer = time.NewTimer(time.Millisecond * 50)
				continue
			} else {
				select {
				case <-collectTimer.C:
					collectTimer = nil
					c.fetchResources(resourcesToRequest)
					resourcesToRequest = make(map[string]bool)

				default:
					// no op
				}
			}

		default:
			if collectTimer != nil {
				select {
				case <-collectTimer.C:
					collectTimer = nil
					c.fetchResources(resourcesToRequest)
					resourcesToRequest = make(map[string]bool)

				default:
					// no op
				}
			}

			time.Sleep(time.Millisecond * 5)
		}
	}
}

func (c *clientCache) fetchResources(resourcesToRequest map[string]bool) {
	// TODO: query only required resources
	resources, _ := c.client.ReadResources(context.Background())
	c.setResources(resources)

	// notify
	c.lock.Lock()
	c.requestDone = true
	c.lock.Unlock()
}

func (c *clientCache) getResource(resourceID string) (*model.Resource, bool) {
	c.lock.RLock()
	res, ok := c.resources[resourceID]
	c.lock.RUnlock()

	if ok {
		return res, ok
	}

	c.requestedResources <- resourceID
	// wait for fetching
LOOP:
	for {
		select {
		case <-c.done():
			break LOOP

		default:
			time.Sleep(time.Millisecond * 10)
		}
	}

	c.lock.RLock()
	res, ok = c.resources[resourceID]
	c.lock.RUnlock()

	return res, ok
}

func (c *clientCache) setResource(resource *model.Resource) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.resources[resource.ID] = resource
}

func (c *clientCache) deleteResource(resourceID string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.resources, resourceID)
}

func (c *clientCache) setResources(resources []*model.Resource) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, resource := range resources {
		res := resource
		c.resources[res.ID] = res
	}
}
