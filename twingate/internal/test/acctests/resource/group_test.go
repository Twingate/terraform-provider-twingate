package resource

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

var (
	userIdsLen = attr.Len(attr.UserIDs)
)

func TestAccTwingateGroupCreateUpdate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test001"
	theResource := acctests.TerraformGroup(terraformResourceName)
	nameBefore := test.RandomName()
	nameAfter := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateGroup(terraformResourceName, nameBefore),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, nameBefore),
				),
			},
			{
				Config: terraformResourceTwingateGroup(terraformResourceName, nameAfter),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, nameAfter),
				),
			},
		},
	})
}

func terraformResourceTwingateGroup(terraformResourceName, name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "%s" {
	  name = "%s"
	}
	`, terraformResourceName, name)
}

func TestAccTwingateGroupDeleteNonExisting(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test002"
	theResource := acctests.TerraformGroup(terraformResourceName)
	groupName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  terraformResourceTwingateGroup(terraformResourceName, groupName),
				Destroy: true,
			},
			{
				Config: terraformResourceTwingateGroup(terraformResourceName, groupName),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateGroupReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test003"
	theResource := acctests.TerraformGroup(terraformResourceName)
	groupName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateGroup(terraformResourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateGroup),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceTwingateGroup(terraformResourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func TestAccTwingateGroupWithSecurityPolicy(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test004"
	theResource := acctests.TerraformGroup(terraformResourceName)
	name := test.RandomName()

	securityPolicies, err := acctests.ListSecurityPolicies()
	if err != nil {
		t.Skip("can't run test:", err)
	}

	testPolicy := securityPolicies[0]

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateGroup(terraformResourceName, name),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
				),
			},
			{
				Config: terraformResourceTwingateGroupWithSecurityPolicy(terraformResourceName, name, testPolicy.ID),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
					sdk.TestCheckResourceAttr(theResource, attr.SecurityPolicyID, testPolicy.ID),
				),
			},
			{
				// expecting no changes
				PlanOnly: true,
				Config:   terraformResourceTwingateGroup(terraformResourceName, name),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name),
				),
			},
		},
	})
}

func terraformResourceTwingateGroupWithSecurityPolicy(terraformResourceName, name, securityPolicyID string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "%s" {
	  name = "%s"
	  security_policy_id = "%s"
	}
	`, terraformResourceName, name, securityPolicyID)
}

func TestAccTwingateGroupUsersAuthoritativeByDefault(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test005"
	theResource := acctests.TerraformGroup(terraformResourceName)
	groupName := test.RandomName()

	users, userIDs := genNewUsers("u005", 3)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateGroupWithUsers(terraformResourceName, groupName, users, userIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "1"),
					acctests.CheckGroupUsersLen(theResource, 1),
				),
			},
			{
				Config: terraformResourceTwingateGroupWithUsers(terraformResourceName, groupName, users, userIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					// added new user to the group though API
					acctests.AddGroupUser(theResource, groupName, userIDs[1]),
					acctests.WaitTestFunc(),
					acctests.CheckGroupUsersLen(theResource, 2),
				),
				// expecting drift - terraform going to remove unknown user
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceTwingateGroupWithUsers(terraformResourceName, groupName, users, userIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "1"),
					acctests.CheckGroupUsersLen(theResource, 1),
				),
			},
			{
				// added 2 new users to the group though terraform
				Config: terraformResourceTwingateGroupWithUsers(terraformResourceName, groupName, users, userIDs[:3]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "3"),
					acctests.CheckGroupUsersLen(theResource, 3),
				),
			},
			{
				Config: terraformResourceTwingateGroupWithUsers(terraformResourceName, groupName, users, userIDs[:3]),
				Check: acctests.ComposeTestCheckFunc(
					// delete one user from the group though API
					acctests.DeleteGroupUser(theResource, userIDs[2]),
					acctests.WaitTestFunc(),
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "3"),
					acctests.CheckGroupUsersLen(theResource, 2),
				),
				// expecting drift - terraform going to restore deleted user
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceTwingateGroupWithUsers(terraformResourceName, groupName, users, userIDs[:3]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "3"),
					acctests.CheckGroupUsersLen(theResource, 3),
				),
			},
			{
				// remove 2 users from the group though terraform
				Config: terraformResourceTwingateGroupWithUsers(terraformResourceName, groupName, users, userIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "1"),
					acctests.CheckGroupUsersLen(theResource, 1),
				),
			},
			{
				// expecting no drift
				Config:   terraformResourceTwingateGroupWithUsersAuthoritative(terraformResourceName, groupName, users, userIDs[:1], true),
				PlanOnly: true,
			},
			{
				Config: terraformResourceTwingateGroupWithUsersAuthoritative(terraformResourceName, groupName, users, userIDs[:2], true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "2"),
					acctests.CheckGroupUsersLen(theResource, 2),
				),
			},
		},
	})
}

