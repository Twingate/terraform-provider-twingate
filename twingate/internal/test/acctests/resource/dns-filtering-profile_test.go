package resource

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	groupsLen = attr.Len(attr.Groups)
)

func TestAccTwingateDNSFilteringProfileCreateWithDefaultValues(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateDNSProfileDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfileBase(testName, profileName, "2"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.Priority, "2"),
					sdk.TestCheckResourceAttr(theResource, attr.FallbackMethod, "STRICT"),
					sdk.TestCheckResourceAttr(theResource, groupsLen, "0"),
				),
			},
		},
	})
}

func testTwingateDNSFilteringProfileBase(testName, profileName, priority string) string {
	return fmt.Sprintf(`
	resource "twingate_dns_filtering_profile" "%[1]s" {
	  name = "%[2]s"
	  priority = "%[3]s"
	}
	`, testName, profileName, priority)

}

func genDomains(count int) []string {
	domains := make([]string, 0, count)

	for i := 0; i < count; i++ {
		domains = append(domains, test.RandomDomain())
	}

	return domains
}

func listToString(list []string) string {
	if len(list) == 0 {
		return "[]"
	}

	return fmt.Sprintf(`["%s"]`, strings.Join(list, `", "`))
}

func TestAccTwingateDNSFilteringProfileCreate(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	groups, groupResources := genNewGroupsWithName(testName, testName, 3)
	groupsTF := strings.Join(groups, "\n")
	groupResourcesTF := fmt.Sprintf(`"%s"`, strings.Join(groupResources, `", "`))

	allowedDomains := genDomains(2)
	deniedDomains := genDomains(1)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateDNSProfileDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfile(groupsTF, testName, profileName, groupResourcesTF, allowedDomains, deniedDomains),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.Priority, "2"),
					sdk.TestCheckResourceAttr(theResource, attr.FallbackMethod, "AUTO"),
					sdk.TestCheckResourceAttr(theResource, groupsLen, "3"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.AllowedDomains, attr.IsAuthoritative), "false"),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.DeniedDomains, attr.IsAuthoritative), "true"),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.DeniedDomains, attr.Domains), "1"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.ContentCategories, attr.BlockAdultContent), "true"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.SecurityCategories, attr.BlockDNSRebinding), "false"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.SecurityCategories, attr.BlockNewlyRegisteredDomains), "false"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.PrivacyCategories, attr.BlockDisguisedTrackers), "true"),
				),
			},
		},
	})
}

