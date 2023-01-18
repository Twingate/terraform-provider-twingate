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
)

const (
	groupIdsLen  = "group_ids.#"
	tcpPolicy    = "protocols.0.tcp.0.policy"
	udpPolicy    = "protocols.0.udp.0.policy"
	firstTCPPort = "protocols.0.tcp.0.ports.0"
	firstUDPPort = "protocols.0.udp.0.ports.0"
	tcpPortsLen  = "protocols.0.tcp.0.ports.#"
	udpPortsLen  = "protocols.0.udp.0.ports.#"

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

func TestAccTwingateResourceLoadsAllGroups(t *testing.T) {
	const theResource = "twingate_resource.test13"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g13", 111)

	sdk.Test(t, sdk.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource13(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "111"),
				),
			},
			{
				Config: createResource13(remoteNetworkName, resourceName, groups[:75], groupsID[:75]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, groupIdsLen, "75"),
				),
			},
		},
	})
}

func createResource13(networkName, resourceName string, groups, groupsID []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test13" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test13" {
	  name = "%s"
	  address = "acc-test.com.13"
	  remote_network_id = twingate_remote_network.test13.id
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
	}
	`, networkName, strings.Join(groups, "\n"), resourceName, strings.Join(groupsID, ", "), model.PolicyRestricted, model.PolicyAllowAll)
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

func newTerraformGroup(resourceName, groupName string) string {
	return fmt.Sprintf(`
    resource "twingate_group" "%s" {
      name = "%s"
    }
	`, resourceName, groupName)
}

func TestAccTwingateCreateResourceWithFlagIsVisible(t *testing.T) {
	const terraformResourceName = "test14"
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
				Config: createResourceWithFlagIsVisible(terraformResourceName, remoteNetworkName, resourceName, false),
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
	const terraformResourceName = "test15"
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
				Config: createResourceWithFlagIsBrowserShortcutEnabled(terraformResourceName, remoteNetworkName, resourceName, false),
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
