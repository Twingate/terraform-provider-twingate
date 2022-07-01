package twingate

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateGroups_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups Basic", func(t *testing.T) {

		groupName := acctest.RandomWithPrefix(testPrefixName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroups(groupName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_group", groupName),
						resource.TestCheckResourceAttr("data.twingate_groups.out", "groups.#", "2"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroups(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test1" {
	  name = "%s"
	}

	resource "twingate_group" "test2" {
	  name = "%s"
	}

	data "twingate_groups" "out" {
	  name = "%s"

	  depends_on = [twingate_group.test1, twingate_group.test2]
	}

	output "my_group" {
	  value = data.twingate_groups.out.groups[0].name
	}
	`, name, name, name)
}

func TestAccDatasourceTwingateGroups_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups - does not exists", func(t *testing.T) {
		groupName := acctest.RandomWithPrefix(testPrefixName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateGroupsDoesNotExists(groupName),
					ExpectError: regexp.MustCompile("Error: failed to read group with name"),
				},
			},
		})
	})
}

func testTwingateGroupsDoesNotExists(name string) string {
	return fmt.Sprintf(`
	data "twingate_groups" "test" {
	  name = "%s"
	}

	output "my_groups" {
	  value = data.twingate_groups.test.groups
	}
	`, name)
}

func TestAccDatasourceTwingateGroups_emptyGroupName(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups - failed parse group name", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateGroupsDoesNotExists(""),
					ExpectError: regexp.MustCompile("Error: failed to read group: group name is empty"),
				},
			},
		})
	})
}
