package resource

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

const (
	groupIdsLen  = "group_ids.#"
	tcpPolicy    = "protocols.0.tcp.0.policy"
	udpPolicy    = "protocols.0.udp.0.policy"
	firstTCPPort = "protocols.0.tcp.0.ports.0"
	firstUDPPort = "protocols.0.udp.0.ports.0"
	tcpPortsLen  = "protocols.0.tcp.0.ports.#"
	udpPortsLen  = "protocols.0.udp.0.ports.#"

	accessGroupIdsLen          = "access.0.group_ids.#"
	accessServiceAccountIdsLen = "access.0.service_account_ids.#"

	nameAttr        = "name"
	addressAttr     = "address"
	accessTokenAttr = "access_token"

	isVisibleAttr                = "is_visible"
	isBrowserShortcutEnabledAttr = "is_browser_shortcut_enabled"
)

func TestAccTwingateResourceCreate(t *testing.T) {
	const terraformResourceName = "test1"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceOnlyWithNetwork(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, groupIdsLen),
				),
			},
		},
	})
}

func createResourceOnlyWithNetwork(terraformResourceName, networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_resource" "%s" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%s.id
	}
	`, terraformResourceName, networkName, terraformResourceName, resourceName, terraformResourceName)
}

func TestAccTwingateResourceCreateWithProtocolsAndGroups(t *testing.T) {
	const theResource = "twingate_resource.test2"
	remoteNetworkName := test.RandomName()
	groupName1 := test.RandomGroupName()
	groupName2 := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithProtocolsAndGroups(remoteNetworkName, groupName1, groupName2, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, addressAttr, "new-acc-test.com"),
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "2"),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
					sdk.TestCheckResourceAttr(theResource, firstTCPPort, "80"),
				),
			},
		},
	})
}

func createResourceWithProtocolsAndGroups(networkName, groupName1, groupName2, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test2" {
	  name = "%s"
	}

    resource "twingate_group" "g21" {
      name = "%s"
    }

    resource "twingate_group" "g22" {
      name = "%s"
    }

	resource "twingate_resource" "test2" {
	  name = "%s"
	  address = "new-acc-test.com"
	  remote_network_id = twingate_remote_network.test2.id
	  group_ids = [twingate_group.g21.id, twingate_group.g22.id]
      protocols {
		allow_icmp = true
        tcp  {
			policy = "%s"
            ports = ["80", "82-83"]
        }
		udp {
 			policy = "%s"
		}
      }
	}
	`, networkName, groupName1, groupName2, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceFullCreationFlow(t *testing.T) {
	const theResource = "twingate_resource.test3"
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: resourceFullCreationFlow(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr("twingate_remote_network.test3", nameAttr, remoteNetworkName),
					sdk.TestCheckResourceAttr(theResource, nameAttr, resourceName),
					sdk.TestMatchResourceAttr("twingate_connector_tokens.test31", accessTokenAttr, regexp.MustCompile(".+")),
				),
			},
		},
	})
}

func resourceFullCreationFlow(networkName, groupName, resourceName string) string {
	return fmt.Sprintf(`
    resource "twingate_remote_network" "test3" {
      name = "%s"
    }
	
    resource "twingate_connector" "test31" {
      remote_network_id = twingate_remote_network.test3.id
    }

    resource "twingate_connector_tokens" "test31" {
      connector_id = twingate_connector.test31.id
    }

    resource "twingate_connector" "test32" {
      remote_network_id = twingate_remote_network.test3.id
    }
	
    resource "twingate_connector_tokens" "test32" {
      connector_id = twingate_connector.test32.id
    }

    resource "twingate_group" "test3" {
      name = "%s"
    }

    resource "twingate_resource" "test3" {
      name = "%s"
      address = "acc-test.com"
      remote_network_id = twingate_remote_network.test3.id
      group_ids = [twingate_group.test3.id]
      protocols {
        allow_icmp = true
        tcp  {
            policy = "%s"
            ports = ["3306"]
        }
        udp {
            policy = "%s"
        }
      }
    }
    `, networkName, groupName, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceWithInvalidGroupId(t *testing.T) {
	resourceName := test.RandomResourceName()
	networkName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      createResourceWithInvalidGroupId(networkName, resourceName),
				ExpectError: regexp.MustCompile("Error: failed to create resource"),
			},
		},
	})
}

