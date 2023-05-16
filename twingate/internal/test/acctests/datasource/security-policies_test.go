package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateSecurityPoliciesBasic(t *testing.T) {
	t.Skip("test with cursor")
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc Security Policies - basic", func(t *testing.T) {
		acctests.SetPageLimit(1)

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
						sdk.TestCheckResourceAttr("data.twingate_security_policies.all", attr.Len(attr.SecurityPolicies), fmt.Sprintf("%d", len(securityPolicies))),
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
