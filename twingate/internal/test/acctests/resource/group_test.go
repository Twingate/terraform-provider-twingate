package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const securityPolicyAttr = "security_policy_id"

func TestAccTwingateGroupCreateUpdate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Create/Update", func(t *testing.T) {
		const terraformResourceName = "test001"
		theResource := acctests.TerraformGroup(terraformResourceName)
		nameBefore := test.RandomName()
		nameAfter := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateGroupDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateGroup(terraformResourceName, nameBefore),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, nameBefore),
					),
				},
				{
					Config: terraformResourceTwingateGroup(terraformResourceName, nameAfter),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, nameAfter),
					),
				},
			},
		})
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
	t.Run("Test Twingate Resource : Acc Group Delete NonExisting", func(t *testing.T) {
		const terraformResourceName = "test002"
		theResource := acctests.TerraformGroup(terraformResourceName)
		groupName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateGroupDestroy,
			Steps: []sdk.TestStep{
				{
					Config:  terraformResourceTwingateGroup(terraformResourceName, groupName),
					Destroy: true,
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceDoesNotExists(theResource),
					),
				},
			},
		})
	})
}

func TestAccTwingateGroupReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Create After Deletion", func(t *testing.T) {
		const terraformResourceName = "test003"
		theResource := acctests.TerraformGroup(terraformResourceName)
		groupName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateGroupDestroy,
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
	})
}

func TestAccTwingateGroupWithSecurityPolicy(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Create/Update - With Security Policy", func(t *testing.T) {
		const terraformResourceName = "test004"
		theResource := acctests.TerraformGroup(terraformResourceName)
		name := test.RandomName()

		securityPolicies, err := acctests.ListSecurityPolicies()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		testPolicy := securityPolicies[0]

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateGroupDestroy,
			Steps: []sdk.TestStep{
				{
					Config: terraformResourceTwingateGroup(terraformResourceName, name),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, name),
					),
				},
				{
					Config: terraformResourceTwingateGroupWithSecurityPolicy(terraformResourceName, name, testPolicy.ID),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, name),
						sdk.TestCheckResourceAttr(theResource, securityPolicyAttr, testPolicy.ID),
					),
				},
				{
					// expecting no changes
					PlanOnly: true,
					Config:   terraformResourceTwingateGroup(terraformResourceName, name),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, nameAttr, name),
					),
				},
			},
		})
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
