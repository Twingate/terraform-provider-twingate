package twingate

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRemoteConnector_basic(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector", func(t *testing.T) {
		remoteNetworkName := acctest.RandomWithPrefix("tf-acc")
		connectorName := acctest.RandomWithPrefix("tf-acc")
		connectorResource := "twingate_connector.test"
		remoteNetworkResource := "twingate_remote_network.test"

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnector(remoteNetworkName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
					),
				},
				{
					Config: testTwingateConnectorWithCustomName(remoteNetworkName, connectorName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
						resource.TestMatchResourceAttr(connectorResource, "name", regexp.MustCompile("tf-acc.*")),
					),
				},
			},
		})
	})
}

func testTwingateConnector(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_connector" "test" {
	  remote_network_id = twingate_remote_network.test.id
	}
	`, remoteNetworkName)
}

func testTwingateConnectorWithCustomName(remoteNetworkName string, connectorName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_connector" "test" {
	  remote_network_id = twingate_remote_network.test.id
      name  = "%s"
	}
	`, remoteNetworkName, connectorName)
}

func testAccCheckTwingateConnectorDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_connector" {
			continue
		}

		connectorId := rs.Primary.ID

		err := client.deleteConnector(connectorId)
		// expecting error here , since the network is already gone
		if err == nil {
			return fmt.Errorf("Connector with ID %s still present : ", connectorId)
		}
	}

	return nil
}

func testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		connector, ok := s.RootModule().Resources[connectorResource]

		if !ok {
			return fmt.Errorf("Not found: %s ", connectorResource)
		}

		if connector.Primary.ID == "" {
			return fmt.Errorf("No connectorID set ")
		}
		remoteNetwork, ok := s.RootModule().Resources[remoteNetworkResource]
		if !ok {
			return fmt.Errorf("Not found: %s ", remoteNetworkResource)
		}
		if connector.Primary.Attributes["remote_network_id"] != remoteNetwork.Primary.ID {
			return fmt.Errorf("Remote Network ID not set properly in the connector ")
		}
		return nil
	}
}
