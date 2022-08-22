package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	connectorsDatasource = "data.twingate_connectors.all"
	connectorsNumber     = "connectors.#"
	firstConnectorName   = "connectors.0.name"
)

func TestAccDatasourceTwingateConnectors_basic(t *testing.T) {
	t.Parallel()
	t.SkipNow() // fixed in other PR
	t.Run("Test Twingate Datasource : Acc Connectors Basic", func(t *testing.T) {
		networkName1 := test.RandomName()
		networkName2 := test.RandomName()
		connectorName := test.RandomConnectorName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateConnectorDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateConnectors(networkName1, connectorName, networkName2, connectorName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(connectorsDatasource, connectorsNumber, "2"),
						resource.TestCheckResourceAttr(connectorsDatasource, firstConnectorName, connectorName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateConnectors(networkName1, connectorName1, networkName2, connectorName2 string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test1" {
		name = "%s"
	}
	resource "twingate_connector" "test1" {
		remote_network_id = twingate_remote_network.test1.id
		name = "%s"
	}
	resource "twingate_remote_network" "test2" {
		name = "%s"
	}
	resource "twingate_connector" "test2" {
		remote_network_id = twingate_remote_network.test2.id
		name = "%s"
	}
	data "twingate_connectors" "all" {
		depends_on = [twingate_connector.test1, twingate_connector.test2]
	}
		`, networkName1, connectorName1, networkName2, connectorName2)
}

func TestAccDatasourceTwingateConnectors_emptyResult(t *testing.T) {
	t.Parallel()
	t.SkipNow() // fixed in other PR
	t.Run("Test Twingate Datasource : Acc Connectors - empty result", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testTwingateConnectorsDoesNotExists(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(connectorsDatasource, connectorsNumber, "0"),
					),
				},
			},
		})
	})
}

func testTwingateConnectorsDoesNotExists() string {
	return `
		data "twingate_connectors" "all" {}
	`
}
