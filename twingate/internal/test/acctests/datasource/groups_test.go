package datasource

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	groupsLen         = attr.Len(attr.Groups)
	groupNamePath     = attr.Path(attr.Groups, attr.Name)
	groupPolicyIDPath = attr.Path(attr.Groups, attr.SecurityPolicyID)
)

func TestAccDatasourceTwingateGroups_basic(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	  types = ["MANUAL"]
	  is_active = true

	  depends_on = [twingate_group.test_dgs2_1, twingate_group.test_dgs2_2]
	}
	`, name, name, name)
}

func TestAccDatasourceTwingateGroupsWithFilters_ErrorNotSupportedTypes(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testTwingateGroupsWithFilterNotSupportedType(),
				ExpectError: regexp.MustCompile("Attribute types.* value must be one of"),
			},
		},
	})
}

func testTwingateGroupsWithFilterNotSupportedType() string {
	return `
	data "twingate_groups" "test" {
	  types = ["OTHER"]
	}

	output "my_groups" {
	  value = data.twingate_groups.test.groups
	}
	`
}

func TestAccDatasourceTwingateGroups_WithEmptyFilters(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:             testTwingateGroupsWithEmptyFilter(test.RandomGroupName()),
				ExpectNonEmptyPlan: true,
				Check: acctests.ComposeTestCheckFunc(
					testCheckResourceAttrNotEqual("data.twingate_groups.all", groupsLen, "0"),
				),
			},
		},
	})
}

func testTwingateGroupsWithEmptyFilter(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test_group" {
	  name = "%s"
	}

	data "twingate_groups" "all" {}

	output "my_groups" {
	  value = data.twingate_groups.all.groups

	  depends_on = [twingate_group.test_group]
	}
	`, name)
}

func TestAccDatasourceTwingateGroups_withTwoDatasource(t *testing.T) {
	t.Parallel()

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

func TestAccDatasourceTwingateGroupsWithFilterByPrefix(t *testing.T) {
	t.Parallel()

	prefix := test.Prefix() + "-" + acctest.RandString(5)
	resourceName := test.RandomResourceName()

	theDatasource := "data.twingate_groups." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateGroupsWithFilter(
					resourceName,
					prefix+"_g1",
					prefix+"_g2",
					prefix,
					attr.FilterByPrefix,
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, groupsLen, "2"),
				),
			},
		},
	})
}

func testDatasourceTwingateGroupsWithFilter(resourceName, name1, name2, name, filter string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "%[1]s_1" {
	  name = "%[2]s"
	}

	resource "twingate_group" "%[1]s_2" {
	  name = "%[3]s"
	}

	data "twingate_groups" "%[1]s" {
	  name%[4]s = "%[5]s"
	  types = ["MANUAL"]
	  is_active = true

	  depends_on = [twingate_group.%[1]s_1, twingate_group.%[1]s_2]
	}
	`, resourceName, name1, name2, filter, name)
}

func TestAccDatasourceTwingateGroupsWithFilterBySuffix(t *testing.T) {
	t.Parallel()

	prefix := test.Prefix()
	resourceName := test.RandomResourceName()
	suffix := acctest.RandString(5)

	theDatasource := "data.twingate_groups." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateGroupsWithFilter(
					resourceName,
					fmt.Sprintf("%s_g1_%s", prefix, suffix),
					fmt.Sprintf("%s_g2_%s", prefix, suffix),
					suffix,
					attr.FilterBySuffix,
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, groupsLen, "2"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateGroupsWithFilterByContains(t *testing.T) {
	t.Parallel()

	prefix := test.Prefix()
	resourceName := test.RandomResourceName()
	contains := acctest.RandString(5)

	theDatasource := "data.twingate_groups." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateGroupsWithFilter(
					resourceName,
					fmt.Sprintf("%s_%s_g1", prefix, contains),
					fmt.Sprintf("%s_%s_g2", prefix, contains),
					contains,
					attr.FilterByContains,
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, groupsLen, "2"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateGroupsWithFilterByRegexp(t *testing.T) {
	t.Parallel()

	prefix := test.Prefix()
	resourceName := test.RandomResourceName()
	contains := acctest.RandString(5)

	theDatasource := "data.twingate_groups." + resourceName

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateGroupsWithFilter(
					resourceName,
					fmt.Sprintf("%s_%s_g1", prefix, contains),
					fmt.Sprintf("%s_%s_g2", prefix, contains),
					fmt.Sprintf(".*_%s_.*", contains),
					attr.FilterByRegexp,
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, groupsLen, "2"),
				),
			},
		},
	})
}