func terraformResourceTwingateGroupWithUsers(terraformResourceName, name string, users, usersID []string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_group" "%s" {
	  name = "%s"
	  user_ids = [%s]
	}
	`, strings.Join(users, "\n"), terraformResourceName, name, strings.Join(usersID, ", "))
}

func terraformResourceTwingateGroupWithUsersAuthoritative(terraformResourceName, name string, users, usersID []string, authoritative bool) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_group" "%s" {
	  name = "%s"
	  user_ids = [%s]
	  is_authoritative = %v
	}
	`, strings.Join(users, "\n"), terraformResourceName, name, strings.Join(usersID, ", "), authoritative)
}

func TestAccTwingateGroupUsersNotAuthoritative(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test006"
	theResource := acctests.TerraformGroup(terraformResourceName)
	groupName := test.RandomName()

	users, userIDs := genNewUsers("u006", 3)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateGroupWithUsersAuthoritative(terraformResourceName, groupName, users, userIDs[:1], false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "1"),
					acctests.CheckGroupUsersLen(theResource, 1),
				),
			},
			{
				Config: terraformResourceTwingateGroupWithUsersAuthoritative(terraformResourceName, groupName, users, userIDs[:1], false),
				Check: acctests.ComposeTestCheckFunc(
					// added new user to the group though API
					acctests.AddGroupUser(theResource, groupName, userIDs[2]),
					acctests.WaitTestFunc(),
					acctests.CheckGroupUsersLen(theResource, 2),
				),
			},
			{
				// added new user to the group though terraform
				Config: terraformResourceTwingateGroupWithUsersAuthoritative(terraformResourceName, groupName, users, userIDs[:2], false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "2"),
					acctests.CheckGroupUsersLen(theResource, 3),
				),
			},
			{
				// remove one user from the group though terraform
				Config: terraformResourceTwingateGroupWithUsersAuthoritative(terraformResourceName, groupName, users, userIDs[:1], false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, userIdsLen, "1"),
					acctests.CheckGroupUsersLen(theResource, 2),
					// remove one user from the group though API
					acctests.DeleteGroupUser(theResource, userIDs[2]),
					acctests.WaitTestFunc(),
					acctests.CheckGroupUsersLen(theResource, 1),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   terraformResourceTwingateGroupWithUsersAuthoritative(terraformResourceName, groupName, users, userIDs[:1], false),
				PlanOnly: true,
			},
		},
	})
}

func TestAccTwingateGroupUsersCursor(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test007"
	theResource := acctests.TerraformGroup(terraformResourceName)
	groupName := test.RandomName()

	users, userIDs := genNewUsers("u007", 3)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateGroupAndUsers(terraformResourceName, groupName, users, userIDs),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckGroupUsersLen(theResource, len(users)),
				),
			},
			{
				Config: terraformResourceTwingateGroupAndUsers(terraformResourceName, groupName, users, userIDs[:2]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckGroupUsersLen(theResource, 2),
				),
			},
		},
	})
}

func terraformResourceTwingateGroupAndUsers(terraformResourceName, name string, users, userIDs []string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_group" "%s" {
	  name = "%s"
	  user_ids = [%s]
	}
	`, strings.Join(users, "\n"), terraformResourceName, name, strings.Join(userIDs, ", "))
}

func TestAccTwingateGroupImport(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test008"
	theResource := acctests.TerraformGroup(terraformResourceName)
	groupName := test.RandomName()

	users, userIDs := genNewUsers("u008", 3)

	securityPolicies, err := acctests.ListSecurityPolicies()
	if err != nil {
		t.Skip("can't run test:", err)
	}

	testPolicy := securityPolicies[0]

	authoritative := true

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateGroupFull(terraformResourceName, groupName, users, userIDs, testPolicy.ID, authoritative),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				ImportState:  true,
				ResourceName: theResource,
				ImportStateCheck: acctests.CheckImportState(map[string]string{
					attr.Name:                  groupName,
					attr.SecurityPolicyID:      testPolicy.ID,
					attr.IsAuthoritative:       fmt.Sprintf("%v", authoritative),
					attr.LenAttr(attr.UserIDs): fmt.Sprintf("%d", len(users)),
				}),
			},
		},
	})
}

func terraformResourceTwingateGroupFull(terraformResourceName, name string, users, usersID []string, securityPolicy string, authoritative bool) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_group" "%s" {
	  name = "%s"
	  user_ids = [%s]
	  security_policy_id = "%s"
	  is_authoritative = %v
	}
	`, strings.Join(users, "\n"), terraformResourceName, name, strings.Join(usersID, ", "), securityPolicy, authoritative)
}