func testTwingateDNSFilteringProfile(groups, testName, profileName, groupResources string, allowedDomains, deniedDomains []string) string {
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
		domains = %[4]s
	  }
	
	  denied_domains {
		is_authoritative = true
		domains = %[5]s
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

	  depends_on = [%[6]s]
	}

	`, groups, testName, profileName, listToString(allowedDomains), listToString(deniedDomains), groupResources)

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

func TestAccTwingateDNSFilteringProfileUpdate(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	groups1, groupResources1 := genNewGroupsWithName(testName, testName, 2)
	groupsTF1 := strings.Join(groups1, "\n")
	groupResourcesTF1 := fmt.Sprintf(`"%s"`, strings.Join(groupResources1, `", "`))

	groups2, groupResources2 := genNewGroupsWithName(testName, testName, 3)
	groupsTF2 := strings.Join(groups2, "\n")
	groupResourcesTF2 := fmt.Sprintf(`"%s"`, strings.Join(groupResources2, `", "`))

	domains1 := genDomains(2)
	domains2 := genDomains(3)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateDNSProfileDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfile1(groupsTF1, testName, profileName, groupResourcesTF1, "3", "AUTO", true, domains1, true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.Priority, "3"),
					sdk.TestCheckResourceAttr(theResource, attr.FallbackMethod, "AUTO"),
					sdk.TestCheckResourceAttr(theResource, groupsLen, "2"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.AllowedDomains, attr.IsAuthoritative), "true"),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.PrivacyCategories, attr.BlockDisguisedTrackers), "true"),
				),
			},
			{
				Config: testTwingateDNSFilteringProfile1(groupsTF2, testName, profileName, groupResourcesTF2, "2.5", "STRICT", true, domains2, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.Priority, "2.5"),
					sdk.TestCheckResourceAttr(theResource, attr.FallbackMethod, "STRICT"),
					sdk.TestCheckResourceAttr(theResource, groupsLen, "3"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.AllowedDomains, attr.IsAuthoritative), "true"),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "3"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.PrivacyCategories, attr.BlockDisguisedTrackers), "false"),
				),
			},
		},
	})
}

func testTwingateDNSFilteringProfile1(groups, testName, profileName, groupResources, priority, fallBack string, isAuthoritative bool, domains []string, blockDisguisedTrackers bool) string {
	return fmt.Sprintf(`
	# groups
	%[1]s

	resource "twingate_dns_filtering_profile" "%[2]s" {
	  name = "%[3]s"
	  priority = %[4]s
	  fallback_method = "%[5]s"
	  groups = toset(data.twingate_groups.test.groups[*].id)
	
	  allowed_domains {
		is_authoritative = %[6]v
		domains = ["%[7]s"]
	  }
	
	  privacy_categories {
		block_disguised_trackers = %[8]v
	  }
	
	}
	
	data "twingate_groups" "test" {
	  name_prefix = "%[2]s"

	  depends_on = [%[9]s]
	}

	`, groups, testName, profileName, priority, fallBack, isAuthoritative, strings.Join(domains, `", "`), blockDisguisedTrackers, groupResources)

}

func TestAccTwingateDNSFilteringProfileUpdateIsAuthoritativeTrue(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	domains1 := genDomains(2)
	newDomains := genDomains(1)
	domains2 := genDomains(4)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateDNSProfileDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, true, domains1),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
					acctests.WaitTestFunc(),
					// set new domains to the DNS profile using API
					acctests.AddDNSProfileAllowedDomains(theResource, newDomains),
				),
				// expecting drift - terraform going to remove unknown domains
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, true, domains2),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "4"),
				),
			},
		},
	})
}

func testTwingateDNSFilteringProfileWithDomains(testName, profileName string, isAuthoritative bool, domains []string) string {
	return fmt.Sprintf(`
	resource "twingate_dns_filtering_profile" "%[1]s" {
	  name = "%[2]s"
	  priority = 5
	
	  allowed_domains {
		is_authoritative = %[3]v
		domains = ["%[4]s"]
	  }
	}
	`, testName, profileName, isAuthoritative, strings.Join(domains, `", "`))

}

func TestAccTwingateDNSFilteringProfileUpdateIsAuthoritativeFalse(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	domains1 := genDomains(2)
	newDomains := genDomains(1)
	domains2 := genDomains(2)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateDNSProfileDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, false, domains1),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
					acctests.WaitTestFunc(),
					// set new domains to the DNS profile using API
					acctests.AddDNSProfileAllowedDomains(theResource, newDomains),
				),
			},
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, false, domains2),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
					// check allowed domains using API
					acctests.CheckDNSProfileAllowedDomainsLen(theResource, 3),
				),
			},
		},
	})
}

func TestAccTwingateDNSFilteringProfileUpdateIsAuthoritativeFalseTrue(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	domains1 := genDomains(2)
	domains2 := genDomains(4)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateDNSProfileDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, false, domains1),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
				),
			},
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, true, domains2),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "4"),
					// check allowed domains using API
					acctests.CheckDNSProfileAllowedDomainsLen(theResource, 4),
				),
			},
		},
	})
}

func TestAccTwingateDNSFilteringProfileUpdateIsAuthoritativeTrueFalse(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	domains1 := genDomains(2)
	domains2 := genDomains(4)
	newDomains := genDomains(5)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateDNSProfileDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, true, domains1),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
				),
			},
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, false, domains2),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "4"),
					// check allowed domains using API
					acctests.CheckDNSProfileAllowedDomainsLen(theResource, 6),
					acctests.WaitTestFunc(),
					// set new domains to the DNS profile using API
					acctests.AddDNSProfileAllowedDomains(theResource, newDomains),
				),
			},
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, false, domains2),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "4"),
					// check allowed domains using API
					acctests.CheckDNSProfileAllowedDomainsLen(theResource, 5),
				),
			},
		},
	})
}

func TestAccTwingateDNSFilteringProfileRemoveAllowedDomains(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	domains1 := genDomains(2)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateDNSProfileDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfileWithDomains(testName, profileName, true, domains1),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "2"),
				),
			},
			{
				Config: testTwingateDNSFilteringProfileBase(testName, profileName, "5"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, profileName),
					sdk.TestCheckResourceAttr(theResource, attr.LenAttr(attr.AllowedDomains, attr.Domains), "0"),
					// check allowed domains using API
					acctests.CheckDNSProfileAllowedDomainsLen(theResource, 0),
				),
			},
		},
	})
}

func TestAccTwingateDNSFilteringProfileImport(t *testing.T) {
	testName := "t" + acctest.RandString(6)
	theResource := acctests.TerraformDNSFilteringProfile(testName)
	profileName := test.RandomName(testName)

	groups, groupResources := genNewGroupsWithName(testName, testName, 3)
	groupsTF := strings.Join(groups, "\n")
	groupResourcesTF := fmt.Sprintf(`"%s"`, strings.Join(groupResources, `", "`))

	allowedDomains := genDomains(2)
	deniedDomains := genDomains(1)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: testTwingateDNSFilteringProfileFull(groupsTF, testName, profileName, groupResourcesTF, allowedDomains, deniedDomains),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				ImportState:  true,
				ResourceName: theResource,
				ImportStateCheck: acctests.CheckImportState(map[string]string{
					attr.Name:                 profileName,
					attr.Priority:             "3",
					attr.FallbackMethod:       "AUTO",
					attr.LenAttr(attr.Groups): "3",

					attr.LenAttr(attr.AllowedDomains, attr.Domains): "2",
					// we can't get this value from backend, and by default we have `true`
					attr.PathAttr(attr.AllowedDomains, attr.IsAuthoritative): "true",

					attr.LenAttr(attr.DeniedDomains, attr.Domains):          "1",
					attr.PathAttr(attr.DeniedDomains, attr.IsAuthoritative): "true",

					attr.PathAttr(attr.ContentCategories, attr.BlockAdultContent):           "true",
					attr.PathAttr(attr.ContentCategories, attr.BlockGambling):               "false",
					attr.PathAttr(attr.ContentCategories, attr.BlockDating):                 "false",
					attr.PathAttr(attr.ContentCategories, attr.BlockSocialMedia):            "false",
					attr.PathAttr(attr.ContentCategories, attr.BlockGames):                  "false",
					attr.PathAttr(attr.ContentCategories, attr.BlockStreaming):              "false",
					attr.PathAttr(attr.ContentCategories, attr.BlockPiracy):                 "false",
					attr.PathAttr(attr.ContentCategories, attr.EnableYoutubeRestrictedMode): "false",
					attr.PathAttr(attr.ContentCategories, attr.EnableSafesearch):            "false",

					attr.PathAttr(attr.SecurityCategories, attr.EnableThreatIntelligenceFeeds):   "true",
					attr.PathAttr(attr.SecurityCategories, attr.EnableGoogleSafeBrowsing):        "true",
					attr.PathAttr(attr.SecurityCategories, attr.BlockCryptojacking):              "true",
					attr.PathAttr(attr.SecurityCategories, attr.BlockIdnHomoglyph):               "true",
					attr.PathAttr(attr.SecurityCategories, attr.BlockTyposquatting):              "true",
					attr.PathAttr(attr.SecurityCategories, attr.BlockDNSRebinding):               "false",
					attr.PathAttr(attr.SecurityCategories, attr.BlockNewlyRegisteredDomains):     "false",
					attr.PathAttr(attr.SecurityCategories, attr.BlockDomainGenerationAlgorithms): "true",
					attr.PathAttr(attr.SecurityCategories, attr.BlockParkedDomains):              "true",

					attr.PathAttr(attr.PrivacyCategories, attr.BlockAffiliateLinks):    "false",
					attr.PathAttr(attr.PrivacyCategories, attr.BlockDisguisedTrackers): "true",
					attr.PathAttr(attr.PrivacyCategories, attr.BlockAdsAndTrackers):    "false",
				}),
			},
		},
	})
}

func testTwingateDNSFilteringProfileFull(groups, testName, profileName, groupResources string, allowedDomains, deniedDomains []string) string {
	return fmt.Sprintf(`
	# groups
	%[1]s

	resource "twingate_dns_filtering_profile" "%[2]s" {
	  name = "%[3]s"
	  priority = 3
	  fallback_method = "AUTO"
	  groups = toset(data.twingate_groups.test.groups[*].id)
	
	  allowed_domains {
		is_authoritative = false
		domains = %[4]s
	  }
	
	  denied_domains {
		is_authoritative = true
		domains = %[5]s
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

	  depends_on = [%[6]s]
	}

	`, groups, testName, profileName, listToString(allowedDomains), listToString(deniedDomains), groupResources)

}
