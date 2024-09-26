package datasource

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDatasourceTwingateDLPPolicies_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateDLPPolicies(),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("test_policy", "Test"),
				),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicies() string {
	return `
	data "twingate_dlp_policies" "test" {
	  name_prefix = "Te"
	}

	output "test_policy" {
	  value = data.twingate_dlp_policies.test.dlp_policies[0].name
	}
	`
}
