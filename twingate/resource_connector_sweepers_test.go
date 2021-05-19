package twingate

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("twingate_connector", &resource.Sweeper{
		Name: "twingate_connector",
		F:    testSweepTwingateConnector,
	})
}

func testSweepTwingateConnector(tenant string) error {
	resourceName := "twingate_connector"
	log.Printf("\"[INFO][SWEEPER_LOG] Starting sweeper for %s\"", resourceName)
	client, err := sharedClient(tenant)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	connectorList, err := client.readAllConnectors()
	if err != nil {
		return fmt.Errorf("[INFO][SWEEPER_LOG] Nothing found in response: %s", err)
	}

	for _, i := range connectorList {
		if i == "" {
			return fmt.Errorf("[INFO][SWEEPER_LOG] %s resource name was nil", resourceName)
		}
		err = client.deleteConnector(i)
		if err != nil {
			return fmt.Errorf("[INFO][SWEEPER_LOG] %s cannot be deleted", err)
		}
	}

	return nil
}