func createResourceWithInvalidGroupId(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test4" {
	  name = "%s"
	}

	resource "twingate_resource" "test4" {
	  name = "%s"
	  address = "acc-test.com"
	  group_ids = ["foo", "bar"]
	  remote_network_id = twingate_remote_network.test4.id
	}
	`, networkName, resourceName)
}

func TestAccTwingateResourceWithTcpDenyAllPolicy(t *testing.T) {
	const theResource = "twingate_resource.test5"
	resourceName := test.RandomResourceName()
	networkName := test.RandomResourceName()
	groupName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithTcpDenyAllPolicy(networkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
				),
			},
			// expecting no changes - empty plan
			{
				Config:   createResourceWithTcpDenyAllPolicy(networkName, groupName, resourceName),
				PlanOnly: true,
			},
		},
	})
}

func createResourceWithTcpDenyAllPolicy(networkName, groupName, resourceName string) string {
	return fmt.Sprintf(`
    resource "twingate_remote_network" "test5" {
      name = "%s"
    }

    resource "twingate_group" "g5" {
      name = "%s"
    }

    resource "twingate_resource" "test5" {
      name = "%s"
      address = "new-acc-test.com"
      remote_network_id = twingate_remote_network.test5.id
      group_ids = [twingate_group.g5.id]
      protocols {
        allow_icmp = true
        tcp {
          policy = "%s"
        }
        udp {
          policy = "%s"
        }
      }
    }
    `, networkName, groupName, resourceName, model.PolicyDenyAll, model.PolicyAllowAll)
}

func TestAccTwingateResourceWithUdpDenyAllPolicy(t *testing.T) {
	const theResource = "twingate_resource.test6"
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, udpPolicy, model.PolicyRestricted),
				),
			},
			// expecting no changes - empty plan
			{
				Config:   createResourceWithUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				PlanOnly: true,
			},
		},
	})
}

func createResourceWithUdpDenyAllPolicy(networkName, groupName, resourceName string) string {
	return fmt.Sprintf(`
    resource "twingate_remote_network" "test6" {
      name = "%s"
    }

    resource "twingate_group" "g6" {
      name = "%s"
    }

    resource "twingate_resource" "test6" {
      name = "%s"
      address = "acc-test.com"
      remote_network_id = twingate_remote_network.test6.id
      group_ids = [twingate_group.g6.id]
      protocols {
        allow_icmp = true
        tcp {
          policy = "%s"
        }
        udp {
          policy = "%s"
        }
      }
    }
	`, networkName, groupName, resourceName, model.PolicyAllowAll, model.PolicyDenyAll)
}

func TestAccTwingateResourceWithRestrictedPolicyAndEmptyPortsList(t *testing.T) {
	const theResource = "twingate_resource.test7"
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithRestrictedPolicyAndEmptyPortsList(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, nameAttr, resourceName),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
					sdk.TestCheckNoResourceAttr(theResource, tcpPortsLen),
					sdk.TestCheckResourceAttr(theResource, udpPolicy, model.PolicyRestricted),
					sdk.TestCheckNoResourceAttr(theResource, udpPortsLen),
				),
			},
		},
	})
}

func createResourceWithRestrictedPolicyAndEmptyPortsList(networkName, groupName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test7" {
	  name = "%s"
	}

    resource "twingate_group" "test7" {
      name = "%s"
    }

	resource "twingate_resource" "test7" {
	  name = "%s"
	  address = "new-acc-test.com"
	  remote_network_id = twingate_remote_network.test7.id
	  group_ids = [twingate_group.test7.id]
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = []
	    }
	    udp {
	      policy = "%s"
	    }
	  }
	}
	`, networkName, groupName, resourceName, model.PolicyRestricted, model.PolicyRestricted)
}

