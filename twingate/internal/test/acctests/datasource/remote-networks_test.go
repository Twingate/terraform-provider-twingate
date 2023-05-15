package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateRemoteNetworks_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Networks Basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateRemoteNetworks(),
					Check: acctests.ComposeTestCheckFunc(
						testCheckResourceAttrNotEqual("data.twingate_remote_networks.all", attr.Len(attr.RemoteNetworks), "0"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateRemoteNetworks() string {
	return `
	data "twingate_remote_networks" "all" {}
	`
}

func TestAccDatasourceTwingateRemoteNetworks_read(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Remote Networks Read", func(t *testing.T) {
		prefix := acctest.RandString(10)
		networkName1 := test.RandomName(prefix)
		networkName2 := test.RandomName(prefix)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateRemoteNetworks2(networkName1, networkName2, prefix),
					Check: acctests.ComposeTestCheckFunc(
						testCheckOutputLength("test_networks", 2),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateRemoteNetworks2(networkName1, networkName2, prefix string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_drn1" {
		name = "%s"
	}
	
	resource "twingate_remote_network" "test_drn2" {
		name = "%s"
	}
	
	data "twingate_remote_networks" "all" {
		depends_on = [twingate_remote_network.test_drn1, twingate_remote_network.test_drn2]
	}

	output "test_networks" {
	  	value = [for n in [for net in data.twingate_remote_networks.all : net if can(net.*.name)][0] : n if length(regexall("%s.*", n.name)) > 0]
	}
		`, networkName1, networkName2, prefix)
}
