package resource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	statusAttr = "status"
)

func createServiceAccountKey(terraformResourceName, serviceAccountName string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_service_account_key" "%s" {
	  service_account_id = twingate_service_account.%s.id
	}
	`, createServiceAccount(terraformResourceName, serviceAccountName), terraformResourceName, terraformResourceName)
}

func createServiceAccountKeyWithName(terraformResourceName, serviceAccountName, serviceAccountKeyName string) string {
	return fmt.Sprintf(`
	%s

	resource "twingate_service_account_key" "%s" {
	  service_account_id = twingate_service_account.%s.id
	  name = "%s"
	}
	`, createServiceAccount(terraformResourceName, serviceAccountName), terraformResourceName, terraformResourceName, serviceAccountKeyName)
}

func TestAccTwingateServiceAccountKeyCreateUpdate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Key Create/Update", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceAccount := acctests.TerraformServiceAccount(terraformResourceName)
		serviceAccountKey := acctests.TerraformServiceAccountKey(terraformResourceName)

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceAccountKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccount),
						sdk.TestCheckResourceAttr(serviceAccount, nameAttr, serviceAccountName),
						acctests.CheckTwingateResourceExists(serviceAccountKey),
						sdk.TestCheckResourceAttr(serviceAccountKey, statusAttr, model.StatusActive),
					),
				},
				{
					Config: createServiceAccountKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccount),
						sdk.TestCheckResourceAttr(serviceAccount, nameAttr, serviceAccountName),
						acctests.CheckTwingateResourceExists(serviceAccountKey),
						sdk.TestCheckResourceAttr(serviceAccountKey, statusAttr, model.StatusActive),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceAccountKeyCreateUpdateWithName(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Key Create/Update With Name", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceAccount := acctests.TerraformServiceAccount(terraformResourceName)
		serviceAccountKey := acctests.TerraformServiceAccountKey(terraformResourceName)
		beforeName := test.RandomName()
		afterName := test.RandomName()

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceAccountKeyWithName(terraformResourceName, serviceAccountName, beforeName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccount),
						sdk.TestCheckResourceAttr(serviceAccount, nameAttr, serviceAccountName),
						acctests.CheckTwingateResourceExists(serviceAccountKey),
						sdk.TestCheckResourceAttr(serviceAccountKey, nameAttr, beforeName),
						sdk.TestCheckResourceAttr(serviceAccountKey, statusAttr, model.StatusActive),
					),
				},
				{
					Config: createServiceAccountKeyWithName(terraformResourceName, serviceAccountName, afterName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccount),
						sdk.TestCheckResourceAttr(serviceAccount, nameAttr, serviceAccountName),
						acctests.CheckTwingateResourceExists(serviceAccountKey),
						sdk.TestCheckResourceAttr(serviceAccountKey, nameAttr, afterName),
						sdk.TestCheckResourceAttr(serviceAccountKey, statusAttr, model.StatusActive),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceAccountKeyReCreateAfterInactive(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Key ReCreate After Inactive", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceAccountKey := acctests.TerraformServiceAccountKey(terraformResourceName)

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceAccountKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccountKey),
						acctests.RevokeTwingateServiceAccountKey(serviceAccountKey),
						acctests.WaitTestFunc(),
						acctests.CheckTwingateServiceAccountKeyStatus(serviceAccountKey, model.StatusRevoked),
					),
				},
				{
					Config: createServiceAccountKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccountKey),
						acctests.CheckTwingateServiceAccountKeyStatus(serviceAccountKey, model.StatusActive),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceAccountKeyDelete(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Key Delete", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceAccountKey := acctests.TerraformServiceAccountKey(terraformResourceName)

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config:  createServiceAccountKey(terraformResourceName, serviceAccountName),
					Destroy: true,
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceDoesNotExists(serviceAccountKey),
					),
				},
			},
		})
	})
}

func TestAccTwingateServiceAccountKeyReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Service Account Key ReCreate After Delete", func(t *testing.T) {
		serviceAccountName := test.RandomName()
		terraformResourceName := test.TerraformRandName("test_key")
		serviceAccountKey := acctests.TerraformServiceAccountKey(terraformResourceName)

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			CheckDestroy:      acctests.CheckTwingateServiceAccountDestroy,
			Steps: []sdk.TestStep{
				{
					Config: createServiceAccountKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccountKey),
						acctests.RevokeTwingateServiceAccountKey(serviceAccountKey),
						acctests.DeleteTwingateResource(serviceAccountKey, resource.TwingateServiceAccountKey),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: createServiceAccountKey(terraformResourceName, serviceAccountName),
					Check: acctests.ComposeTestCheckFunc(
						acctests.CheckTwingateResourceExists(serviceAccountKey),
					),
				},
			},
		})
	})
}
