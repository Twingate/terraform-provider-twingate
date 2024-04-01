package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var securityPolicyNamePath = attr.Path(attr.SecurityPolicies, attr.Name)

func TestAccDatasourceTwingateSecurityPoliciesBasic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Security Policies - basic", func(t *testing.T) {
		acctests.SetPageLimit(1)

		securityPolicies, err := acctests.ListSecurityPolicies()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		sdk.Test(t, sdk.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
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

func testDatasourceTwingateSecurityPoliciesFilter(filter, name string) string {
	return fmt.Sprintf(`
	data "twingate_security_policies" "filtered" {
	  name%[1]s = "%[2]s"
	}
	`, filter, name)
}

func TestAccDatasourceTwingateSecurityPoliciesFilterByPrefix(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_security_policies.filtered"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateSecurityPoliciesFilter(attr.FilterByPrefix, "Def"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theDatasource, attr.Len(attr.SecurityPolicies), "1"),
					sdk.TestCheckResourceAttr(theDatasource, securityPolicyNamePath, "Default Policy"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateSecurityPoliciesFilterBySuffix(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_security_policies.filtered"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateSecurityPoliciesFilter(attr.FilterBySuffix, "ault Policy"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theDatasource, attr.Len(attr.SecurityPolicies), "1"),
					sdk.TestCheckResourceAttr(theDatasource, securityPolicyNamePath, "Default Policy"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateSecurityPoliciesFilterByContains(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_security_policies.filtered"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateSecurityPoliciesFilter(attr.FilterByContains, "ault"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theDatasource, attr.Len(attr.SecurityPolicies), "1"),
					sdk.TestCheckResourceAttr(theDatasource, securityPolicyNamePath, "Default Policy"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateSecurityPoliciesFilterByRegexp(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_security_policies.filtered"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateSecurityPoliciesFilter(attr.FilterByRegexp, ".*ault .*"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theDatasource, attr.Len(attr.SecurityPolicies), "1"),
					sdk.TestCheckResourceAttr(theDatasource, securityPolicyNamePath, "Default Policy"),
				),
			},
		},
	})
}
