package twingate

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateResources_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resources Basic", func(t *testing.T) {

		networkName := acctest.RandomWithPrefix(testPrefixName)
		resourceName := acctest.RandomWithPrefix(testPrefixName + "-resource")

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateResources(networkName, resourceName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_resource_name", resourceName),
						resource.TestCheckResourceAttr("data.twingate_resources.out", "resources.#", "2"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateResources(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}

	resource "twingate_resource" "test1" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
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

	resource "twingate_resource" "test2" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
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

	data "twingate_resources" "out" {
	  name = "%s"

	  depends_on = [twingate_resource.test1, twingate_resource.test2]
	}

	output "my_resource_name" {
	  value = data.twingate_resources.out.resources[0].name
	}
	`, networkName, resourceName, resourceName, resourceName)
}

func TestAccDatasourceTwingateResources_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resources - does not exists", func(t *testing.T) {
		resourceName := acctest.RandomWithPrefix(testPrefixName + "-resource")

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateResourcesDoesNotExists(resourceName),
					ExpectError: regexp.MustCompile("Error: failed to read resource with id All"),
				},
			},
		})
	})
}

func testTwingateResourcesDoesNotExists(name string) string {
	return fmt.Sprintf(`
	data "twingate_resources" "test" {
	  name = "%s"
	}

	output "my_resources" {
	  value = data.twingate_resources.test.resources
	}
	`, name)
}

func TestAccDatasourceTwingateResources_emptyResourceName(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resources - failed parse resource name", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateResourcesDoesNotExists(""),
					ExpectError: regexp.MustCompile("Error: failed to read resource with id All: not found"),
				},
			},
		})
	})
}
