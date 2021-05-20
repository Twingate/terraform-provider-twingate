package twingate

import (
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
		log.Printf("[ERROR][SWEEPER_LOG] error getting client: %s", err)
		return err
	}

	networkList, err := client.readAllRemoteNetwork()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response: %s", resourceName)
		return nil
	}

	if len(networkList) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List %s is empty", resourceName)
		return nil
	}

	for _, i := range networkList {
		if i == "" {
			log.Printf("[INFO][SWEEPER_LOG] %s: %s name was empty value", resourceName, i)
			return nil
		}
		err = client.deleteRemoteNetwork(i)
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] %s cannot be deleted, error: %s", i, err)
			return nil
		}
	}

	return nil
}
