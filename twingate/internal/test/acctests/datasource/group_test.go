package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateGroup_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group Basic", func(t *testing.T) {
		groupName := test.RandomName()

		securityPolicies, err := acctests.ListSecurityPolicies()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		testPolicy := securityPolicies[0]

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroup(groupName, testPolicy.ID),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_group_dg1", groupName),
						resource.TestCheckOutput("my_group_is_active_dg1", "true"),
						resource.TestCheckOutput("my_group_type_dg1", "MANUAL"),
						resource.TestCheckOutput("my_group_policy_dg1", testPolicy.ID),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroup(name, securityPolicyID string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "foo_dg1" {
	  name = "%s"
	  security_policy_id = "%s"
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

	output "my_group_policy_dg1" {
	  value = data.twingate_group.bar_dg1.security_policy_id
	}
	`, name, securityPolicyID)
}

func TestAccDatasourceTwingateGroup_negative(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Group - does not exists", func(t *testing.T) {
		groupID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Group:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
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
			ProtoV6ProviderFactories: acctests.ProviderFactories,
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
