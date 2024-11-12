package datasource

import (
	"fmt"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateDNSFilteringProfile_basic(t *testing.T) {
	t.Parallel()

	testName := "t" + acctest.RandString(6)
	profileName := test.RandomName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateDNSFilteringProfile(testName, profileName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("profile_name", profileName),
					resource.TestCheckOutput("profile_priority", "3"),
					resource.TestCheckOutput("profile_fallback_method", "AUTO"),
				),
			},
		},
	})
}

func testDatasourceTwingateDNSFilteringProfile(testName, profileName string) string {
	return fmt.Sprintf(`
	resource "twingate_dns_filtering_profile" "%[1]s" {
	  name = "%[2]s"
	  priority = 3
	  fallback_method = "AUTO"
	}

	data "twingate_dns_filtering_profile" "%[1]s" {
		id = twingate_dns_filtering_profile.%[1]s.id
	}

	output "profile_name" {
	  	value = data.twingate_dns_filtering_profile.%[1]s.name
	}

	output "profile_priority" {
	  	value = data.twingate_dns_filtering_profile.%[1]s.priority
	}

	output "profile_fallback_method" {
	  	value = data.twingate_dns_filtering_profile.%[1]s.fallback_method
	}
		`, testName, profileName)
}
