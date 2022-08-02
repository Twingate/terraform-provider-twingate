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
		groupName := acctest.RandomWithPrefix(testPrefixName)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroup(groupName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_group", groupName),
						resource.TestCheckOutput("my_group_is_active", "true"),
						resource.TestCheckOutput("my_group_type", "MANUAL"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroup(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "foo" {
	  name = "%s"
	}

	data "twingate_group" "bar" {
	  id = twingate_group.foo.id
	}

	output "my_group" {
	  value = data.twingate_group.bar.name
	}

	output "my_group_is_active" {
	  value = data.twingate_group.bar.is_active
	}

	output "my_group_type" {
	  value = data.twingate_group.bar.type
	}
	`, name)
}

func TestAccDatasourceTwingateGroup_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group - does not exists", func(t *testing.T) {
		groupID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Group:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
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
	data "twingate_group" "foo" {
	  id = "%s"
	}
	`, id)
}

func TestAccDatasourceTwingateGroup_invalidGroupID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group - failed parse group ID", func(t *testing.T) {
		groupID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
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
