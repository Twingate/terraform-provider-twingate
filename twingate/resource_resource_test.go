package twingate

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTwingateResource_basic(t *testing.T) {
	remoteNetworkName := acctest.RandomWithPrefix("tf-acc")
	resourceName := acctest.RandomWithPrefix("tf-acc-resource")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test"),
					resource.TestCheckNoResourceAttr("twingate_resource.test", "group_ids.#"),
				),
			},
			{
				Config: testTwingateResource_withProtocolsAndGroups(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test"),
					resource.TestCheckResourceAttr("twingate_resource.test", "address", "updated-acc-test.com"),
					resource.TestCheckResourceAttr("twingate_resource.test", "group_ids.#", "2"),
					resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.tcp.0.policy", "RESTRICTED"),
					resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.tcp.0.ports.0", "80"),
				),
			},
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test"),
					resource.TestCheckNoResourceAttr("twingate_resource.test", "group_ids.#"),
					resource.TestCheckNoResourceAttr("twingate_resource.test", "protocols.0.tcp.0.ports.0"),
				),
			},
		})
	})
}

func testTwingateResource_Simple(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
	}
	`, networkName, resourceName)
}

func testTwingateResource_withProtocolsAndGroups(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "updated-acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
	  group_ids = ["R3JvdXA6MjMxNTQ=", "R3JvdXA6MTk0MjQ="]
      protocols {
		allow_icmp = true
        tcp  {
			policy = "RESTRICTED"
            ports = ["80", "82-83"]
        }
		udp {
 			policy = "ALLOW_ALL"
		}
      }
	}
	`, networkName, resourceName)
}

// adding test when policy is restricted and ports are empty list

func testAccCheckTwingateResourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_resource" {
			continue
		}

		resourceId := rs.Primary.ID

		err := client.deleteResource(resourceId)
		// expecting error here , since the resource is already gone
		if err == nil {
			return fmt.Errorf("resource with ID %s still present : ", resourceId)
		}
	}

	return nil
}

func testAccCheckTwingateResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ResourceId set ")
		}

		return nil
	}
}
