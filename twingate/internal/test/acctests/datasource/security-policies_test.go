package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateSecurityPoliciesBasic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Security Policies - basic", func(t *testing.T) {
		securityPolicies, err := acctests.ListSecurityPolicies()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		sdk.Test(t, sdk.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			Steps: []sdk.TestStep{
				{
					Config: testDatasourceTwingateSecurityPolicies(),
					Check: acctests.ComposeTestCheckFunc(
						sdk.TestCheckResourceAttr("data.twingate_security_policies.all", "security_policies.#", fmt.Sprintf("%d", len(securityPolicies))),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateSecurityPolicies() string {
	return `
	data "twingate_security_policies" "all" {}
	`
}
