package twingate

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateResource_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resource Basic", func(t *testing.T) {
		networkName := getRandomName()
		resourceName := getRandomResourceName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateResource(networkName, resourceName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_resource.out_dr1", "name", resourceName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateResource(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test_dr1" {
	  name = "%s"
	}

	resource "twingate_resource" "test_dr1" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test_dr1.id
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

	data "twingate_resource" "out_dr1" {
	  id = twingate_resource.test_dr1.id
	}
	`, networkName, resourceName)
}

func TestAccDatasourceTwingateResource_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resource - does not exists", func(t *testing.T) {
		resourceID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Resource:%d", acctest.RandInt())))

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateResourceDoesNotExists(resourceID),
					ExpectError: regexp.MustCompile("Error: failed to read resource with id"),
				},
			},
		})
	})
}

func testTwingateResourceDoesNotExists(id string) string {
	return fmt.Sprintf(`
	data "twingate_resource" "test_dr2" {
	  id = "%s"
	}

	output "my_resource_dr2" {
	  value = data.twingate_resource.test_dr2.name
	}
	`, id)
}

func TestAccDatasourceTwingateResource_invalidID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Resource - failed parse resource ID", func(t *testing.T) {
		networkID := acctest.RandString(10)

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateResourceDoesNotExists(networkID),
					ExpectError: regexp.MustCompile("Unable to parse global ID"),
				},
			},
		})
	})
}