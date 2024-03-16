package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func createServiceAccount(resourceName, serviceAccountName string) string {
	return fmt.Sprintf(`
	resource "twingate_service_account" "%s" {
	  name = "%s"
	}
	`, resourceName, serviceAccountName)
}

func TestAccTwingateServiceAccountCreateUpdate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Create/Update", func(t *testing.T) {
		const terraformResourceName = "test01"
		theResource := acctests.TerraformServiceAccount(terraformResourceName)
		nameBefore := test.RandomName()
		nameAfter := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceAccount(terraformResourceName, nameBefore),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, attr.Name, nameBefore),
					),
				},
				{
					Config: createServiceAccount(terraformResourceName, nameAfter),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						sdk.TestCheckResourceAttr(theResource, attr.Name, nameAfter),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceAccountDeleteNonExisting(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Delete NonExisting", func(t *testing.T) {
		const terraformResourceName = "test02"
		theResource := acctests.TerraformServiceAccount(terraformResourceName)
		name := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config:  createServiceAccount(terraformResourceName, name),
					Destroy: true,
				},
				{
					Config: createServiceAccount(terraformResourceName, name),
					ConfigPlanChecks: sdk.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
						},
					},
				},
			},
		})
	})
}

func TestAccTwingateServiceAccountReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Create After Deletion", func(t *testing.T) {
		const terraformResourceName = "test03"
		theResource := acctests.TerraformServiceAccount(terraformResourceName)
		name := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			CheckDestroy:             acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceAccount(terraformResourceName, name),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
						acctests.DeleteTwingateResource(theResource, resource.TwingateServiceAccount),
						acctests.WaitTestFunc(),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: createServiceAccount(terraformResourceName, name),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(theResource),
					),
				},
			},
		})
	})
}
