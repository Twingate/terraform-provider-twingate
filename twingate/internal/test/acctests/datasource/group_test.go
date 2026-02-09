package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateGroup_basic(t *testing.T) {
	t.Parallel()

	groupName := test.RandomName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
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
	t.Parallel()

	groupID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Group:%d", acctest.RandInt())))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testTwingateGroupDoesNotExists(groupID),
				ExpectError: regexp.MustCompile("failed to read group with id"),
			},
		},
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
	t.Parallel()

	groupID := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testTwingateGroupDoesNotExists(groupID),
				ExpectError: regexp.MustCompile("failed to read group with id"),
			},
		},
	})
}
