package twingate

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("twingate_remote_network", &resource.Sweeper{
		Name: "twingate_remote_network",
		F:    testSweepTwingateRemoteNetwork,
	})
}

func testSweepTwingateRemoteNetwork(tenant string) error {
	resourceName := "twingate_remote_network"
	log.Printf("\"[INFO][SWEEPER_LOG] Starting sweeper for %s\"", resourceName)
	client, err := sharedClient(tenant)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	networkList, err := client.readAllRemoteNetwork()
	if err != nil {
		return fmt.Errorf("[INFO][SWEEPER_LOG] Nothing found in response: %s", err)
	}

	if len(networkList) == 0 {
		fmt.Printf("[INFO][SWEEPER_LOG] List %s is empty", resourceName)
		return nil
	}

	for _, i := range networkList {
		if i == "" {
			return fmt.Errorf("[INFO][SWEEPER_LOG] %s resource name was nil", resourceName)
		}
		err = client.deleteRemoteNetwork(i)
		if err != nil {
			return fmt.Errorf("[INFO][SWEEPER_LOG] %s cannot be deleted", err)
		}
	}

	return nil
}
