package twingate

import (
	"fmt"
	"log"

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
		return fmt.Errorf("error getting client: %s", err)
	}

	resourceList, ok := client.readAllResources()
	if !ok {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	for _, i := range resourceList {
		if i == nil {
			log.Printf("[INFO][SWEEPER_LOG] %s resource name was nil", resourceName)
			return nil
		}
		resourceResourceDelete(i)
	}

	return nil
}
