package datasource

import (
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"regexp"
	"testing"
)

func TestAccDatasourceTwingateDLPPolicy_queryByName(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateDLPPolicy(),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("test_policy", "Test"),
				),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicy() string {
	return `
	data "twingate_dlp_policy" "test" {
	  name = "Test"
	}

	output "test_policy" {
	  value = data.twingate_dlp_policy.test.name
	}
	`
}

func TestAccDatasourceTwingateDLPPolicy_shouldFail(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testDatasourceTwingateDLPPolicyShouldFail(),
				ExpectError: regexp.MustCompile("Error: Invalid Attribute Combination"),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicyShouldFail() string {
	return `
	data "twingate_dlp_policy" "test" {
	  name = "policy-name"
	  id = "policy-id"
	}

	output "test_policy" {
	  value = data.twingate_dlp_policy.test.name
	}
	`
}

func TestAccDatasourceTwingateDLPPolicy_queryByID(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateDLPPolicyQueryByID(),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("test_policy", "Test"),
				),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicyQueryByID() string {
	return `
	data "twingate_dlp_policy" "test" {
	  name = "Test"
	}

	data "twingate_dlp_policy" "test_by_id" {
	  id = data.twingate_dlp_policy.test.id
	}

	output "test_policy" {
	  value = data.twingate_dlp_policy.test_by_id.name
	}
	`
}
