package twingate

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateConnector_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Connector Basic", func(t *testing.T) {
		networkName := getRandomName()
		connectorName := getRandomConnectorName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateConnector(networkName, connectorName),
					Check: ComposeTestCheckFunc(
						resource.TestCheckOutput("my_connector", connectorName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateConnector(remoteNetworkName, connectorName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_dc1" {
	  name = "%s"
	}
	resource "twingate_connector" "test_dc1" {
	  remote_network_id = twingate_remote_network.test_dc1.id
	  name  = "%s"
	}

	data "twingate_connector" "out_dc1" {
	  id = twingate_connector.test_dc1.id
	}

	output "my_connector" {
	  value = data.twingate_connector.out_dc1.name
	}
	`, remoteNetworkName, connectorName)
}

func TestAccDatasourceTwingateConnector_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Connector - does not exists", func(t *testing.T) {
		connectorID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Connector:%d", acctest.RandInt())))

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateConnectorDoesNotExists(connectorID),
					ExpectError: regexp.MustCompile("Error: failed to read connector with id"),
				},
			},
		})
	})
}

func testTwingateConnectorDoesNotExists(id string) string {
	return fmt.Sprintf(`
	data "twingate_connector" "test_dc2" {
	  id = "%s"
	}

	output "my_connector" {
	  value = data.twingate_connector.test_dc2.name
	}
	`, id)
}

func TestAccDatasourceTwingateConnector_invalidID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Connector - failed parse ID", func(t *testing.T) {
		connectorID := acctest.RandString(10)

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateConnectorDoesNotExists(connectorID),
					ExpectError: regexp.MustCompile("Unable to parse global ID"),
				},
			},
		})
	})
}
