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

func TestAccDatasourceTwingateGroups_emptyResult(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups - empty result", func(t *testing.T) {
		groupName := acctest.RandomWithPrefix(testPrefixName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config: testTwingateGroupsDoesNotExists(groupName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_groups.test", "groups.#", "0"),
					),
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

func TestAccDatasourceTwingateGroupsWithFilters_basic(t *testing.T) {
	groupName := acctest.RandomWithPrefix(testPrefixName)

	t.Run("Test Twingate Datasource : Acc Groups with filters - basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroupsWithFilters(groupName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_group", groupName),
						resource.TestCheckResourceAttr("data.twingate_groups.out", "groups.#", "2"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroupsWithFilters(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test1" {
	  name = "%s"
	}

	resource "twingate_group" "test2" {
	  name = "%s"
	}

	data "twingate_groups" "out" {
	  name = "%s"
	  type = "MANUAL"
	  is_active = true

	  depends_on = [twingate_group.test1, twingate_group.test2]
	}

	output "my_group" {
	  value = data.twingate_groups.out.groups[0].name
	}
	`, name, name, name)
}

func TestAccDatasourceTwingateGroupsWithFilters_ErrorNotSupportedTypes(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups with filters - error not supported types", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateGroupsWithFilterNotSupportedType(),
					ExpectError: regexp.MustCompile("Error: expected type to be one of"),
				},
			},
		})
	})
}

func testTwingateGroupsWithFilterNotSupportedType() string {
	return `
	data "twingate_groups" "test" {
	  type = "OTHER"
	}

	output "my_groups" {
	  value = data.twingate_groups.test.groups
	}
	`
}

func TestAccDatasourceTwingateGroups_WithEmptyFilters(t *testing.T) {

	t.Run("Test Twingate Datasource : Acc Groups - with empty filters", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config: testTwingateGroupsWithEmptyFilter(),
				},
			},
		})
	})
}

func testTwingateGroupsWithEmptyFilter() string {
	return `
	data "twingate_groups" "all" {}

	output "my_groups" {
	  value = data.twingate_groups.all.groups
	}
	`
}

func TestAccDatasourceTwingateGroups_withTwoDatasource(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups with two datasource", func(t *testing.T) {

		groupName := acctest.RandomWithPrefix(testPrefixName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroupsWithDatasource(groupName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_group", groupName),
						resource.TestCheckResourceAttr("data.twingate_groups.out", "groups.#", "2"),
						resource.TestCheckResourceAttr("data.twingate_groups.all", "groups.#", "2"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroupsWithDatasource(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test1" {
	  name = "%s"
	}

	resource "twingate_group" "test2" {
	  name = "%s"
	}

	data "twingate_groups" "all" {
	  name = "%s"
	  is_active = true
	  type = "MANUAL"

	  depends_on = [twingate_group.test1, twingate_group.test2]
	}

	data "twingate_groups" "out" {
	  name = "%s"

	  depends_on = [twingate_group.test1, twingate_group.test2]
	}

	output "my_group" {
	  value = data.twingate_groups.out.groups[0].name
	}

	output "all_groups" {
	  value = data.twingate_groups.all.groups
	}
	`, name, name, name, name)
}
