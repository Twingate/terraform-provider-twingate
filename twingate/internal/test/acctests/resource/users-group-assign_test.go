package resource

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const userIDsAttr = "user_ids"

func TestAccTwingateGroupUsersAssign(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Users Assign Create/Update", func(t *testing.T) {
		const terraformResourceName = "admins"
		theResource := acctests.TerraformUsersGroupAssign(terraformResourceName)
		groupName := test.RandomName()

		users, err := acctests.GetTestUsers()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		admins := utils.Filter[*model.User](users, func(user *model.User) bool {
			return user.IsAdmin()
		})

		if len(admins) == 0 {
			t.Skip("test requires at least one admin user")
		}

		adminIDs := utils.Map[*model.User, string](admins, func(user *model.User) string {
			return user.ID
		})

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateGroupDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceGroupUsersAssign(groupName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, userIDsAttr+".#", fmt.Sprintf("%d", len(adminIDs))),
						sdk.TestCheckResourceAttr(theResource, userIDsAttr+".0", adminIDs[0]),
					),
				},
				{
					Config: terraformResourceGroupUsersAssignWithUserIDs(terraformResourceName, groupName, adminIDs[:1]),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, userIDsAttr+".#", "1"),
					),
				},
			},
		})
	})
}

func terraformResourceGroupUsersAssign(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "admins" {
	  name = "%s"
	}

	data "twingate_users" "all" {}

	locals {
	  admin_users = [for user in data.twingate_users.all.users : user.id if user.is_admin == true]
	}

	resource "twingate_users_group_assign" "admins" {
	  user_ids = local.admin_users
	  group_id = twingate_group.admins.id
	}
	`, name)
}

func terraformResourceGroupUsersAssignWithUserIDs(terraformResourceName, name string, userIDs []string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "%s" {
	  name = "%s"
	}

	resource "twingate_users_group_assign" "%s" {
	  user_ids = ["%s"]
	  group_id = twingate_group.%s.id
	}
	`, terraformResourceName, name, terraformResourceName, strings.Join(userIDs, `", "`), terraformResourceName)
}

func TestAccTwingateGroupUsersAssignWithInvalidGroupId(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Users Assign With Invalid Group ID", func(t *testing.T) {
		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			Steps: []sdk.TestStep{
				{
					Config:      terraformResourceGroupUsersAssignWithInvalidGroupID(),
					ExpectError: regexp.MustCompile("Error: failed to update group with id foo"),
				},
			},
		})
	})
}

func terraformResourceGroupUsersAssignWithInvalidGroupID() string {
	return `
	resource "twingate_users_group_assign" "admins" {
	  user_ids = []
	  group_id = "foo"
	}
	`
}

func TestAccTwingateGroupUsersAssignWithDifferentIDsOrder(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group User Assign With Different IDs Order", func(t *testing.T) {
		const terraformResourceName = "group_assign_1"
		theResource := acctests.TerraformUsersGroupAssign(terraformResourceName)
		groupName := test.RandomName()

		users, err := acctests.GetTestUsers()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		if len(users) < 2 {
			t.Skip("test requires at least two user")
		}

		userIDs := utils.Map[*model.User, string](users[:2], func(user *model.User) string {
			return user.ID
		})

		otherOrder := []string{userIDs[1], userIDs[0]}

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateGroupDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceGroupUsersAssignWithUserIDs(terraformResourceName, groupName, userIDs),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, userIDsAttr+".#", "2"),
						sdk.TestCheckResourceAttr(theResource, userIDsAttr+".0", userIDs[0]),
					),
				},
				// expecting no changes - empty plan
				{
					Config: terraformResourceGroupUsersAssignWithUserIDs(terraformResourceName, groupName, otherOrder),
					//Check: acctests.ComposeTestCheckFunc(
					//	acctests.CheckTwingateResourceExists(theResource),
					//	sdk.TestCheckResourceAttr(theResource, userIDsAttr+".#", "2"),
					//	sdk.TestCheckResourceAttr(theResource, userIDsAttr+".0", userIDs[0]),
					//),
					PlanOnly: true,
				},
			},
		})
	})
}
