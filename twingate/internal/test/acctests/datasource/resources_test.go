package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	resourcesLen     = attr.Len(attr.Resources)
	resourceNamePath = attr.Path(attr.Resources, attr.Name)
)

func TestAccDatasourceTwingateResources_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resources Basic", func(t *testing.T) {
		acctests.SetPageLimit(10)
		networkName := test.RandomName()
		resourceName := test.RandomResourceName()
		const theDatasource = "data.twingate_resources.out_drs1"

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateResources(networkName, resourceName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(theDatasource, resourcesLen, "2"),
						resource.TestCheckResourceAttr(theDatasource, resourceNamePath, resourceName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateResources(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_drs1" {
	  name = "%s"
	}

	resource "twingate_resource" "test_drs1_1" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test_drs1.id
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "RESTRICTED"
	      ports = ["80-83", "85"]
	    }
	    udp {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	  }
	}

	resource "twingate_resource" "test_drs1_2" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test_drs1.id
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	    udp {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	  }
	}

	data "twingate_resources" "out_drs1" {
	  name = "%s"

	  depends_on = [twingate_resource.test_drs1_1, twingate_resource.test_drs1_2]
	}
	`, networkName, resourceName, resourceName, resourceName)
}

func TestAccDatasourceTwingateResources_emptyResult(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resources - empty result", func(t *testing.T) {
		resourceName := test.RandomResourceName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config: testTwingateResourcesDoesNotExists(resourceName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_resources.out_drs2", resourcesLen, "0"),
					),
				},
			},
		})
	})
}

func testTwingateResourcesDoesNotExists(name string) string {
	return fmt.Sprintf(`
	data "twingate_resources" "out_drs2" {
	  name = "%s"
	}

	output "my_resources_drs2" {
	  value = data.twingate_resources.out_drs2.resources
	}
	`, name)
}