func TestAccTwingateResourceWithInvalidPortRange(t *testing.T) {
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	expectedError := regexp.MustCompile("Error: failed to parse protocols port range")

	genConfig := func(portRange string) string {
		return createResourceWithRestrictedPolicyAndPortRange(remoteNetworkName, resourceName, portRange)
	}

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      genConfig(`""`),
				ExpectError: expectedError,
			},
			{
				Config:      genConfig(`" "`),
				ExpectError: expectedError,
			},
			{
				Config:      genConfig(`"foo"`),
				ExpectError: expectedError,
			},
			{
				Config:      genConfig(`"80-"`),
				ExpectError: expectedError,
			},
			{
				Config:      genConfig(`"-80"`),
				ExpectError: expectedError,
			},
			{
				Config:      genConfig(`"80-90-100"`),
				ExpectError: expectedError,
			},
			{
				Config:      genConfig(`"80-70"`),
				ExpectError: expectedError,
			},
			{
				Config:      genConfig(`"0-65536"`),
				ExpectError: expectedError,
			},
		},
	})
}

func createResourceWithRestrictedPolicyAndPortRange(networkName, resourceName, portRange string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test8" {
	  name = "%s"
	}

	resource "twingate_resource" "test8" {
	  name = "%s"
	  address = "new-acc-test.com"
	  remote_network_id = twingate_remote_network.test8.id
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = [%s]
	    }
	    udp {
	      policy = "%s"
	    }
	  }
	}
	`, networkName, resourceName, model.PolicyRestricted, portRange, model.PolicyAllowAll)
}

func TestAccTwingateResourcePortReorderingCreatesNoChanges(t *testing.T) {
	const theResource = "twingate_resource.test9"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"82-83", "80"`),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, firstTCPPort, "80"),
					sdk.TestCheckResourceAttr(theResource, firstUDPPort, "80"),
				),
			},
			// no changes
			{
				Config:   createResourceWithPortRange(remoteNetworkName, resourceName, `"82-83", "80"`),
				PlanOnly: true,
			},
			// no changes
			{
				Config:   createResourceWithPortRange(remoteNetworkName, resourceName, `"82", "83", "80"`),
				PlanOnly: true,
			},
			// new changes applied
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"82-83", "70"`),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, firstTCPPort, "70"),
					sdk.TestCheckResourceAttr(theResource, firstUDPPort, "70"),
				),
			},
		},
	})
}

func createResourceWithPortRange(networkName, resourceName, portRange string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test9" {
	  name = "%s"
	}

	resource "twingate_resource" "test9" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test9.id
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = [%s]
	    }
	    udp {
	      policy = "%s"
	      ports = [%s]
	    }
	  }
	}
	`, networkName, resourceName, model.PolicyRestricted, portRange, model.PolicyRestricted, portRange)
}

func TestAccTwingateResourceSetActiveStateOnUpdate(t *testing.T) {
	const terraformResourceName = "test10"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceOnlyWithNetwork(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeactivateTwingateResource(theResource),
					acctests.WaitTestFunc(),
					acctests.CheckTwingateResourceActiveState(theResource, false),
				),
			},
			{
				Config: createResourceOnlyWithNetwork(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceActiveState(theResource, true),
				),
			},
		},
	})
}

func TestAccTwingateResourceReCreateAfterDeletion(t *testing.T) {
	const terraformResourceName = "test10"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceOnlyWithNetwork(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateResource),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResourceOnlyWithNetwork(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func TestAccTwingateResourceImport(t *testing.T) {
	const theResource = "twingate_resource.test12"
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	groupName2 := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource12(remoteNetworkName, groupName, groupName2, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				ImportState:  true,
				ResourceName: theResource,
				ImportStateCheck: acctests.CheckImportState(map[string]string{
					addressAttr:  "acc-test.com.12",
					tcpPolicy:    model.PolicyRestricted,
					tcpPortsLen:  "2",
					firstTCPPort: "80",
					udpPolicy:    model.PolicyAllowAll,
					udpPortsLen:  "0",
				}),
			},
		},
	})
}

