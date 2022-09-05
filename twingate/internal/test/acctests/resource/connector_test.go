package resource

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	provider_resource "github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	connectorResourceName     = "connector"
	connectorResource         = "twingate_connector.test"
	remoteNetworkResource     = "twingate_remote_network.test"
	remoteNetworkResourceName = "remote network"
	nameAttr                  = "name"
)

var testRegexp = regexp.MustCompile(test.Prefix() + ".*")

func TestAccRemoteConnector_basic(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Remote Connector", func(t *testing.T) {
		remoteNetworkName := test.RandomName()
		connectorName := test.RandomConnectorName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnector(remoteNetworkName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
						resource.TestCheckResourceAttrSet(connectorResource, nameAttr),
					),
				},
				{
					Config: testTwingateConnectorWithCustomName(remoteNetworkName, connectorName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
						resource.TestMatchResourceAttr(connectorResource, nameAttr, testRegexp),
					),
				},
			},
		})
	})
}

func TestAccRemoteConnector_withName(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Remote Connector", func(t *testing.T) {
		remoteNetworkName := test.RandomName()
		connectorName := test.RandomConnectorName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnectorWithCustomName(remoteNetworkName, connectorName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
						resource.TestMatchResourceAttr(connectorResource, nameAttr, testRegexp),
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

func testTwingateConnectorWithAnotherNetwork(remoteNetworkName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test1" {
	  name = "%s"
	}
	resource "twingate_connector" "test" {
	  remote_network_id = twingate_remote_network.test1.id
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
	client := acctests.Provider.Meta().(*transport.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_connector" {
			continue
		}

		connectorId := rs.Primary.ID

		err := client.DeleteConnector(context.Background(), connectorId)
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

func TestAccRemoteConnector_import(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Remote Connector - Import", func(t *testing.T) {
		remoteNetworkName := test.RandomName()
		connectorName := test.RandomConnectorName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnectorWithCustomName(remoteNetworkName, connectorName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
						resource.TestMatchResourceAttr(connectorResource, nameAttr, testRegexp),
					),
				},
				{
					ResourceName:      connectorResource,
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccRemoteConnector_notAllowedToChangeRemoteNetworkId(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Remote Connector - should fail on remote_network_id update", func(t *testing.T) {
		remoteNetworkName := test.RandomName()
		remoteNetworkName1 := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnector(remoteNetworkName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
					),
				},
				{
					Config:      testTwingateConnectorWithAnotherNetwork(remoteNetworkName1),
					ExpectError: regexp.MustCompile(provider_resource.ErrNotAllowChangeRemoteNetworkID.Error()),
				},
			},
		})
	})
}

func TestAccTwingateConnector_createAfterDeletion(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Remote Connector Create After Deletion", func(t *testing.T) {
		remoteNetworkName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnector(remoteNetworkName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
						deleteTwingateResource(connectorResource, connectorResourceName),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: testTwingateConnector(remoteNetworkName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateConnectorExists(connectorResource, remoteNetworkResource),
					),
				},
			},
		})
	})
}