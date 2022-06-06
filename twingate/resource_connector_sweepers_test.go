package twingate

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testPrefixName = "tf-acc"

func init() {
	resource.AddTestSweepers("twingate_connector", &resource.Sweeper{
		Name: "twingate_connector",
		F:    testSweepTwingateConnector,
	})
}

func testSweepTwingateConnector(tenant string) error {
	resourceName := "twingate_connector"
	log.Printf("[INFO][SWEEPER_LOG] Starting sweeper for %s", resourceName)
	client, err := sharedClient(tenant)
	if err != nil {
		log.Printf("[ERROR][SWEEPER_LOG] error getting client: %s", err)
		return err
	}

	ctx := context.Background()

	connectorMap, err := client.readConnectors(ctx)
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response: %s", resourceName)
		return nil
	}

	if len(connectorMap) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List %s is empty", resourceName)
		return nil
	}

	var testConnectors = make([]string, 0)

	for _, elem := range connectorMap {
		if strings.HasPrefix(elem.Name, testPrefixName) {
			testConnectors = append(testConnectors, elem.ID)
		}
	}

	if len(testConnectors) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List with test connectors %s is empty", resourceName)
		return nil
	}

	for _, i := range testConnectors {
		if i == "" {
			log.Printf("[INFO][SWEEPER_LOG] %s: %s name was empty value", resourceName, i)
			return nil
		}
		err = client.deleteConnector(ctx, i)
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] %s cannot be deleted, error: %s", i, err)
			return nil
		}
	}

	return nil
}
