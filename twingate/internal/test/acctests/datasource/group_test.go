package datasource

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatasourceTwingateGroup_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group Basic", func(t *testing.T) {
		groupName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroup(groupName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_group_dg1", groupName),
						resource.TestCheckOutput("my_group_is_active_dg1", "true"),
						resource.TestCheckOutput("my_group_type_dg1", "MANUAL"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroup(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "foo_dg1" {
	  name = "%s"
	}

	data "twingate_group" "bar_dg1" {
	  id = twingate_group.foo_dg1.id
	}

	output "my_group_dg1" {
	  value = data.twingate_group.bar_dg1.name
	}

	output "my_group_is_active_dg1" {
	  value = data.twingate_group.bar_dg1.is_active
	}

	output "my_group_type_dg1" {
	  value = data.twingate_group.bar_dg1.type
	}
	`, name)
}

func testAccCheckTwingateGroupDestroy(s *terraform.State) error {
	c := acctests.Provider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_group" {
			continue
		}

		groupId := rs.Primary.ID

		err := c.DeleteGroup(context.Background(), groupId)
		if err == nil {
			return fmt.Errorf("Group with ID %s still present : ", groupId)
		}
	}

	return nil
}

func TestAccDatasourceTwingateGroup_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group - does not exists", func(t *testing.T) {
		groupID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Group:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateGroupDoesNotExists(groupID),
					ExpectError: regexp.MustCompile("Error: failed to read group with id"),
				},
			},
		})
	})
}

func testTwingateGroupDoesNotExists(id string) string {
	return fmt.Sprintf(`
	data "twingate_group" "foo_dg2" {
	  id = "%s"
	}
	`, id)
}

func TestAccDatasourceTwingateGroup_invalidGroupID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group - failed parse group ID", func(t *testing.T) {
		groupID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateGroupDoesNotExists(groupID),
					ExpectError: regexp.MustCompile("Unable to parse global ID"),
				},
			},
		})
	})
}
