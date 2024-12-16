package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateConnector_basic(t *testing.T) {
	t.Parallel()

	networkName := test.RandomName()
	connectorName := test.RandomConnectorName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateConnector(networkName, connectorName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("my_connector", connectorName),
					resource.TestCheckOutput("my_connector_notification_status", "true"),
					resource.TestCheckOutput("my_connector_state", "DEAD_NO_HEARTBEAT"),
					resource.TestCheckOutput("my_connector_version", ""),
					resource.TestCheckOutput("my_connector_hostname", ""),
					resource.TestCheckOutput("my_connector_public_ip", ""),
				),
			},
		},
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

	output "my_connector_notification_status" {
	  value = data.twingate_connector.out_dc1.status_updates_enabled
	}

	output "my_connector_state" {
	  value = data.twingate_connector.out_dc1.state
	}

	output "my_connector_version" {
	  value = data.twingate_connector.out_dc1.version
	}

	output "my_connector_hostname" {
	  value = data.twingate_connector.out_dc1.hostname
	}

	output "my_connector_public_ip" {
	  value = data.twingate_connector.out_dc1.public_ip
	}

	`, remoteNetworkName, connectorName)
}

func TestAccDatasourceTwingateConnector_negative(t *testing.T) {
	t.Parallel()

	connectorID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Connector:%d", acctest.RandInt())))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testTwingateConnectorDoesNotExists(connectorID),
				ExpectError: regexp.MustCompile("failed to read connector with id"),
			},
		},
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
	t.Parallel()

	connectorID := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testTwingateConnectorDoesNotExists(connectorID),
				ExpectError: regexp.MustCompile("failed to read connector with id"),
			},
		},
	})
}
