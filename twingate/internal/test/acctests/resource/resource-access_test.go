package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTwingateResourceAccessGroup(t *testing.T) {
	t.Parallel()

	const terraformName = "ra1"
	theAccess := acctests.TerraformResourceAccess(terraformName)
	theResource := acctests.TerraformResource(terraformName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groupName := test.RandomGroupName()

	_, testPolicy := preparePolicies(t)

	withPolicy := fmt.Sprintf(`
	  security_policy_id = "%s"
	  access_policy {
	    mode          = "AUTO_LOCK"
	    approval_mode = "AUTOMATIC"
	    duration      = "30d"
	  }`, testPolicy)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceAccessDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceAccessGroup(terraformName, remoteNetworkName, resourceName, groupName, ""),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckNoResourceAttr(theAccess, attr.SecurityPolicyID),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
			{
				Config: terraformResourceAccessGroup(terraformName, remoteNetworkName, resourceName, groupName, withPolicy),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theAccess, plancheck.ResourceActionUpdate),
					},
				},
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theAccess, attr.SecurityPolicyID, testPolicy),
					sdk.TestCheckResourceAttr(theAccess, attr.Path(attr.AccessPolicy, attr.Mode), "AUTO_LOCK"),
					sdk.TestCheckResourceAttr(theAccess, attr.Path(attr.AccessPolicy, attr.Duration), "30d"),
					acctests.CheckTwingateResourceSecurityPolicyOnGroupAccess(theResource, testPolicy),
				),
			},
			{
				ResourceName:      theAccess,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: terraformResourceAccessGroup(terraformName, remoteNetworkName, resourceName, groupName, ""),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckNoResourceAttr(theAccess, attr.SecurityPolicyID),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
			{
				Config: terraformResourceAccessGroup(terraformName, remoteNetworkName, resourceName, groupName, ""),
				Check: acctests.ComposeTestCheckFunc(
					acctests.DeleteResourceGroup(theResource, acctests.TerraformGroup(terraformName)),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceAccessGroup(terraformName, remoteNetworkName, resourceName, groupName, ""),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theAccess, plancheck.ResourceActionCreate),
					},
				},
				Check: acctests.CheckResourceGroupsLen(theResource, 1),
			},
			{
				Config: resourceAccessPrerequisites(terraformName, remoteNetworkName, resourceName, groupName),
				Check:  acctests.CheckResourceGroupsLen(theResource, 0),
			},
		},
	})
}

func terraformResourceAccessGroup(terraformName, remoteNetworkName, resourceName, groupName, extra string) string {
	return resourceAccessPrerequisites(terraformName, remoteNetworkName, resourceName, groupName) + fmt.Sprintf(`
	resource "twingate_resource_access" "%[1]s" {
	  resource_id = twingate_resource.%[1]s.id
	  group_id    = twingate_group.%[1]s.id
	  %[2]s
	}
	`, terraformName, extra)
}

func resourceAccessPrerequisites(terraformName, remoteNetworkName, resourceName, groupName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[2]s"
	}

	resource "twingate_group" "%[1]s" {
	  name = "%[4]s"
	}

	resource "twingate_resource" "%[1]s" {
	  name              = "%[3]s"
	  address           = "acc-test-resource-access.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  is_authoritative  = false
	}
	`, terraformName, remoteNetworkName, resourceName, groupName)
}
