package resource

import (
	"fmt"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strings"
	"testing"
)

var (
	groupsLen = attr.Len(attr.Groups)
)

func TestAccDatasourceTwingateDNSFilteringProfileCreate(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName()

	groups, groupResources := genNewGroupsWithName(testName, testName, 3)
	groupsTF := strings.Join(groups, "\n")
	groupResourcesTF := fmt.Sprintf(`"%s"`, strings.Join(groupResources, `", "`))

	cfg := testDatasourceTwingateDNSFilteringProfile(groupsTF, testName, profileName, groupResourcesTF)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateConnectorDestroy,
		Steps: []sdk.TestStep{
			{
				Config: cfg,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.Priority, "2"),
					sdk.TestCheckResourceAttr(theResource, attr.FallbackMethod, "AUTO"),
					sdk.TestCheckResourceAttr(theResource, groupsLen, "3"),
					sdk.TestCheckResourceAttr(theResource, groupsLen, "3"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.AllowedDomains, attr.IsAuthoritative), "false"),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.DeniedDomains, attr.IsAuthoritative), "true"),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.DeniedDomains, attr.Domains), "1"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.ContentCategories, attr.BlockAdultContent), "true"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.SecurityCategories, attr.BlockDnsRebinding), "false"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.SecurityCategories, attr.BlockNewlyRegisteredDomains), "false"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.PrivacyCategories, attr.BlockDisguisedTrackers), "true"),
				),
			},
		},
	})
}

func testDatasourceTwingateDNSFilteringProfile(groups, testName, profileName, groupResources string) string {
	return fmt.Sprintf(`
	# groups
	%[1]s

	resource "twingate_dns_filtering_profile" "%[2]s" {
	  name = "%[3]s"
	  priority = 2
	  fallback_method = "AUTO"
	  groups = toset(data.twingate_groups.test.groups[*].id)
	
	  allowed_domains {
		is_authoritative = false
		domains = [
		  "twingate.com",
		  "zoom.us"
		]
	  }
	
	  denied_domains {
		is_authoritative = true
		domains = [
		  "evil.example"
		]
	  }
	
	  content_categories {
		block_adult_content = true
	  }
	
	  security_categories {
		block_dns_rebinding = false
		block_newly_registered_domains = false
	  }
	
	  privacy_categories {
		block_disguised_trackers = true
	  }
	
	}
	
	data "twingate_groups" "test" {
	  name_prefix = "%[2]s"

	  depends_on = [%[4]s]
	}

	`, groups, testName, profileName, groupResources)

}

func genNewGroupsWithName(resourcePrefix, namePrefix string, count int) ([]string, []string) {
	groups := make([]string, 0, count)
	groupsResources := make([]string, 0, count)

	for i := 0; i < count; i++ {
		resourceName := fmt.Sprintf("%s_%d", resourcePrefix, i+1)
		groupName := fmt.Sprintf("%s_%d", namePrefix, i+1)
		groups = append(groups, newTerraformGroup(resourceName, groupName))
		groupsResources = append(groupsResources, fmt.Sprintf("twingate_group.%s", resourceName))
	}

	return groups, groupsResources
}
