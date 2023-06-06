package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateRemoteNetwork_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Network Basic", func(t *testing.T) {

		networkName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateRemoteNetwork(networkName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_remote_network.test_dn1_2", attr.Name, networkName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateRemoteNetwork(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_dn1_1" {
	  name = "%s"
	}

	data "twingate_remote_network" "test_dn1_2" {
	  id = twingate_remote_network.test_dn1_1.id
	}

	output "my_network_dn1_" {
	  value = data.twingate_remote_network.test_dn1_2.name
	}
	`, name)
}

func TestAccDatasourceTwingateRemoteNetworkByName_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Network Basic", func(t *testing.T) {

		networkName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateRemoteNetworkByName(networkName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_remote_network.test_dn2_2", attr.Name, networkName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateRemoteNetworkByName(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_dn2_1" {
	  name = "%s"
	}

	data "twingate_remote_network" "test_dn2_2" {
	  name = "%s"
	  depends_on = [resource.twingate_remote_network.test_dn2_1]
	}

	output "my_network_dn2" {
	  value = data.twingate_remote_network.test_dn2_2.name
	}
	`, name, name)
}

func TestAccDatasourceTwingateRemoteNetwork_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Network - does not exists", func(t *testing.T) {
		networkID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("RemoteNetwork:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config:      testTwingateRemoteNetworkDoesNotExists(networkID),
					ExpectError: regexp.MustCompile("failed to read remote network with id"),
				},
			},
		})
	})
}

func testTwingateRemoteNetworkDoesNotExists(id string) string {
	return fmt.Sprintf(`
	data "twingate_remote_network" "test_dn3" {
	  id = "%s"
	}

	output "my_network_dn3" {
	  value = data.twingate_remote_network.test_dn3.name
	}
	`, id)
}

func TestAccDatasourceTwingateRemoteNetwork_invalidNetworkID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Network - failed parse network ID", func(t *testing.T) {
		networkID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config:      testTwingateRemoteNetworkDoesNotExists(networkID),
					ExpectError: regexp.MustCompile("failed to read remote network with id"),
				},
			},
		})
	})
}

func TestAccDatasourceTwingateRemoteNetwork_bothNetworkIDAndName(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Network - failed passing both network ID and name", func(t *testing.T) {
		networkID := acctest.RandString(10)
		networkName := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateRemoteNetworkValidationFailed(networkID, networkName),
					ExpectError: regexp.MustCompile("invalid combination of arguments"),
				},
			},
		})
	})
}

func testTwingateRemoteNetworkValidationFailed(id, name string) string {
	return fmt.Sprintf(`
	data "twingate_remote_network" "test_dn4" {
	  id = "%s"
	  name = "%s"
	}

	output "my_network_dn4" {
	  value = data.twingate_remote_network.test_dn4.name
	}
	`, id, name)
}