func createResource12(networkName, groupName1, groupName2, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test12" {
	  name = "%s"
	}

    resource "twingate_group" "g121" {
      name = "%s"
    }

    resource "twingate_group" "g122" {
      name = "%s"
    }

	resource "twingate_resource" "test12" {
	  name = "%s"
	  address = "acc-test.com.12"
	  remote_network_id = twingate_remote_network.test12.id
	  group_ids = [twingate_group.g121.id, twingate_group.g122.id]
      protocols {
		allow_icmp = true
        tcp  {
			policy = "%s"
            ports = ["80", "82-83"]
        }
		udp {
 			policy = "%s"
		}
      }
	}
	`, networkName, groupName1, groupName2, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func genNewGroups(resourcePrefix string, count int) ([]string, []string) {
	groups := make([]string, 0, count)
	groupsID := make([]string, 0, count)

	for i := 0; i < count; i++ {
		resourceName := fmt.Sprintf("%s_%d", resourcePrefix, i+1)
		groups = append(groups, newTerraformGroup(resourceName, test.RandomName()))
		groupsID = append(groupsID, fmt.Sprintf("twingate_group.%s.id", resourceName))
	}

	return groups, groupsID
}

func getResourceNameFromID(resourceID string) string {
	idx := strings.LastIndex(resourceID, ".id")
	if idx == -1 {
		return ""
	}

	return resourceID[:idx]
}

func genNewServiceAccounts(resourcePrefix string, count int) ([]string, []string) {
	serviceAccounts := make([]string, 0, count)
	serviceAccountIDs := make([]string, 0, count)

	for i := 0; i < count; i++ {
		resourceName := fmt.Sprintf("%s_%d", resourcePrefix, i+1)
		serviceAccounts = append(serviceAccounts, createServiceAccount(resourceName, test.RandomName()))
		serviceAccountIDs = append(serviceAccountIDs, acctests.TerraformServiceAccount(resourceName)+".id")
	}

	return serviceAccounts, serviceAccountIDs
}

func newTerraformGroup(resourceName, groupName string) string {
	return fmt.Sprintf(`
    resource "twingate_group" "%s" {
      name = "%s"
    }
	`, resourceName, groupName)
}

func TestAccTwingateResourceAddAccessGroups(t *testing.T) {
	const theResource = "twingate_resource.test14"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g14", 2)

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource14(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "2"),
				),
			},
		},
	})
}

func createResource14(networkName, resourceName string, groups, groupsID []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test14" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test14" {
	  name = "%s"
	  address = "acc-test.com.14"
	  remote_network_id = twingate_remote_network.test14.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  access {
	    group_ids = [%s]
	  }

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func TestAccTwingateResourceAddAccessServiceAccounts(t *testing.T) {
	const theResource = "twingate_resource.test15"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource15(remoteNetworkName, resourceName, createServiceAccount(resourceName, "s15")),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
				),
			},
		},
	})
}

func createResource15(networkName, resourceName string, terraformServiceAccount string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test15" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test15" {
	  name = "%s"
	  address = "acc-test.com.15"
	  remote_network_id = twingate_remote_network.test15.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  access {
	    service_account_ids = [%s]
	  }

	}
	`, networkName, terraformServiceAccount, resourceName, model.PolicyRestricted, model.PolicyAllowAll, acctests.TerraformServiceAccount(resourceName)+".id")
}

func TestAccTwingateResourceAddAccessGroupsAndServiceAccounts(t *testing.T) {
	const theResource = "twingate_resource.test16"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g16", 1)

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource16(remoteNetworkName, resourceName, groups, groupsID, createServiceAccount(resourceName, "s16")),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
				),
			},
		},
	})
}

