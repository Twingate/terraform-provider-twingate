package sweepers

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("twingate_resource", &resource.Sweeper{
		Name: "twingate_resource",
		F:    testSweepTwingateResource,
	})
}

func testSweepTwingateResource(tenant string) error {
	resourceName := "twingate_resource"
	log.Printf("\"[INFO][SWEEPER_LOG] Starting sweeper for %s\"", resourceName)
	client, err := sharedClient(tenant)
	if err != nil {
		log.Printf("[ERROR][SWEEPER_LOG] error getting client: %s", err)
		return err
	}

	ctx := context.Background()

	resources, err := client.readResources(ctx)
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response: %s", resourceName)
		return nil
	}

	if len(resources) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List %s is empty", resourceName)
		return nil
	}

	var testResources = make([]string, 0)

	testPrefix := getTestPrefix()
	for _, elem := range resources {
		if strings.HasPrefix(elem.Node.StringName(), testPrefix) {
			testResources = append(testResources, elem.Node.StringID())
		}
	}

	if len(testResources) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List with test networks %s is empty", resourceName)
		return nil
	}

	for _, i := range testResources {
		if i == "" {
			log.Printf("[INFO][SWEEPER_LOG] %s: %s name was empty value", resourceName, i)
			continue
		}
		err = client.deleteResource(ctx, i)
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] %s cannot be deleted, error: %s", i, err)
			continue
		}
	}

	return nil
}
