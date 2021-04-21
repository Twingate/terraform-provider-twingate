package twingate

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTwingateRemoteNetwork_basic(t *testing.T) {

	remoteNetworkNameBefore := acctest.RandomWithPrefix("tf-acc")
	remoteNetworkNameAfter := acctest.RandomWithPrefix("tf-acc")
	resourceName := "twingate_remote_network.test"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateRemoteNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateRemoteNetwork(remoteNetworkNameBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateRemoteNetworkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", remoteNetworkNameBefore),
				),
			},
			{
				Config: testTwingateRemoteNetwork(remoteNetworkNameAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateRemoteNetworkExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", remoteNetworkNameAfter),
				),
			},
		},
	})
}

func TestAccTwingateRemoteNetwork_deleteNonExisting(t *testing.T) {

	remoteNetworkNameBefore := acctest.RandomWithPrefix("tf-acc")
	resourceName := "twingate_remote_network.test"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateRemoteNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testTwingateRemoteNetwork(remoteNetworkNameBefore),
				Destroy: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateRemoteNetworkDoesNotExists(resourceName),
				),
			},
		},
	})
}

func testTwingateRemoteNetwork(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	`, name)
}

func testAccCheckTwingateRemoteNetworkDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_remote_network" {
			continue
		}

		remoteNetworkId := rs.Primary.ID

		err := client.deleteRemoteNetwork(&remoteNetworkId)
		// expecting error here , since the network is already gone
		if err == nil {
			return fmt.Errorf("Remote network with ID %s still present : ", remoteNetworkId)
		}
	}

	return nil
}

func testAccCheckTwingateRemoteNetworkExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RemoteNetworkID set ")
		}

		return nil
	}
}

func testAccCheckTwingateRemoteNetworkDoesNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		_ = rs
		if !ok {
			return nil
		}

		return fmt.Errorf("this resource should not be here: %s ", resourceName)
	}
}