func createResource16(networkName, resourceName string, groups, groupsID []string, terraformServiceAccount string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test16" {
	  name = "%s"
	}

	%s

	%s

	resource "twingate_resource" "test16" {
	  name = "%s"
	  address = "acc-test.com.16"
	  remote_network_id = twingate_remote_network.test16.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  access {
	    group_ids = [%s]
	    service_account_ids = [%s]
	  }

	}
	`, networkName, strings.Join(groups, "\n"), terraformServiceAccount, resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "), acctests.TerraformServiceAccount(resourceName)+".id")
}

func TestAccTwingateResourceAccessServiceAccountsNotAuthoritative(t *testing.T) {
	const theResource = "twingate_resource.test17"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	serviceAccounts, serviceAccountIDs := genNewServiceAccounts("s17", 3)

	serviceAccountResource := getResourceNameFromID(serviceAccountIDs[2])

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource17(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added new service account to the resource though API
					acctests.AddResourceServiceAccount(theResource, serviceAccountResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceServiceAccountsLen(theResource, 2),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource17(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.CheckResourceServiceAccountsLen(theResource, 2),
				),
			},
			{
				// added new service account to the resource though terraform
				Config: createResource17(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:2]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "2"),
					acctests.CheckResourceServiceAccountsLen(theResource, 3),
				),
			},
			{
				// remove one service account from the resource though terraform
				Config: createResource17(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.CheckResourceServiceAccountsLen(theResource, 2),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource17(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.CheckResourceServiceAccountsLen(theResource, 2),
					// delete service account from the resource though API
					acctests.DeleteResourceServiceAccount(theResource, serviceAccountResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceServiceAccountsLen(theResource, 1),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource17(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.CheckResourceServiceAccountsLen(theResource, 1),
				),
			},
		},
	})
}

func createResource17(networkName, resourceName string, serviceAccounts, serviceAccountIDs []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test17" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test17" {
	  name = "%s"
	  address = "acc-test.com.17"
	  remote_network_id = twingate_remote_network.test17.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  is_authoritative = false
	  access {
	    service_account_ids = [%s]
	  }

	}
	`, networkName, strings.Join(serviceAccounts, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(serviceAccountIDs, ", "))
}

func TestAccTwingateResourceAccessServiceAccountsAuthoritative(t *testing.T) {
	const theResource = "twingate_resource.test13"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	serviceAccounts, serviceAccountIDs := genNewServiceAccounts("s13", 3)

	serviceAccountResource := getResourceNameFromID(serviceAccountIDs[2])

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added new service account to the resource though API
					acctests.AddResourceServiceAccount(theResource, serviceAccountResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceServiceAccountsLen(theResource, 2),
				),
				// expecting drift - terraform going to remove unknown service account
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.CheckResourceServiceAccountsLen(theResource, 1),
				),
			},
			{
				// added 2 new service accounts to the resource though terraform
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "3"),
					acctests.CheckResourceServiceAccountsLen(theResource, 3),
				),
			},
			{
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs),
				Check: acctests.ComposeTestCheckFunc(
					// delete one service account from the resource though API
					acctests.DeleteResourceServiceAccount(theResource, serviceAccountResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceServiceAccountsLen(theResource, 2),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "3"),
				),
				// expecting drift - terraform going to restore deleted service account
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckResourceServiceAccountsLen(theResource, 3),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "3"),
				),
			},
			{
				// remove 2 service accounts from the resource though terraform
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckResourceServiceAccountsLen(theResource, 1),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
				),
			},
		},
	})
}

func createResource13(networkName, resourceName string, serviceAccounts, serviceAccountIDs []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test13" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test13" {
	  name = "%s"
	  address = "acc-test.com.13"
	  remote_network_id = twingate_remote_network.test13.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  is_authoritative = true
	  access {
	    service_account_ids = [%s]
	  }

	}
	`, networkName, strings.Join(serviceAccounts, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(serviceAccountIDs, ", "))
}

func TestAccTwingateResourceAccessWitEmptyGroups(t *testing.T) {
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource18(remoteNetworkName, resourceName),
				ExpectError: regexp.MustCompile("Error: Not enough list items"),
			},
		},
	})
}

func createResource18(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test18" {
	  name = "%s"
	}

	resource "twingate_resource" "test18" {
	  name = "%s"
	  address = "acc-test.com.18"
	  remote_network_id = twingate_remote_network.test18.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  access {
	    group_ids = []
	  }

	}
	`, networkName, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceAccessWitEmptyServiceAccounts(t *testing.T) {
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource19(remoteNetworkName, resourceName),
				ExpectError: regexp.MustCompile("Error: Not enough list items"),
			},
		},
	})
}

