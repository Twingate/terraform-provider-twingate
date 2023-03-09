package sweepers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func getEnv(key string, duration time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		parsedDuration, err := time.ParseDuration(value)
		if err != nil {
			return duration
		}
		return parsedDuration
	}
	return duration
}

// sharedClient returns a common TwingateClient setup needed for the sweeper
func sharedClient(tenant string) (*client.Client, error) {
	if os.Getenv(twingate.EnvAPIToken) == "" {
		return nil, fmt.Errorf("must provide environment variable %s", twingate.EnvAPIToken)
	}

	if os.Getenv(twingate.EnvNetwork) == "" {
		return nil, fmt.Errorf("must provide environment variable %s", twingate.EnvNetwork)
	}

	if os.Getenv(twingate.EnvURL) == "" {
		return nil, fmt.Errorf("must provide environment variable %s", twingate.EnvURL)
	}

	return client.NewClient(
			os.Getenv(twingate.EnvURL),
			os.Getenv(twingate.EnvAPIToken),
			os.Getenv(twingate.EnvNetwork),
			getEnv(twingate.EnvHTTPTimeout, 30*time.Second),
			2,
			"sweeper"),
		nil
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

		client, err := sharedClient(tenant)
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
