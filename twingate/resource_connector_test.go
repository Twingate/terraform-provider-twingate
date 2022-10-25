package twingate

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	connectorResource     = "twingate_connector.test"
	remoteNetworkResource = "twingate_remote_network.test"
	nameAttr              = "name"
)

var testRegexp = regexp.MustCompile(getTestPrefix() + ".*")

func TestAccRemoteConnectorCreate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector", func(t *testing.T) {
		remoteNetworkName := getRandomName()

		const theResource = "twingate_connector.test_c1"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: createConnectorC1(remoteNetworkName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(theResource, "twingate_remote_network.test_c1"),
						resource.TestCheckResourceAttrSet(theResource, "name"),
					),
				},
			},
		})
	})
}

func createConnectorC1(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_c1" {
	  name = "%s"
	}
	resource "twingate_connector" "test_c1" {
	  remote_network_id = twingate_remote_network.test_c1.id
	}
	`, remoteNetworkName)
}

func TestAccRemoteConnectorWithCustomName(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector With Custom Name", func(t *testing.T) {
		remoteNetworkName := getRandomName()
		connectorName := getRandomConnectorName()

		const theResource = "twingate_connector.test_c2"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: createConnectorC2(remoteNetworkName, connectorName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(theResource, "twingate_remote_network.test_c2"),
						resource.TestMatchResourceAttr(theResource, "name", regexp.MustCompile(connectorName)),
					),
				},
			},
		})
	})
}

func createConnectorC2(remoteNetworkName string, connectorName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_c2" {
	  name = "%s"
	}
	resource "twingate_connector" "test_c2" {
	  remote_network_id = twingate_remote_network.test_c2.id
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

		err := client.deleteConnector(context.Background(), connectorId)
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

func TestAccRemoteConnectorImport(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector - Import", func(t *testing.T) {
		remoteNetworkName := getRandomName()
		connectorName := getRandomConnectorName()
		const theResource = "twingate_connector.test_c3"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: createConnectorC3(remoteNetworkName, connectorName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(theResource, "twingate_remote_network.test_c3"),
						resource.TestMatchResourceAttr(theResource, "name", testRegexp),
					),
				},
				{
					ResourceName:      theResource,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func createConnectorC3(remoteNetworkName string, connectorName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_c3" {
	  name = "%s"
	}
	resource "twingate_connector" "test_c3" {
	  remote_network_id = twingate_remote_network.test_c3.id
      name  = "%s"
	}
	`, remoteNetworkName, connectorName)
}

func TestAccRemoteConnectorNotAllowedToChangeRemoteNetworkId(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector - should fail on remote_network_id update", func(t *testing.T) {
		remoteNetworkName := getRandomName()
		newRemoteNetworkName := getRandomName()
		const theResource = "twingate_connector.test_c4"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: createConnectorC4(remoteNetworkName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(theResource, "twingate_remote_network.test_c4_1"),
					),
				},
				{
					Config:      createConnectorC4WithAnotherNetwork(newRemoteNetworkName),
					ExpectError: regexp.MustCompile(ErrNotAllowChangeRemoteNetworkID.Error()),
				},
			},
		})
	})
}

func createConnectorC4(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_c4_1" {
	  name = "%s"
	}
	resource "twingate_connector" "test_c4" {
	  remote_network_id = twingate_remote_network.test_c4_1.id
	}
	`, remoteNetworkName)
}

func createConnectorC4WithAnotherNetwork(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_c4_2" {
	  name = "%s"
	}
	resource "twingate_connector" "test_c4" {
	  remote_network_id = twingate_remote_network.test_c4_2.id
	}
	`, remoteNetworkName)
}

func TestAccTwingateConnectorReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Connector ReCreate After Deletion", func(t *testing.T) {
		remoteNetworkName := getRandomName()
		const theResource = "twingate_connector.test_c5"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: createConnectorC5(remoteNetworkName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(theResource, "twingate_remote_network.test_c5"),
						deleteTwingateResource(theResource, connectorResourceName),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: createConnectorC5(remoteNetworkName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(theResource, "twingate_remote_network.test_c5"),
					),
				},
			},
		})
	})
}

func createConnectorC5(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_c5" {
	  name = "%s"
	}
	resource "twingate_connector" "test_c5" {
	  remote_network_id = twingate_remote_network.test_c5.id
	}
	`, remoteNetworkName)
}