func createResource19(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test19" {
	  name = "%s"
	}

	resource "twingate_resource" "test19" {
	  name = "%s"
	  address = "acc-test.com.19"
	  remote_network_id = twingate_remote_network.test19.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  access {
	    service_account_ids = []
	  }

	}
	`, networkName, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceAccessWitEmptyBlock(t *testing.T) {
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource20(remoteNetworkName, resourceName),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
		},
	})
}

func createResource20(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test20" {
	  name = "%s"
	}

	resource "twingate_resource" "test20" {
	  name = "%s"
	  address = "acc-test.com.20"
	  remote_network_id = twingate_remote_network.test20.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  access {
	  }

	}
	`, networkName, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceWithAccessBlockAndDeprecatedGroups(t *testing.T) {
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	groups, groupsID := genNewGroups("g21", 2)

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource21(remoteNetworkName, resourceName, groups, groupsID),
				ExpectError: regexp.MustCompile("Error: Conflicting configuration arguments"),
			},
		},
	})
}

func createResource21(networkName, resourceName string, groups, groupsID []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test21" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test21" {
	  name = "%s"
	  address = "acc-test.com.21"
	  remote_network_id = twingate_remote_network.test21.id
	
	  group_ids = [%s]

	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  access {
	    group_ids = [%s]
	  }

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, strings.Join(groupsID, ", "), model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func TestAccTwingateResourceAccessGroupsNotAuthoritative(t *testing.T) {
	const theResource = "twingate_resource.test22"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g22", 3)

	groupResource := getResourceNameFromID(groupsID[2])

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource22(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added new group to the resource though API
					acctests.AddResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 2),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource22(remoteNetworkName, resourceName, groups, groupsID[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 2),
				),
			},
			{
				// added new group to the resource though terraform
				Config: createResource22(remoteNetworkName, resourceName, groups, groupsID[:2]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "2"),
					acctests.CheckResourceGroupsLen(theResource, 3),
				),
			},
			{
				// remove one group from the resource though terraform
				Config: createResource22(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 2),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource22(remoteNetworkName, resourceName, groups, groupsID[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 2),
					// remove one group from the resource though API
					acctests.DeleteResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource22(remoteNetworkName, resourceName, groups, groupsID[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
		},
	})
}

func createResource22(networkName, resourceName string, groups, groupsID []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test22" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test22" {
	  name = "%s"
	  address = "acc-test.com.22"
	  remote_network_id = twingate_remote_network.test22.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  is_authoritative = false
	  access {
	    group_ids = [%s]
	  }

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func TestAccTwingateResourceAccessGroupsAuthoritative(t *testing.T) {
	const theResource = "twingate_resource.test23"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g23", 3)

	groupResource := getResourceNameFromID(groupsID[2])

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added new group to the resource though API
					acctests.AddResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 2),
				),
				// expecting drift - terraform going to remove unknown group
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
			{
				// added 2 new groups to the resource though terraform
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "3"),
					acctests.CheckResourceGroupsLen(theResource, 3),
				),
			},
			{
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					// delete one group from the resource though API
					acctests.DeleteResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 2),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "3"),
				),
				// expecting drift - terraform going to restore deleted group
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckResourceGroupsLen(theResource, 3),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "3"),
				),
			},
			{
				// remove 2 groups from the resource though terraform
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
		},
	})
}

func createResource23(networkName, resourceName string, groups, groupsID []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test23" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test23" {
	  name = "%s"
	  address = "acc-test.com.23"
	  remote_network_id = twingate_remote_network.test23.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  is_authoritative = true
	  access {
	    group_ids = [%s]
	  }

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func TestGetResourceNameFromID(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "twingate_resource.test.id",
			expected: "twingate_resource.test",
		},
		{
			input:    "twingate_resource.test",
			expected: "",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, c := range cases {
		actual := getResourceNameFromID(c.input)
		assert.Equal(t, c.expected, actual)
	}
}

func TestAccTwingateCreateResourceWithFlagIsVisible(t *testing.T) {
	const terraformResourceName = "test24"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createSimpleResource(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, isVisibleAttr),
				),
			},
			{
				// expecting no changes
				PlanOnly: true,
				Config:   createResourceWithFlagIsVisible(terraformResourceName, remoteNetworkName, resourceName, true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, isVisibleAttr, "true"),
				),
			},
			{
				ExpectNonEmptyPlan: true,
				Config:             createResourceWithFlagIsVisible(terraformResourceName, remoteNetworkName, resourceName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, isVisibleAttr, "false"),
				),
			},
			{
				// expecting no changes
				PlanOnly: true,
				Config:   createSimpleResource(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckNoResourceAttr(theResource, isVisibleAttr),
				),
			},
		},
	})
}

