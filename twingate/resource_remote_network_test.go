package twingate

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccTwingateRemoteNetworkCreateUpdate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Network Basic", func(t *testing.T) {

		remoteNetworkNameBefore := getRandomName()
		remoteNetworkNameAfter := getRandomName()

		const theResource = "twingate_remote_network.test001"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config: createRemoteNetwork001(remoteNetworkNameBefore),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateRemoteNetworkExists(theResource),
						resource.TestCheckResourceAttr(theResource, "name", remoteNetworkNameBefore),
					),
				},
				{
					Config: createRemoteNetwork001(remoteNetworkNameAfter),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateRemoteNetworkExists(theResource),
						resource.TestCheckResourceAttr(theResource, "name", remoteNetworkNameAfter),
					),
				},
			},
		})
	})
}

func createRemoteNetwork001(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test001" {
	  name = "%s"
	}
	`, name)
}

func TestAccTwingateRemoteNetworkDeleteNonExisting(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Network Delete NonExisting", func(t *testing.T) {

		remoteNetworkNameBefore := getRandomName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config:  createRemoteNetwork002(remoteNetworkNameBefore),
					Destroy: true,
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateRemoteNetworkDoesNotExists("twingate_remote_network.test002"),
					),
				},
			},
		})
	})
}

func createRemoteNetwork002(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test002" {
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

		err := client.deleteRemoteNetwork(context.Background(), remoteNetworkId)
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

func TestAccTwingateRemoteNetworkReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Remote Network Re Create After Deletion", func(t *testing.T) {

		remoteNetworkName := getRandomName()

		const theResource = "twingate_remote_network.test003"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateRemoteNetworkDestroy,
			Steps: []resource.TestStep{
				{
					Config: createRemoteNetwork003(remoteNetworkName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateRemoteNetworkExists(theResource),
						deleteTwingateResource(theResource, remoteNetworkResourceName),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: createRemoteNetwork003(remoteNetworkName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateRemoteNetworkExists(theResource),
					),
				},
			},
		})
	})
}

func createRemoteNetwork003(name string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test003" {
	  name = "%s"
	}
	`, name)
}
