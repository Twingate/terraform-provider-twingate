package sweepers

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

type Resource interface {
	GetID() string
	GetName() string
}

type readResourcesFunc func(client *client.Client, ctx context.Context) ([]Resource, error)

type deleteResourceFunc func(client *client.Client, ctx context.Context, id string) error

func newTestSweeper(resourceName string, readResources readResourcesFunc, deleteResource deleteResourceFunc) func(tenant string) error {
	return func(tenant string) error {
		log.Printf("[INFO][SWEEPER_LOG] %s: starting sweeper", resourceName)
		defer log.Printf("[INFO][SWEEPER_LOG] %s: DONE", resourceName)

		client, err := test.TwingateClient()
		if err != nil {
			log.Printf("[ERROR][SWEEPER_LOG] %s: failed to create client: %v", resourceName, err)
			return err
		}

		ctx := context.Background()

		resources, err := readResources(client, ctx)
		if err != nil {
			log.Printf("[ERROR][SWEEPER_LOG] %s: failed to read resources: %v", resourceName, err)
			return nil
		}

		if len(resources) == 0 {
			log.Printf("[INFO][SWEEPER_LOG] %s: empty result", resourceName)
			return nil
		}

		var ids = make([]string, 0, len(resources))

		testPrefix := test.Prefix()
		for _, elem := range resources {
			if strings.HasPrefix(elem.GetName(), testPrefix) {
				ids = append(ids, elem.GetID())
			}
		}

		if len(ids) == 0 {
			log.Printf("[INFO][SWEEPER_LOG] %s: after filter by test prefix got empty result", resourceName)
			return nil
		}

		for _, id := range ids {
			if id == "" {
				log.Printf("[WARN][SWEEPER_LOG] %s: got resource with empty id", resourceName)
				continue
			}

			err = deleteResource(client, ctx, id)
			if err != nil {
				log.Printf("[ERROR][SWEEPER_LOG] %s: failed to delete resource with id %s: %v", resourceName, id, err)
				continue
			}
		}

		return nil
	}
}
