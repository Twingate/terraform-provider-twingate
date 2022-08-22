package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	groupResourceName = "group"
	groupResource     = "twingate_group.test"
)

func TestAccTwingateGroup_basic(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Group Basic", func(t *testing.T) {

		groupNameBefore := test.RandomName()
		groupNameAfter := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testTwingateGroup(groupNameBefore),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(groupResource),
						resource.TestCheckResourceAttr(groupResource, nameAttr, groupNameBefore),
					),
				},
				{
					Config: testTwingateGroup(groupNameAfter),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(groupResource),
						resource.TestCheckResourceAttr(groupResource, nameAttr, groupNameAfter),
					),
				},
			},
		})
	})
}

func TestAccTwingateGroup_deleteNonExisting(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Group Delete NonExisting", func(t *testing.T) {

		groupNameBefore := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config:  testTwingateGroup(groupNameBefore),
					Destroy: true,
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateGroupDoesNotExists(groupResource),
					),
				},
			},
		})
	})
}

func testTwingateGroup(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test" {
	  name = "%s"
	}
	`, name)
}

func testAccCheckTwingateGroupDestroy(s *terraform.State) error {
	client := acctests.Provider.Meta().(*transport.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_group" {
			continue
		}

		groupId := rs.Primary.ID

		err := client.DeleteGroup(context.Background(), groupId)
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

func TestAccTwingateGroup_createAfterDeletion(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Acc Group Create After Deletion", func(t *testing.T) {
		groupName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testTwingateGroup(groupName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(groupResource),
						deleteTwingateResource(groupResource, groupResourceName),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: testTwingateGroup(groupName),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(groupResource),
					),
				},
			},
		})
	})
}
