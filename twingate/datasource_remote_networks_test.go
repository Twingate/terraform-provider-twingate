package twingate

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateRemoteNetworks_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Networks Basic", func(t *testing.T) {

		networkName := acctest.RandomWithPrefix("tf-acc")

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateRemoteNetworks(networkName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_network", networkName),
						resource.TestCheckResourceAttr("data.twingate_remote_networks.out", "remote_networks.#", "2"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateRemoteNetworks(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test1" {
	  name = "%s"
	}

	resource "twingate_remote_network" "test2" {
	  name = "%s"
	}

	data "twingate_remote_networks" "out" {
	  name = "%s"

	  depends_on = [twingate_remote_network.test1, twingate_remote_network.test2]
	}

	output "my_network" {
	  value = data.twingate_remote_networks.out.remote_networks[0].name
	}
	`, name, name, name)
}

func TestAccDatasourceTwingateRemoteNetworks_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Networks - does not exists", func(t *testing.T) {
		networkName := acctest.RandomWithPrefix("tf-acc")

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateRemoteNetworksDoesNotExists(networkName),
					ExpectError: regexp.MustCompile("Error: failed to read remote network with name"),
				},
			},
		})
	})
}

func testTwingateRemoteNetworksDoesNotExists(name string) string {
	return fmt.Sprintf(`
	data "twingate_remote_networks" "test" {
	  name = "%s"
	}

	output "my_network" {
	  value = data.twingate_remote_networks.test.remote_networks
	}
	`, name)
}

func TestAccDatasourceTwingateRemoteNetworks_emptyNetworkName(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Networks - failed parse network name", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateRemoteNetworksDoesNotExists(""),
					ExpectError: regexp.MustCompile("Error: failed to read remote network: network name is empty"),
				},
			},
		})
	})
}
