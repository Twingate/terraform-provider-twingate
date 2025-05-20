package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateDLPPolicyInvalidID(t *testing.T) {
	t.Parallel()

	randStr := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testDatasourceTwingateDLPPolicy(randStr),
				ExpectError: regexp.MustCompile("failed to read dlp policy"),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicy(id string) string {
	return fmt.Sprintf(`
	data "twingate_dlp_policy" "test_1" {
	  id = "%s"
	}

	output "dlp_policy_name" {
	  value = data.twingate_dlp_policy.test_1.name
	}
	`, id)
}

func TestAccDatasourceTwingateDLPPolicyReadWithNameAndID(t *testing.T) {
	t.Parallel()

	randStr := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testDatasourceTwingateDLPPolicyWithNameAndID(randStr, randStr),
				ExpectError: regexp.MustCompile("Error: Invalid Attribute Combination"),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicyWithNameAndID(id, name string) string {
	return fmt.Sprintf(`
	data "twingate_dlp_policy" "test_2" {
	  id = "%s"
	  name = "%s"
	}

	output "dlp_policy_name" {
	  value = data.twingate_dlp_policy.test_2.name
	}
	`, id, name)
}

func TestAccDatasourceTwingateDLPPolicyDoesNotExists(t *testing.T) {
	t.Parallel()

	dlpPolicyID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("DLPPolicy:%d", acctest.RandInt())))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testDatasourceTwingateDLPPolicy(dlpPolicyID),
				ExpectError: regexp.MustCompile("failed to read dlp policy with id"),
			},
		},
	})
}

// Note: The following tests are commented out since we don't have access to real DLP policies in the test environment.
// They would follow the same pattern as the security policy tests.

/*
func TestAccDatasourceTwingateDLPPolicyReadOkByID(t *testing.T) {
	t.Parallel()

	dlpPolicies, err := acctests.ListDLPPolicies()
	if err != nil {
		t.Skip("can't run test:", err)
	}

	testPolicy := dlpPolicies[0]

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateDLPPolicyByID(testPolicy.ID),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("dlp_policy_name", testPolicy.Name),
				),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicyByID(id string) string {
	return fmt.Sprintf(`
	data "twingate_dlp_policy" "test" {
	  id = "%s"
	}

	output "dlp_policy_name" {
	  value = data.twingate_dlp_policy.test.name
	}
	`, id)
}

func TestAccDatasourceTwingateDLPPolicyReadOkByName(t *testing.T) {
	t.Parallel()

	dlpPolicies, err := acctests.ListDLPPolicies()
	if err != nil {
		t.Skip("can't run test:", err)
	}

	testPolicy := dlpPolicies[0]

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateDLPPolicyByName(testPolicy.Name),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("dlp_policy_id", testPolicy.ID),
				),
			},
		},
	})
}

func testDatasourceTwingateDLPPolicyByName(name string) string {
	return fmt.Sprintf(`
	data "twingate_dlp_policy" "test" {
	  name = "%s"
	}

	output "dlp_policy_id" {
	  value = data.twingate_dlp_policy.test.id
	}
	`, name)
}
*/