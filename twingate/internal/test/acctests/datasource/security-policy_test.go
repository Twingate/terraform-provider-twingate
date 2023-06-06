package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateSecurityPolicyInvalidID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Security Policy - failed parse ID", func(t *testing.T) {
		randStr := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testDatasourceTwingateSecurityPolicy(randStr),
					ExpectError: regexp.MustCompile("failed to read security policy with id"),
				},
			},
		})
	})
}

func testDatasourceTwingateSecurityPolicy(id string) string {
	return fmt.Sprintf(`
	data "twingate_security_policy" "test_1" {
	  id = "%s"
	}

	output "security_policy_name" {
	  value = data.twingate_security_policy.test_1.name
	}
	`, id)
}

func TestAccDatasourceTwingateSecurityPolicyReadWithNameAndID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Security Policy - read with name and id", func(t *testing.T) {
		randStr := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testDatasourceTwingateSecurityPolicyWithNameAndID(randStr, randStr),
					ExpectError: regexp.MustCompile("invalid combination of arguments"),
				},
			},
		})
	})
}

func testDatasourceTwingateSecurityPolicyWithNameAndID(id, name string) string {
	return fmt.Sprintf(`
	data "twingate_security_policy" "test_2" {
	  id = "%s"
	  name = "%s"
	}

	output "security_policy_name" {
	  value = data.twingate_security_policy.test_2.name
	}
	`, id, name)
}

func TestAccDatasourceTwingateSecurityPolicyDoesNotExists(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Security Policy - does not exists", func(t *testing.T) {
		securityPolicyID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("SecurityPolicy:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testDatasourceTwingateSecurityPolicy(securityPolicyID),
					ExpectError: regexp.MustCompile("failed to read security policy with id"),
				},
			},
		})
	})
}

func TestAccDatasourceTwingateSecurityPolicyReadOkByID(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Security Policy - read Ok By ID", func(t *testing.T) {
		securityPolicies, err := acctests.ListSecurityPolicies()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		testPolicy := securityPolicies[0]

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateSecurityPolicyByID(testPolicy.ID),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckOutput("security_policy_name", testPolicy.Name),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateSecurityPolicyByID(id string) string {
	return fmt.Sprintf(`
	data "twingate_security_policy" "test" {
	  id = "%s"
	}

	output "security_policy_name" {
	  value = data.twingate_security_policy.test.name
	}
	`, id)
}

func TestAccDatasourceTwingateSecurityPolicyReadOkByName(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Security Policy - read Ok By Name", func(t *testing.T) {
		securityPolicies, err := acctests.ListSecurityPolicies()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		testPolicy := securityPolicies[0]

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateSecurityPolicyByName(testPolicy.Name),
					Check: acctests.ComposeTestCheckFunc(
						resource.TestCheckOutput("security_policy_id", testPolicy.ID),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateSecurityPolicyByName(name string) string {
	return fmt.Sprintf(`
	data "twingate_security_policy" "test" {
	  name = "%s"
	}

	output "security_policy_id" {
	  value = data.twingate_security_policy.test.id
	}
	`, name)
}
