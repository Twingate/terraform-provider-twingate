package datasource

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	groupsLen         = attr.Len(attr.Groups)
	groupNamePath     = attr.Path(attr.Groups, attr.Name)
	groupPolicyIDPath = attr.Path(attr.Groups, attr.SecurityPolicyID)
)

func TestAccDatasourceTwingateGroups_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups Basic", func(t *testing.T) {
		groupName := test.RandomName()

		const theDatasource = "data.twingate_groups.out_dgs1"

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
					Config: testDatasourceTwingateGroups(groupName, testPolicy.ID),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(theDatasource, groupsLen, "2"),
						resource.TestCheckResourceAttr(theDatasource, groupNamePath, groupName),
						resource.TestCheckResourceAttr(theDatasource, groupPolicyIDPath, testPolicy.ID),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroups(name, securityPolicyID string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test_dgs1_1" {
	  name = "%s"
	  security_policy_id = "%s"
	}

	resource "twingate_group" "test_dgs1_2" {
	  name = "%s"
	  security_policy_id = "%s"
	}

	data "twingate_groups" "out_dgs1" {
	  name = "%s"

	  depends_on = [twingate_group.test_dgs1_1, twingate_group.test_dgs1_2]
	}
	`, name, securityPolicyID, name, securityPolicyID, name)
}

func TestAccDatasourceTwingateGroups_emptyResult(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups - empty result", func(t *testing.T) {
		groupName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testTwingateGroupsDoesNotExists(groupName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_groups.out_dgs2", groupsLen, "0"),
					),
				},
			},
		})
	})
}

func testTwingateGroupsDoesNotExists(name string) string {
	return fmt.Sprintf(`
	data "twingate_groups" "out_dgs2" {
	  name = "%s"
	}
	`, name)
}

func TestAccDatasourceTwingateGroupsWithFilters_basic(t *testing.T) {
	acctests.SetPageLimit(1)
	groupName := test.RandomName()

	const theDatasource = "data.twingate_groups.out_dgs2"

	t.Run("Test Twingate Datasource : Acc Groups with filters - basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroupsWithFilters(groupName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(theDatasource, groupsLen, "2"),
						resource.TestCheckResourceAttr(theDatasource, groupNamePath, groupName),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroupsWithFilters(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test_dgs2_1" {
	  name = "%s"
	}

	resource "twingate_group" "test_dgs2_2" {
	  name = "%s"
	}

	data "twingate_groups" "out_dgs2" {
	  name = "%s"
	  type = "MANUAL"
	  is_active = true

	  depends_on = [twingate_group.test_dgs2_1, twingate_group.test_dgs2_2]
	}
	`, name, name, name)
}

func TestAccDatasourceTwingateGroupsWithFilters_ErrorNotSupportedTypes(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Groups with filters - error not supported types", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateGroupsWithFilterNotSupportedType(),
					ExpectError: regexp.MustCompile("Attribute type value must be one of"),
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
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
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

		groupName := test.RandomName()

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateGroupsWithDatasource(groupName),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.twingate_groups.two_dgs3", groupNamePath, groupName),
						resource.TestCheckResourceAttr("data.twingate_groups.one_dgs3", groupsLen, "1"),
						resource.TestCheckResourceAttr("data.twingate_groups.two_dgs3", groupsLen, "2"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateGroupsWithDatasource(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test_dgs3_1" {
	  name = "%s"
	}

	resource "twingate_group" "test_dgs3_2" {
	  name = "%s"
	}

	resource "twingate_group" "test_dgs3_3" {
	  name = "%s-1"
	}

	data "twingate_groups" "two_dgs3" {
	  name = "%s"

	  depends_on = [twingate_group.test_dgs3_1, twingate_group.test_dgs3_2, twingate_group.test_dgs3_3]
	}

	data "twingate_groups" "one_dgs3" {
	  name = "%s-1"

	  depends_on = [twingate_group.test_dgs3_1, twingate_group.test_dgs3_2, twingate_group.test_dgs3_3]
	}
	`, name, name, name, name, name)
}
