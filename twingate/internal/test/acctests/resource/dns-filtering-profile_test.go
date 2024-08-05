package resource

import (
	"fmt"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccDatasourceTwingateDNSFilteringProfileCreate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateDNSFilteringProfile("RG5zRmlsdGVyaW5nUHJvZmlsZTplZWMzYTI3MjQ4"),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("profile_name", "DNS Filtering Profile"),
					resource.TestCheckOutput("profile_priority", "1"),
					resource.TestCheckOutput("profile_fallback_method", "AUTO"),
					//resource.TestCheckOutput("profile_allowed_domains", "[]"),
					//resource.TestCheckOutput("profile_denied_domains", "[]"),
					//resource.TestCheckOutput("profile_groups", "[]"),
				),
			},
		},
	})
}

func testDatasourceTwingateDNSFilteringProfile(profileID string) string {
	return fmt.Sprintf(`
	data "twingate_dns_filtering_profile" "test_profile" {
		id = "%[1]s"
	}

	output "profile_name" {
	  	value = data.twingate_dns_filtering_profile.test_profile.name
	}

	output "profile_priority" {
	  	value = data.twingate_dns_filtering_profile.test_profile.priority
	}

	output "profile_fallback_method" {
	  	value = data.twingate_dns_filtering_profile.test_profile.fallback_method
	}

	output "profile_allowed_domains" {
	  	value = data.twingate_dns_filtering_profile.test_profile.allowed_domains.domains
	}

	output "profile_denied_domains" {
	  	value = data.twingate_dns_filtering_profile.test_profile.denied_domains.domains
	}

	output "profile_groups" {
	  	value = data.twingate_dns_filtering_profile.test_profile.groups
	}
		`, profileID)
}
