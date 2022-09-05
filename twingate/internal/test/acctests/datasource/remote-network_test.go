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

func TestAccDatasourceTwingateRemoteNetwork_basic(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Remote Network Basic", func(t *testing.T) {

		networkName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateRemoteNetwork(networkName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_remote_network.test2", "name", networkName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateRemoteNetwork(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test1" {
	  name = "%s"
	}

	data "twingate_remote_network" "test2" {
	  id = twingate_remote_network.test1.id
	}

	output "my_network" {
	  value = data.twingate_remote_network.test2.name
	}
	`, name)
}

func testAccCheckTwingateRemoteNetworkDestroy(s *terraform.State) error {
	client := acctests.Provider.Meta().(*transport.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_remote_network" {
			continue
		}

		remoteNetworkId := rs.Primary.ID

		err := client.DeleteRemoteNetwork(context.Background(), remoteNetworkId)
		// expecting error here , since the network is already gone
		if err == nil {
			return fmt.Errorf("Remote network with ID %s still present : ", remoteNetworkId)
		}
	}

	return nil
}

func TestAccDatasourceTwingateRemoteNetworkByName_basic(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Remote Network Basic", func(t *testing.T) {

		networkName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateRemoteNetworkByName(networkName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_remote_network.test2", "name", networkName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateRemoteNetworkByName(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test1" {
	  name = "%s"
	}

	data "twingate_remote_network" "test2" {
	  name = "%s"
	  depends_on = [resource.twingate_remote_network.test1]
	}

	output "my_network" {
	  value = data.twingate_remote_network.test2.name
	}
	`, name, name)
}

func TestAccDatasourceTwingateRemoteNetwork_negative(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Remote Network - does not exists", func(t *testing.T) {
		networkID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("RemoteNetwork:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateRemoteNetworkDoesNotExists(networkID),
					ExpectError: regexp.MustCompile("Error: failed to read remote network with id"),
				},
			},
		})
	})
}

func testTwingateRemoteNetworkDoesNotExists(id string) string {
	return fmt.Sprintf(`
	data "twingate_remote_network" "test" {
	  id = "%s"
	}

	output "my_network" {
	  value = data.twingate_remote_network.test.name
	}
	`, id)
}

func TestAccDatasourceTwingateRemoteNetwork_invalidNetworkID(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Remote Network - failed parse network ID", func(t *testing.T) {
		networkID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateRemoteNetworkDoesNotExists(networkID),
					ExpectError: regexp.MustCompile("Unable to parse global ID"),
				},
			},
		})
	})
}

func TestAccDatasourceTwingateRemoteNetwork_bothNetworkIDAndName(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Remote Network - failed passing both network ID and name", func(t *testing.T) {
		networkID := acctest.RandString(10)
		networkName := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateRemoteNetworkValidationFailed(networkID, networkName),
					ExpectError: regexp.MustCompile("Invalid combination of arguments"),
				},
			},
		})
	})
}

func testTwingateRemoteNetworkValidationFailed(id, name string) string {
	return fmt.Sprintf(`
	data "twingate_remote_network" "test" {
	  id = "%s"
	  name = "%s"
	}

	output "my_network" {
	  value = data.twingate_remote_network.test.name
	}
	`, id, name)
}