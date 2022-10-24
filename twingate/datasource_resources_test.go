package twingate

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	resourcesDatasource = "data.twingate_resources.out"
	resourcesNumber     = "resources.#"
	firstResourceName   = "resources.0.name"
)

func TestAccDatasourceTwingateResources_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resources Basic", func(t *testing.T) {

		networkName := getRandomName()
		resourceName := getRandomResourceName()
		const theDatasource = "data.twingate_resources.out_drs1"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateResources(networkName, resourceName),
					Check: ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(theDatasource, resourcesNumber, "2"),
						resource.TestCheckResourceAttr(theDatasource, firstResourceName, resourceName),
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
		resourceName := getRandomResourceName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config: testTwingateResourcesDoesNotExists(resourceName),
					Check: ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_resources.out_drs2", resourcesNumber, "0"),
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
