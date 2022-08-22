package datasource

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatasourceTwingateConnector_basic(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Connector Basic", func(t *testing.T) {
		networkName := test.RandomName()
		connectorName := test.RandomConnectorName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateConnector(networkName, connectorName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_connector", connectorName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateConnector(remoteNetworkName, connectorName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_connector" "test" {
	  remote_network_id = twingate_remote_network.test.id
	  name  = "%s"
	}

	data "twingate_connector" "out" {
	  id = twingate_connector.test.id
	}

	output "my_connector" {
	  value = data.twingate_connector.out.name
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

func TestAccDatasourceTwingateConnector_negative(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Connector - does not exists", func(t *testing.T) {
		connectorID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Connector:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
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
	data "twingate_connector" "test" {
	  id = "%s"
	}

	output "my_connector" {
	  value = data.twingate_connector.test.name
	}
	`, id)
}

func TestAccDatasourceTwingateConnector_invalidID(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Connector - failed parse ID", func(t *testing.T) {
		connectorID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
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