func createSimpleResource(terraformResourceName, networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_resource" "%s" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%s.id
	}
	`, terraformResourceName, networkName, terraformResourceName, resourceName, terraformResourceName)
}

func createResourceWithFlagIsVisible(terraformResourceName, networkName, resourceName string, isVisible bool) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_resource" "%s" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%s.id
	  is_visible = %v
	}
	`, terraformResourceName, networkName, terraformResourceName, resourceName, terraformResourceName, isVisible)
}

func TestAccTwingateCreateResourceWithFlagIsBrowserShortcutEnabled(t *testing.T) {
	const terraformResourceName = "test25"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createSimpleResource(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, isBrowserShortcutEnabledAttr),
				),
			},
			{
				// expecting no changes
				PlanOnly: true,
				Config:   createResourceWithFlagIsBrowserShortcutEnabled(terraformResourceName, remoteNetworkName, resourceName, true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, isBrowserShortcutEnabledAttr, "true"),
				),
			},
			{
				ExpectNonEmptyPlan: true,
				Config:             createResourceWithFlagIsBrowserShortcutEnabled(terraformResourceName, remoteNetworkName, resourceName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, isBrowserShortcutEnabledAttr, "false"),
				),
			},
			{
				// expecting no changes
				PlanOnly: true,
				Config:   createSimpleResource(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckNoResourceAttr(theResource, isBrowserShortcutEnabledAttr),
				),
			},
		},
	})
}

func createResourceWithFlagIsBrowserShortcutEnabled(terraformResourceName, networkName, resourceName string, isBrowserShortcutEnabled bool) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_resource" "%s" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%s.id
	  is_browser_shortcut_enabled = %v
	}
	`, terraformResourceName, networkName, terraformResourceName, resourceName, terraformResourceName, isBrowserShortcutEnabled)
}

func TestAccTwingateResourceGroupsAuthoritativeByDefault(t *testing.T) {
	const theResource = "twingate_resource.test26"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g26", 3)

	groupResource := getResourceNameFromID(groupsID[2])

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added new group to the resource though API
					acctests.AddResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 2),
				),
				// expecting drift - terraform going to remove unknown group
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
			{
				// added 2 new groups to the resource though terraform
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "3"),
					acctests.CheckResourceGroupsLen(theResource, 3),
				),
			},
			{
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					// delete one group from the resource though API
					acctests.DeleteResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 2),
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "3"),
				),
				// expecting drift - terraform going to restore deleted group
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckResourceGroupsLen(theResource, 3),
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "3"),
				),
			},
			{
				// remove 2 groups from the resource though terraform
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
		},
	})
}

func createResource26(networkName, resourceName string, groups, groupsID []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test26" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test26" {
	  name = "%s"
	  address = "acc-test.com.26"
	  remote_network_id = twingate_remote_network.test26.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  group_ids = [%s]

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func TestAccTwingateResourceGroupsNotAuthoritative(t *testing.T) {
	const theResource = "twingate_resource.test27"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g27", 3)

	groupResource := getResourceNameFromID(groupsID[2])

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource27(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added new group to the resource though API
					acctests.AddResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 2),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource27(remoteNetworkName, resourceName, groups, groupsID[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 2),
				),
			},
			{
				// added new group to the resource though terraform
				Config: createResource27(remoteNetworkName, resourceName, groups, groupsID[:2]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "2"),
					acctests.CheckResourceGroupsLen(theResource, 3),
				),
			},
			{
				// remove one group from the resource though terraform
				Config: createResource27(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 2),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource27(remoteNetworkName, resourceName, groups, groupsID[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 2),
					// remove one group from the resource though API
					acctests.DeleteResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
			{
				// expecting no drift - empty plan
				Config:   createResource27(remoteNetworkName, resourceName, groups, groupsID[:1]),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
		},
	})
}

func createResource27(networkName, resourceName string, groups, groupsID []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test27" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test27" {
	  name = "%s"
	  address = "acc-test.com.27"
	  remote_network_id = twingate_remote_network.test27.id
	  
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp {
	      policy = "%s"
	    }
	  }

	  is_authoritative = false
	  group_ids = [%s]

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}
