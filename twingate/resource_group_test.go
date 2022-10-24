package twingate

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const groupResource = "twingate_group.test"

func TestAccTwingateGroupCreateUpdate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Create/Update", func(t *testing.T) {

		groupNameBefore := getRandomName()
		groupNameAfter := getRandomName()

		const theResource = "twingate_group.test001"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: createGroup001(groupNameBefore),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
						resource.TestCheckResourceAttr(theResource, "name", groupNameBefore),
					),
				},
				{
					Config: createGroup001(groupNameAfter),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
						resource.TestCheckResourceAttr(theResource, "name", groupNameAfter),
					),
				},
			},
		})
	})
}

func createGroup001(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test001" {
	  name = "%s"
	}
	`, name)
}

func TestAccTwingateGroupDeleteNonExisting(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Delete NonExisting", func(t *testing.T) {

		groupNameBefore := getRandomName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config:  createGroup002(groupNameBefore),
					Destroy: true,
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupDoesNotExists("twingate_group.test002"),
					),
				},
			},
		})
	})
}

func createGroup002(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test002" {
	  name = "%s"
	}
	`, name)
}

func testAccCheckTwingateGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_group" {
			continue
		}

		groupId := rs.Primary.ID

		err := client.deleteGroup(context.Background(), groupId)
		if err == nil {
			return fmt.Errorf("Group with ID %s still present : ", groupId)
		}
	}

	return nil
}

func testAccCheckTwingateGroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GroupID set ")
		}

		return nil
	}
}

func testAccCheckTwingateGroupDoesNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		_ = rs
		if !ok {
			return nil
		}

		return fmt.Errorf("this resource should not be here: %s ", resourceName)
	}
}

func TestAccTwingateGroupReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Create After Deletion", func(t *testing.T) {
		groupName := getRandomName()

		const theResource = "twingate_group.test003"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: createGroup003(groupName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
						deleteTwingateResource(theResource, groupResourceName),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: createGroup003(groupName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
					),
				},
			},
		})
	})
}

func createGroup003(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test003" {
	  name = "%s"
	}
	`, name)
}
