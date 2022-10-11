package twingate

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateGroup_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group Basic", func(t *testing.T) {
		groupName := getRandomName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroup(groupName),
					Check: resource.ComposeTestCheckFunc(
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

func TestAccDatasourceTwingateGroup_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group - does not exists", func(t *testing.T) {
		groupID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Group:%d", acctest.RandInt())))

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
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

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
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
