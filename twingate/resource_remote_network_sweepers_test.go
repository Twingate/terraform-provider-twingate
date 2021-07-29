package twingate

import (
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/twingate/go-graphql-client"
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

	networkMap, err := client.readRemoteNetworks()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response: %s", resourceName)
		return nil
	}

	if len(networkMap) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List %s is empty", resourceName)
		return nil
	}

	var testNetworks = make([]graphql.ID, 0)

	for _, elem := range networkMap {
		if strings.HasPrefix(string(elem.Name), "tf-acc") {
			testNetworks = append(testNetworks, elem.ID)
		}
	}

	if len(testNetworks) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List with test networks %s is empty", resourceName)
		return nil
	}

	for _, i := range testNetworks {
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
