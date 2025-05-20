package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var dlpPolicyNamePath = attr.Path(attr.DLPPolicies, attr.Name)

// Note: All tests are commented out since we don't have access to real DLP policies in the test environment.
// They would follow the same pattern as the security policies tests.

/*
func TestAccDatasourceTwingateDLPPoliciesBasic(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_dlp_policies.all"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateDLPPolicies(),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theDatasource, attr.ID),
				),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicies() string {
	return fmt.Sprintf(`
data "twingate_dlp_policies" "all" {
}
`)
}

func testDatasourceTwingateDLPPoliciesFilter(filter, name string) string {
	return fmt.Sprintf(`
data "twingate_dlp_policies" "filtered" {
  name%s = "%s"
}
`, filter, name)
}

func TestAccDatasourceTwingateDLPPoliciesFilterByPrefix(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_dlp_policies.filtered"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateDLPPoliciesFilter(attr.FilterByPrefix, "De"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theDatasource, attr.Len(attr.DLPPolicies), "1"),
					sdk.TestCheckResourceAttr(theDatasource, dlpPolicyNamePath, "Default DLP Policy"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateDLPPoliciesFilterBySuffix(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_dlp_policies.filtered"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateDLPPoliciesFilter(attr.FilterBySuffix, "DLP Policy"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theDatasource, attr.Len(attr.DLPPolicies), "1"),
					sdk.TestCheckResourceAttr(theDatasource, dlpPolicyNamePath, "Default DLP Policy"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateDLPPoliciesFilterByContains(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_dlp_policies.filtered"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateDLPPoliciesFilter(attr.FilterByContains, "fault"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theDatasource, attr.Len(attr.DLPPolicies), "1"),
					sdk.TestCheckResourceAttr(theDatasource, dlpPolicyNamePath, "Default DLP Policy"),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateDLPPoliciesFilterByRegexp(t *testing.T) {
	t.Parallel()

	theDatasource := "data.twingate_dlp_policies.filtered"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testDatasourceTwingateDLPPoliciesFilter(attr.FilterByRegexp, "^Default.*"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theDatasource, attr.Len(attr.DLPPolicies), "1"),
					sdk.TestCheckResourceAttr(theDatasource, dlpPolicyNamePath, "Default DLP Policy"),
				),
			},
		},
	})
}
*/