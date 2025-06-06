package resource

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
)

var (
	tcpPolicy                  = attr.PathAttr(attr.Protocols, attr.TCP, attr.Policy)
	udpPolicy                  = attr.PathAttr(attr.Protocols, attr.UDP, attr.Policy)
	firstTCPPort               = attr.FirstAttr(attr.Protocols, attr.TCP, attr.Ports)
	firstUDPPort               = attr.FirstAttr(attr.Protocols, attr.UDP, attr.Ports)
	tcpPortsLen                = attr.LenAttr(attr.Protocols, attr.TCP, attr.Ports)
	udpPortsLen                = attr.LenAttr(attr.Protocols, attr.UDP, attr.Ports)
	accessGroupIdsLen          = attr.Len(attr.AccessGroup)
	accessServiceAccountIdsLen = attr.Len(attr.AccessService)
)

func TestAccTwingateResourceCreate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test1"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceOnlyWithNetwork(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, accessGroupIdsLen),
					sdk.TestCheckResourceAttr(acctests.TerraformRemoteNetwork(terraformResourceName), attr.Name, remoteNetworkName),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckResourceAttr(theResource, attr.Address, "acc-test.com"),
				),
			},
		},
	})
}

func TestAccTwingateResourceUpdateProtocols(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test1u"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceOnlyWithNetwork(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithSimpleProtocols(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
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

func createResourceWithSimpleProtocols(terraformResourceName, networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_resource" "%s" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%s.id

	  protocols = {
        allow_icmp = true
        tcp = {
            policy = "DENY_ALL"
        }
        udp = {
            policy = "DENY_ALL"
        }
      }
	}
	`, terraformResourceName, networkName, terraformResourceName, resourceName, terraformResourceName)
}

func TestAccTwingateResourceCreateWithProtocolsAndGroups(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test2"
	remoteNetworkName := test.RandomName()
	groupName1 := test.RandomGroupName()
	groupName2 := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithProtocolsAndGroups(remoteNetworkName, groupName1, groupName2, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Address, "new-acc-test.com"),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "2"),
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

      protocols = {
		allow_icmp = true
        tcp = {
			policy = "%s"
            ports = ["80", "82-83"]
        }
		udp = {
 			policy = "%s"
		}
      }

      dynamic "access_group" {
		for_each = [twingate_group.g21.id, twingate_group.g22.id]
		content {
			group_id = access_group.value
		}
      }
	}
	`, networkName, groupName1, groupName2, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceFullCreationFlow(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test3"
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: resourceFullCreationFlow(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr("twingate_remote_network.test3", attr.Name, remoteNetworkName),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestMatchResourceAttr("twingate_connector_tokens.test31", attr.AccessToken, regexp.MustCompile(".+")),
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

      protocols = {
        allow_icmp = true
        tcp = {
            policy = "%s"
            ports = ["3306"]
        }
        udp = {
            policy = "%s"
        }
      }

      access_group {
        group_id = twingate_group.test3.id
      }
    }
    `, networkName, groupName, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceWithInvalidGroupId(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	networkName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      createResourceWithInvalidGroupId(networkName, resourceName),
				ExpectError: regexp.MustCompile("Unable to parse global ID"),
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

	  dynamic "access_group" {
		for_each = ["foo", "bar"]
		content {
			group_id = access_group.value
		}
      }
	  remote_network_id = twingate_remote_network.test4.id
	}
	`, networkName, resourceName)
}

func TestAccTwingateResourceWithTcpDenyAllPolicy(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test5"
	resourceName := test.RandomResourceName()
	networkName := test.RandomResourceName()
	groupName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithTcpDenyAllPolicy(networkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyDenyAll),
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
      access_group {
        group_id = twingate_group.g5.id
      }
      protocols = {
        allow_icmp = true
        tcp = {
          policy = "%s"
        }
        udp = {
          policy = "%s"
        }
      }
    }
    `, networkName, groupName, resourceName, model.PolicyDenyAll, model.PolicyAllowAll)
}

func TestAccTwingateResourceWithUdpDenyAllPolicy(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test6"
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, udpPolicy, model.PolicyDenyAll),
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
      access_group {
        group_id = twingate_group.g6.id
      }
      protocols = {
        allow_icmp = true
        tcp = {
          policy = "%s"
        }
        udp = {
          policy = "%s"
        }
      }
    }
	`, networkName, groupName, resourceName, model.PolicyAllowAll, model.PolicyDenyAll)
}

func TestAccTwingateResourceWithDenyAllPolicyAndEmptyPortsList(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test7"
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithDenyAllPolicyAndEmptyPortsList(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyDenyAll),
					sdk.TestCheckNoResourceAttr(theResource, tcpPortsLen),
					sdk.TestCheckResourceAttr(theResource, udpPolicy, model.PolicyDenyAll),
					sdk.TestCheckNoResourceAttr(theResource, udpPortsLen),
				),
			},
			{
				Config:   createResourceWithDenyAllPolicyAndEmptyPortsList(remoteNetworkName, groupName, resourceName),
				PlanOnly: true,
			},
		},
	})
}

func createResourceWithDenyAllPolicyAndEmptyPortsList(networkName, groupName, resourceName string) string {
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
	  access_group {
	    group_id = twingate_group.test7.id
	  }
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = []
	    }
	    udp = {
	      policy = "%s"
	    }
	  }
	}
	`, networkName, groupName, resourceName, model.PolicyDenyAll, model.PolicyDenyAll)
}

func TestAccTwingateResourceWithInvalidPortRange(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	expectedError := regexp.MustCompile("failed to parse protocols port range")

	genConfig := func(portRange string) string {
		return createResourceWithRestrictedPolicyAndPortRange(remoteNetworkName, resourceName, portRange)
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
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
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = [%s]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }
	}
	`, networkName, resourceName, model.PolicyRestricted, portRange, model.PolicyAllowAll)
}

func TestAccTwingateResourcePortReorderingCreatesNoChanges(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test9"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"80", "82-83"`),
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
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"70", "82-83"`),
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
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = [%s]
	    }
	    udp = {
	      policy = "%s"
	      ports = [%s]
	    }
	  }
	}
	`, networkName, resourceName, model.PolicyRestricted, portRange, model.PolicyRestricted, portRange)
}

func TestAccTwingateResourcePortsRepresentationChanged(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test9"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"82", "83", "80"`),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "3"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePortsNotChanged(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test9"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"82", "83", "80"`),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "3"),
				),
			},
			{
				PlanOnly: true,
				Config:   createResourceWithPortRange(remoteNetworkName, resourceName, `"80", "82-83"`),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "2"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePortReorderingNoChanges(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test9"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"82", "83", "80"`),
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
				Config:   createResourceWithPortRange(remoteNetworkName, resourceName, `"82-83", "80"`),
				PlanOnly: true,
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, udpPortsLen, "2"),
				),
			},
			// new changes applied
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"70", "82-83"`),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, firstTCPPort, "70"),
					sdk.TestCheckResourceAttr(theResource, firstUDPPort, "70"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePortsReshape(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test9"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"80", "81", "90"`),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "3"),
				),
			},
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"80", "81", "91"`),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "3"),
				),
			},
		},
	})
}

func TestAccTwingateResourceSetActiveStateOnUpdate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test10"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceOnlyWithNetwork(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeactivateTwingateResource(theResource),
					acctests.WaitTestFunc(),
					acctests.CheckTwingateResourceActiveState(theResource, false),
				),
				// provider noticed drift and tried to change it to true
				ExpectNonEmptyPlan: true,
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
	t.Parallel()

	const terraformResourceName = "test10"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
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
	t.Parallel()

	const theResource = "twingate_resource.test12"
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	groupName2 := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
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
					attr.Address:                        "acc-test.com.12",
					attr.Alias:                          "test.alias",
					tcpPolicy:                           model.PolicyRestricted,
					tcpPortsLen:                         "2",
					firstTCPPort:                        "80",
					udpPolicy:                           model.PolicyAllowAll,
					udpPortsLen:                         "0",
					accessGroupIdsLen:                   "2",
					attr.ApprovalMode:                   "MANUAL",
					attr.UsageBasedAutolockDurationDays: "10",
					attr.Path(attr.AccessGroup, attr.UsageBasedAutolockDurationDays): "13",
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
	  approval_mode = "MANUAL"
	  usage_based_autolock_duration_days = 10
	  alias = "test.alias"
	  
      dynamic "access_group" {
		for_each = [twingate_group.g121.id, twingate_group.g122.id]
		content {
			group_id = access_group.value
			usage_based_autolock_duration_days = 13
		}
      }
      
      protocols = {
		allow_icmp = true
        tcp = {
			policy = "%s"
            ports = ["80", "82-83"]
        }
		udp = {
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

func TestAccTwingateResourceAddAccessServiceAccounts(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test15"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	serviceAccountName := test.RandomName("s15")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource15(remoteNetworkName, resourceName, createServiceAccount(resourceName, serviceAccountName)),
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  dynamic "access_service" {
		for_each = [%s]
		content {
			service_account_id = access_service.value
		}
      }

	}
	`, networkName, terraformServiceAccount, resourceName, model.PolicyRestricted, model.PolicyAllowAll, acctests.TerraformServiceAccount(resourceName)+".id")
}

func TestAccTwingateResourceAddAccessGroupsAndServiceAccounts(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test16"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	serviceAccountName := test.RandomName("s16")
	groups, groupsID := genNewGroups("g16", 1)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource16(remoteNetworkName, resourceName, groups, groupsID, createServiceAccount(resourceName, serviceAccountName)),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
				),
			},
			{
				Config: createResource16WithoutServiceAccounts(remoteNetworkName, resourceName, groups, groupsID, createServiceAccount(resourceName, serviceAccountName)),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "0"),
				),
			},
			{
				Config: createResource16WithoutGroups(remoteNetworkName, resourceName, groups, groupsID, createServiceAccount(resourceName, serviceAccountName)),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "0"),
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  dynamic "access_group" {
		for_each = [%s]
		content {
			group_id = access_group.value
		}
      }
      
	  dynamic "access_service" {
		for_each = [%s]
		content {
			service_account_id = access_service.value
		}
      }

	}
	`, networkName, strings.Join(groups, "\n"), terraformServiceAccount, resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "), acctests.TerraformServiceAccount(resourceName)+".id")
}

func createResource16WithoutServiceAccounts(networkName, resourceName string, groups, groupsID []string, terraformServiceAccount string) string {
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  dynamic "access_group" {
		for_each = [%s]
		content {
			group_id = access_group.value
		}
      }

	}
	`, networkName, strings.Join(groups, "\n"), terraformServiceAccount, resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func createResource16WithoutGroups(networkName, resourceName string, groups, groupsID []string, terraformServiceAccount string) string {
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  dynamic "access_service" {
		for_each = [%s]
		content {
			service_account_id = access_service.value
		}
      }

	}
	`, networkName, strings.Join(groups, "\n"), terraformServiceAccount, resourceName, model.PolicyRestricted, model.PolicyAllowAll, acctests.TerraformServiceAccount(resourceName)+".id")
}

func TestAccTwingateResourceAccessServiceAccountsNotAuthoritative(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test17"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	serviceAccounts, serviceAccountIDs := genNewServiceAccounts("s17", 3)

	serviceAccountResource := getResourceNameFromID(serviceAccountIDs[2])

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource17(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added a new service account to the resource using API
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
				// added a new service account to the resource using terraform
				Config: createResource17(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:2]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "2"),
					acctests.CheckResourceServiceAccountsLen(theResource, 3),
				),
			},
			{
				// remove one service account from the resource using terraform
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
					// delete service account from the resource using API
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  is_authoritative = false
	  dynamic "access_service" {
		for_each = [%s]
		content {
			service_account_id = access_service.value
		}
      }

	}
	`, networkName, strings.Join(serviceAccounts, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(serviceAccountIDs, ", "))
}

func TestAccTwingateResourceAccessServiceAccountsAuthoritative(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test13"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	serviceAccounts, serviceAccountIDs := genNewServiceAccounts("s13", 3)

	serviceAccountResource := getResourceNameFromID(serviceAccountIDs[2])

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added new service account to the resource using API
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
				// added 2 new service accounts to the resource using terraform
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "3"),
					acctests.CheckResourceServiceAccountsLen(theResource, 3),
				),
			},
			{
				Config: createResource13(remoteNetworkName, resourceName, serviceAccounts, serviceAccountIDs),
				Check: acctests.ComposeTestCheckFunc(
					// delete one service account from the resource using API
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
				// remove 2 service accounts from the resource using terraform
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  is_authoritative = true
	  
	  dynamic "access_service" {
		for_each = [%s]
		content {
			service_account_id = access_service.value
		}
      }

	}
	`, networkName, strings.Join(serviceAccounts, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(serviceAccountIDs, ", "))
}

func TestAccTwingateResourceAccessWithEmptyGroups(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource18(remoteNetworkName, resourceName),
				ExpectError: regexp.MustCompile("Error: Invalid Attribute Value"),
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  access_group {
	    group_id = ""
	  }

	}
	`, networkName, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceAccessWithEmptyServiceAccounts(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource19(remoteNetworkName, resourceName),
				ExpectError: regexp.MustCompile("Error: Invalid Attribute Value"),
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  access_service {
	    service_account_id = ""
	  }

	}
	`, networkName, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceAccessWithEmptyBlock(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource20(remoteNetworkName, resourceName),
				ExpectError: regexp.MustCompile("invalid attribute combination"),
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  access_group {
	  }

	}
	`, networkName, resourceName, model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceAccessGroupsNotAuthoritative(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test22"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g22", 3)

	groupResource := getResourceNameFromID(groupsID[2])

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource22(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added a new group to the resource using API
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
				// added a new group to the resource using terraform
				Config: createResource22(remoteNetworkName, resourceName, groups, groupsID[:2]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "2"),
					acctests.CheckResourceGroupsLen(theResource, 3),
				),
			},
			{
				// remove one group from the resource using terraform
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
					// remove one group from the resource using API
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  is_authoritative = false
	  dynamic "access_group" {
		for_each = [%s]
		content {
			group_id = access_group.value
		}
      }

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func TestAccTwingateResourceAccessGroupsAuthoritative(t *testing.T) {
	t.Parallel()

	const theResource = "twingate_resource.test23"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g23", 3)

	groupResource := getResourceNameFromID(groupsID[2])

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added a new group to the resource using API
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
				// added 2 new groups to the resource using terraform
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "3"),
					acctests.CheckResourceGroupsLen(theResource, 3),
				),
			},
			{
				Config: createResource23(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					// delete one group from the resource using API
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
				// remove 2 groups from the resource using terraform
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  is_authoritative = true
	  dynamic "access_group" {
		for_each = [%s]
		content {
			group_id = access_group.value
		}
      }

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func TestGetResourceNameFromID(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	const terraformResourceName = "test24"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createSimpleResource(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.IsVisible, "true"),
				),
			},
			{
				Config: createResourceWithFlagIsVisible(terraformResourceName, remoteNetworkName, resourceName, true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsVisible, "true"),
				),
			},
			{
				Config: createResourceWithFlagIsVisible(terraformResourceName, remoteNetworkName, resourceName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsVisible, "false"),
				),
			},
			{
				// expecting no changes - no drift after re-applying changes
				PlanOnly: true,
				Config:   createResourceWithFlagIsVisible(terraformResourceName, remoteNetworkName, resourceName, false),
			},
			{
				Config: createSimpleResource(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsVisible, "true"),
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
	t.Parallel()

	const terraformResourceName = "test25"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createSimpleResource(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithFlagIsBrowserShortcutEnabled(terraformResourceName, remoteNetworkName, resourceName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsBrowserShortcutEnabled, "false"),
				),
			},
			{
				Config: createResourceWithFlagIsBrowserShortcutEnabled(terraformResourceName, remoteNetworkName, resourceName, true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsBrowserShortcutEnabled, "true"),
				),
			},
			{
				// expecting no changes - no drift after re-applying changes
				PlanOnly: true,
				Config:   createResourceWithFlagIsBrowserShortcutEnabled(terraformResourceName, remoteNetworkName, resourceName, true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsBrowserShortcutEnabled, "true"),
				),
			},
			{
				Config: createResourceWithFlagIsBrowserShortcutEnabled(terraformResourceName, remoteNetworkName, resourceName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsBrowserShortcutEnabled, "false"),
				),
			},
			{
				// expecting no changes - flag not set
				PlanOnly: true,
				Config:   createSimpleResource(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckNoResourceAttr(theResource, attr.IsBrowserShortcutEnabled),
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
	t.Parallel()

	const theResource = "twingate_resource.test26"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g26", 3)

	groupResource := getResourceNameFromID(groupsID[2])

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.WaitTestFunc(),
					// added a new group to the resource using API
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
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
			{
				// added 2 new groups to the resource using terraform
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "3"),
					acctests.CheckResourceGroupsLen(theResource, 3),
				),
			},
			{
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					// delete one group from the resource using API
					acctests.DeleteResourceGroup(theResource, groupResource),
					acctests.WaitTestFunc(),
					acctests.CheckResourceGroupsLen(theResource, 2),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "3"),
				),
				// expecting drift - terraform going to restore deleted group
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckResourceGroupsLen(theResource, 3),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "3"),
				),
			},
			{
				// remove 2 groups from the resource using terraform
				Config: createResource26(remoteNetworkName, resourceName, groups, groupsID[:1]),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "1"),
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
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  dynamic "access_group" {
		for_each = [%s]
		content {
			group_id = access_group.value
		}
      }

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "))
}

func TestAccTwingateResourceDoesNotSupportOldGroups(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	groups, groupsID := genNewGroups("g28", 2)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource28(remoteNetworkName, resourceName, groups, groupsID),
				ExpectError: regexp.MustCompile("Error: Unsupported argument"),
			},
		},
	})
}

func createResource28(networkName, resourceName string, groups, groupsID []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test28" {
	  name = "%s"
	}

	%s

	resource "twingate_resource" "test28" {
	  name = "%s"
	  address = "acc-test.com.28"
	  remote_network_id = twingate_remote_network.test28.id
	
	  group_ids = [%s]

	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	}
	`, networkName, strings.Join(groups, "\n"), resourceName, strings.Join(groupsID, ", "), model.PolicyRestricted, model.PolicyAllowAll)
}

func TestAccTwingateResourceCreateWithAlias(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test29"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	const aliasName = "test.com"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource29(terraformResourceName, remoteNetworkName, resourceName, aliasName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckResourceAttr(theResource, attr.Alias, aliasName),
				),
			},
			{
				// alias attr commented out, means it has nil state
				Config: createResource29WithoutAlias(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckNoResourceAttr(theResource, attr.Alias),
				),
			},
			{
				// alias attr set with emtpy string
				Config: createResource29(terraformResourceName, remoteNetworkName, resourceName, ""),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Alias, ""),
				),
			},
		},
	})
}

func TestAccTwingateResourceUpdateWithInvalidAlias(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test29_update_invalid"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource29WithoutAlias(terraformResourceName, remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckNoResourceAttr(theResource, attr.Alias),
				),
			},
			{
				Config:      createResource29(terraformResourceName, remoteNetworkName, resourceName, "test-com"),
				ExpectError: regexp.MustCompile("Alias must be a[\\n\\s]+valid DNS name"),
			},
		},
	})
}

func TestAccTwingateResourceCreateWithInvalidAlias(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test29_create_invalid"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResource29(terraformResourceName, remoteNetworkName, resourceName, "test-com"),
				ExpectError: regexp.MustCompile("Alias must be a[\\n\\s]+valid DNS name"),
			},
		},
	})
}

func createResource29(terraformResourceName, networkName, resourceName, aliasName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_resource" "%s" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%s.id
	  alias = "%s"
	}
	`, terraformResourceName, networkName, terraformResourceName, resourceName, terraformResourceName, aliasName)
}

func createResource29WithoutAlias(terraformResourceName, networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_resource" "%s" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%s.id
	  # alias = "some.value"
	}
	`, terraformResourceName, networkName, terraformResourceName, resourceName, terraformResourceName)
}

func TestAccTwingateResourceGroupsCursor(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test27"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	groups, groupsID := genNewGroups("g27", 3)
	serviceAccounts, serviceAccountIDs := genNewServiceAccounts("s27", 3)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithGroupsAndServiceAccounts(terraformResourceName, remoteNetworkName, resourceName, groups, groupsID, serviceAccounts, serviceAccountIDs),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "3"),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "3"),
				),
			},
			{
				Config: createResourceWithGroupsAndServiceAccounts(terraformResourceName, remoteNetworkName, resourceName, groups, groupsID[:2], serviceAccounts, serviceAccountIDs[:2]),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, accessGroupIdsLen, "2"),
					sdk.TestCheckResourceAttr(theResource, accessServiceAccountIdsLen, "2"),
				),
			},
		},
	})
}

func createResourceWithGroupsAndServiceAccounts(name, networkName, resourceName string, groups, groupsID, serviceAccounts, serviceAccountIDs []string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}

	%s

	%s

	resource "twingate_resource" "%s" {
	  name = "%s"
	  address = "acc-test.com.26"
	  remote_network_id = twingate_remote_network.%s.id
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%s"
	    }
	  }

	  dynamic "access_group" {
		for_each = [%s]
		content {
			group_id = access_group.value
		}
      }

      dynamic "access_service" {
		for_each = [%s]
		content {
			service_account_id = access_service.value
		}
      }

	}
	`, name, networkName, strings.Join(groups, "\n"), strings.Join(serviceAccounts, "\n"), name, resourceName, name, model.PolicyRestricted, model.PolicyAllowAll, strings.Join(groupsID, ", "), strings.Join(serviceAccountIDs, ", "))
}

func TestAccTwingateResourceCreateWithPort(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResourceWithPort(remoteNetworkName, resourceName, "0"),
				ExpectError: regexp.MustCompile("port 0 not in the range of 1-65535"),
			},
			{
				Config:      createResourceWithPort(remoteNetworkName, resourceName, "65536"),
				ExpectError: regexp.MustCompile("port 65536 not in the range"),
			},
			{
				Config:      createResourceWithPort(remoteNetworkName, resourceName, "0-10"),
				ExpectError: regexp.MustCompile("port 0 not in the range"),
			},
			{
				Config:      createResourceWithPort(remoteNetworkName, resourceName, "65535-65536"),
				ExpectError: regexp.MustCompile("port 65536 not in the[\\n\\s]+range"),
			},
		},
	})
}

func createResourceWithPort(networkName, resourceName, port string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test30" {
	  name = "%s"
	}
	resource "twingate_resource" "test30" {
	  name = "%s"
	  address = "new-acc-test.com"
	  remote_network_id = twingate_remote_network.test30.id
	  protocols = {
		allow_icmp = true
		tcp = {
			policy = "%s"
			ports = ["%s"]
		}
		udp = {
			policy = "%s"
		}
	  }
	}
	`, networkName, resourceName, model.PolicyRestricted, port, model.PolicyAllowAll)
}

func TestAccTwingateResourceUpdateWithPort(t *testing.T) {
	t.Parallel()

	theResource := acctests.TerraformResource("test30")
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPort(remoteNetworkName, resourceName, "1"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, firstTCPPort, "1"),
				),
			},
			{
				Config:      createResourceWithPort(remoteNetworkName, resourceName, "0"),
				ExpectError: regexp.MustCompile("port 0 not in the range of 1-65535"),
			},
		},
	})
}

func TestAccTwingateResourceWithPortsFailsForAllowAllAndDenyAllPolicy(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test28"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyAllowAll),
				ExpectError: regexp.MustCompile(resource.ErrPortsWithPolicyAllowAll.Error()),
			},
			{
				Config:      createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyDenyAll),
				ExpectError: regexp.MustCompile(resource.ErrPortsWithPolicyDenyAll.Error()),
			},
		},
	})
}

func createResourceWithPorts(name, networkName, resourceName, policy string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[2]s"
	}
	resource "twingate_resource" "%[1]s" {
	  name = "%[3]s"
	  address = "acc-test-%[1]s.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%[4]s"
	      ports = ["80", "82-83"]
	    }
	    udp = {
	      policy = "%[5]s"
	    }
	  }
	}
	`, name, networkName, resourceName, policy, model.PolicyAllowAll)
}

func TestAccTwingateResourceWithoutPortsOkForAllowAllAndDenyAllPolicy(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test29"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(terraformResourceName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyAllowAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyAllowAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyDenyAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyDenyAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
		},
	})
}

func createResourceWithoutPorts(name, networkName, resourceName, policy string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[2]s"
	}
	resource "twingate_resource" "%[1]s" {
	  name = "%[3]s"
	  address = "acc-test-%[1]s.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "%[4]s"
	    }
	    udp = {
	      policy = "%[5]s"
	    }
	  }
	}
	`, name, networkName, resourceName, policy, model.PolicyAllowAll)
}

func TestAccTwingateResourceWithRestrictedPolicy(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test30"
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(terraformResourceName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyRestricted),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "2"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePolicyTransitionDenyAllToRestricted(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test31"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyDenyAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyDenyAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
			{
				Config: createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyRestricted),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "2"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePolicyTransitionDenyAllToAllowAll(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test32"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyDenyAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyDenyAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyAllowAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyAllowAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePolicyTransitionRestrictedToDenyAll(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test33"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyRestricted),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "2"),
				),
			},
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyDenyAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyDenyAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePolicyTransitionRestrictedToAllowAll(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test34"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyRestricted),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "2"),
				),
			},
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyAllowAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyAllowAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePolicyTransitionRestrictedToAllowAllWithPortsShouldFail(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test35"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyRestricted),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "2"),
				),
			},
			{
				Config:      createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyAllowAll),
				ExpectError: regexp.MustCompile(resource.ErrPortsWithPolicyAllowAll.Error()),
			},
		},
	})
}

func TestAccTwingateResourcePolicyTransitionAllowAllToRestricted(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test36"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyAllowAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyAllowAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
			{
				Config: createResourceWithPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyRestricted),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyRestricted),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "2"),
				),
			},
		},
	})
}

func TestAccTwingateResourcePolicyTransitionAllowAllToDenyAll(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test37"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyAllowAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyAllowAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
			{
				Config: createResourceWithoutPorts(terraformResourceName, remoteNetworkName, resourceName, model.PolicyDenyAll),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, tcpPolicy, model.PolicyDenyAll),
					sdk.TestCheckResourceAttr(theResource, tcpPortsLen, "0"),
				),
			},
		},
	})
}

func TestAccTwingateResourceTestCaseInsensitiveAlias(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test38"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	const aliasName = "test.com"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource29(terraformResourceName, remoteNetworkName, resourceName, aliasName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckResourceAttr(theResource, attr.Alias, aliasName),
				),
			},
			{
				// expecting no changes
				PlanOnly: true,
				Config:   createResource29(terraformResourceName, remoteNetworkName, resourceName, strings.ToUpper(aliasName)),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Alias, aliasName),
				),
			},
		},
	})
}

func TestAccTwingateResourceWithBrowserOption(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test40"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	wildcardAddress := "*.acc-test.com"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutBrowserOption(terraformResourceName, remoteNetworkName, resourceName, wildcardAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, wildcardAddress, false),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config:      createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, wildcardAddress, true),
				ExpectError: regexp.MustCompile("Resources with a CIDR range or wildcard"),
			},
		},
	})
}

func TestAccTwingateResourceWithBrowserOptionFailOnUpdate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test41"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	wildcardAddress := "*.acc-test.com"
	simpleAddress := "acc-test.com"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutBrowserOption(terraformResourceName, remoteNetworkName, resourceName, simpleAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, simpleAddress, true),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config:      createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, wildcardAddress, true),
				ExpectError: regexp.MustCompile("Resources with a CIDR range or wildcard"),
			},
		},
	})
}

func TestAccTwingateResourceWithBrowserOptionRecovered(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test42"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	wildcardAddress := "*.acc-test.com"
	simpleAddress := "acc-test.com"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, simpleAddress, true),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithoutBrowserOption(terraformResourceName, remoteNetworkName, resourceName, wildcardAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func TestAccTwingateResourceWithBrowserOptionIP(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "testIPWithBrowserOption"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	ipAddress := "10.0.0.1"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutBrowserOption(terraformResourceName, remoteNetworkName, resourceName, ipAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, ipAddress, false),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, ipAddress, true),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func TestAccTwingateResourceWithBrowserOptionCIDR(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "testCIDRWithBrowserOption"
	theResource := acctests.TerraformResource(terraformResourceName)
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	cidrAddress := "10.0.0.0/24"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutBrowserOption(terraformResourceName, remoteNetworkName, resourceName, cidrAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, cidrAddress, false),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config:      createResourceWithBrowserOption(terraformResourceName, remoteNetworkName, resourceName, cidrAddress, true),
				ExpectError: regexp.MustCompile("Resources with a CIDR range or wildcard"),
			},
		},
	})
}

func createResourceWithoutBrowserOption(name, networkName, resourceName, address string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[2]s"
	}
	resource "twingate_resource" "%[1]s" {
	  name = "%[3]s"
	  address = "%[4]s"
	  remote_network_id = twingate_remote_network.%[1]s.id
	}
	`, name, networkName, resourceName, address)
}

func createResourceWithBrowserOption(name, networkName, resourceName, address string, browserOption bool) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[2]s"
	}
	resource "twingate_resource" "%[1]s" {
	  name = "%[3]s"
	  address = "%[4]s"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  is_browser_shortcut_enabled = %[5]v
	}
	`, name, networkName, resourceName, address, browserOption)
}

func createResourceWithSecurityPolicy(remoteNetwork, resource, policyID string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  security_policy_id = "%[3]s"
	}
	`, remoteNetwork, resource, policyID)
}

func createResourceWithoutSecurityPolicy(remoteNetwork, resource string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	}
	`, remoteNetwork, resource)
}

func TestAccTwingateResourceUpdateWithDefaultProtocols(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithProtocols(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				Config: createResourceWithoutProtocols(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func createResourceWithProtocols(remoteNetwork, resource string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "RESTRICTED"
	      ports = ["80-83"]
	    }
	    udp = {
	      policy = "RESTRICTED"
	      ports = ["80"]
	    }
	  }
	}
	`, remoteNetwork, resource)
}

func createResourceWithoutProtocols(remoteNetwork, resource string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	}
	`, remoteNetwork, resource)
}

func TestAccTwingateResourceUpdatePortsFromEmptyListToNull(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithEmptyArrayPorts(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				// expect no changes
				PlanOnly: true,
				Config:   createResourceWithDefaultPorts(remoteNetworkName, resourceName),
			},
		},
	})
}

func TestAccTwingateResourceUpdatePortsFromNullToEmptyList(t *testing.T) {
	t.Parallel()

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithDefaultPorts(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				// expect no changes
				PlanOnly: true,
				Config:   createResourceWithEmptyArrayPorts(remoteNetworkName, resourceName),
			},
		},
	})
}

func createResourceWithDefaultPorts(remoteNetwork, resource string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "ALLOW_ALL"
	    }
	    udp = {
	      policy = "ALLOW_ALL"
	    }
	  }
	}
	`, remoteNetwork, resource)
}

func createResourceWithEmptyArrayPorts(remoteNetwork, resource string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	    udp = {
	      policy = "ALLOW_ALL"
	      ports = []
	    }
	  }
	}
	`, remoteNetwork, resource)
}

func TestAccTwingateResourceUpdateSecurityPolicy(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()

	defaultPolicy, testPolicy := preparePolicies(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutSecurityPolicy(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, attr.SecurityPolicyID),
				),
			},
			{
				Config: createResourceWithSecurityPolicy(remoteNetworkName, resourceName, testPolicy),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.SecurityPolicyID, testPolicy),
				),
			},
			{
				Config: createResourceWithoutSecurityPolicy(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.SecurityPolicyID, defaultPolicy),
				),
			},
			{
				Config: createResourceWithSecurityPolicy(remoteNetworkName, resourceName, ""),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.SecurityPolicyID, ""),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func preparePolicies(t *testing.T) (string, string) {
	policies, err := acctests.ListSecurityPolicies()
	if err != nil {
		t.Skipf("failed to retrieve security policies: %v", err)
	}

	if len(policies) < 2 {
		t.Skip("requires at least 2 security policy for the test")
	}

	var defaultPolicy, testPolicy string
	if policies[0].Name == resource.DefaultSecurityPolicyName {
		defaultPolicy = policies[0].ID
		testPolicy = policies[1].ID
	} else {
		testPolicy = policies[0].ID
		defaultPolicy = policies[1].ID
	}

	return defaultPolicy, testPolicy
}

func TestAccTwingateResourceSetDefaultSecurityPolicyByDefault(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()

	defaultPolicy, testPolicy := preparePolicies(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithSecurityPolicy(remoteNetworkName, resourceName, testPolicy),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.SecurityPolicyID, testPolicy),
				),
			},
			{
				Config: createResourceWithoutSecurityPolicy(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.SecurityPolicyID, defaultPolicy),
					acctests.CheckResourceSecurityPolicy(theResource, defaultPolicy),
					// set new policy via API
					acctests.UpdateResourceSecurityPolicy(theResource, testPolicy),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResourceWithSecurityPolicy(remoteNetworkName, resourceName, ""),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckResourceSecurityPolicy(theResource, defaultPolicy),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResourceWithoutSecurityPolicy(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckResourceSecurityPolicy(theResource, defaultPolicy),
				),
			},
		},
	})
}

func TestAccTwingateResourceSecurityPolicy(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()

	_, testPolicy := preparePolicies(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithoutSecurityPolicy(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, attr.SecurityPolicyID),
				),
			},
			{
				Config: createResourceWithSecurityPolicy(remoteNetworkName, resourceName, testPolicy),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.SecurityPolicyID, testPolicy),
				),
			},
		},
	})
}

func TestAccTwingateResourceCreateInactive(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithIsActiveFlag(remoteNetworkName, resourceName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsActive, "false"),
					acctests.CheckTwingateResourceActiveState(theResource, false),
				),
			},
		},
	})
}

func createResourceWithIsActiveFlag(networkName, resourceName string, isActive bool) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  is_active = %[3]v
	}
	`, networkName, resourceName, isActive)
}

func TestAccTwingateResourceTestInactiveFlag(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithIsActiveFlag(remoteNetworkName, resourceName, true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsActive, "true"),
				),
			},
			{
				Config: createResourceWithIsActiveFlag(remoteNetworkName, resourceName, false),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.IsActive, "false"),
					acctests.CheckTwingateResourceActiveState(theResource, false),
				),
			},
		},
	})
}

func TestAccTwingateResourceTestPlanOnDisabledResource(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceActiveState(theResource, true),
					acctests.DeactivateTwingateResource(theResource),
					acctests.CheckTwingateResourceActiveState(theResource, false),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionUpdate),
						acctests.CheckResourceActiveState(theResource, false),
					},
				},
			},
		},
	})
}

func createResource(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	}
	`, networkName, resourceName)
}

func TestAccTwingateResourceUpdateSecurityPolicyOnGroupAccess(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	defaultPolicy, testPolicy := preparePolicies(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithSecurityPolicyOnGroupAccess(remoteNetworkName, resourceName, testPolicy, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceSecurityPolicyOnGroupAccess(theResource, testPolicy),
				),
			},
			{
				Config: createResourceWithSecurityPolicyOnGroupAccess(remoteNetworkName, resourceName, defaultPolicy, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceSecurityPolicyOnGroupAccess(theResource, defaultPolicy),
				),
			},
			{
				Config: createResourceWithoutSecurityPolicyOnGroupAccess(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceSecurityPolicyIsNullOnGroupAccess(theResource),
				),
			},
		},
	})
}

func createResourceWithSecurityPolicyOnGroupAccess(remoteNetwork, resource, policyID, groupName string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[4]s"
    }

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}

	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  access_group {
		group_id = twingate_group.g21.id
		security_policy_id = "%[3]s"
      }
	}
	`, remoteNetwork, resource, policyID, groupName)
}

func createResourceWithoutSecurityPolicyOnGroupAccess(remoteNetwork, resource, groupName string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}

	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  access_group {
		group_id = twingate_group.g21.id
      }
	}
	`, remoteNetwork, resource, groupName)
}

func createResourceWithNullSecurityPolicyOnGroupAccess(remoteNetwork, resource, groupName string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}

	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  access_group {
		group_id = twingate_group.g21.id
      }
	}
	`, remoteNetwork, resource, groupName)
}

func TestAccTwingateResourceUnsetSecurityPolicyOnGroupAccess(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	_, testPolicy := preparePolicies(t)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithSecurityPolicyOnGroupAccess(remoteNetworkName, resourceName, testPolicy, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceSecurityPolicyOnGroupAccess(theResource, testPolicy),
				),
			},
			{
				Config: createResourceWithNullSecurityPolicyOnGroupAccess(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceSecurityPolicyIsNullOnGroupAccess(theResource),
				),
			},
		},
	})
}

func TestAccTwingateResourceWithUsageBasedOnGroupAccess(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	var usageBasedDuration int64 = 2

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithUsageBasedOnGroupAccess(remoteNetworkName, resourceName, groupName, usageBasedDuration),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceUsageBasedOnGroupAccess(theResource, usageBasedDuration),
				),
			},
			{
				Config: createResourceWithNullSecurityPolicyOnGroupAccess(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceUsageBasedIsNullOnGroupAccess(theResource),
				),
			},
		},
	})
}

func createResourceWithUsageBasedOnGroupAccess(remoteNetwork, resource, groupName string, daysDuration int64) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}

	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  access_group {
		group_id = twingate_group.g21.id
		usage_based_autolock_duration_days = %[4]v
      }
	}
	`, remoteNetwork, resource, groupName, daysDuration)
}

func TestAccTwingateWithMultipleResource(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	theResource1 := acctests.TerraformResource(resourceName + "-1")
	theResource2 := acctests.TerraformResource(resourceName + "-2")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createMultipleResourcesN(remoteNetworkName, resourceName, groupName, 15),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource1),
					acctests.CheckTwingateResourceExists(theResource2),
					sdk.TestCheckResourceAttr(theResource1, accessGroupIdsLen, "2"),
					sdk.TestCheckResourceAttr(theResource2, accessGroupIdsLen, "2"),
				),
			},
		},
	})
}

func createMultipleResourcesN(remoteNetwork, resource, groupName string, n int) string {
	return fmt.Sprintf(`
	resource "twingate_group" "%[2]s-group-1" {
      name = "%[3]s-1"
    }

	resource "twingate_group" "%[2]s-group-2" {
      name = "%[3]s-2"
    }

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}

	`+genMultipleResource(n),
		remoteNetwork, resource, groupName)
}

func genMultipleResource(n int) string {
	res := make([]string, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, fmtResource(i+1))
	}

	return strings.Join(res, "\n\n")
}

func fmtResource(index int) string {
	return fmt.Sprintf(`
	resource "twingate_resource" "%%[2]s-%[1]v" {
	  name = "%%[2]s-%[1]v"
	  address = "acc-test-address-%[1]v.com"
	  remote_network_id = twingate_remote_network.%%[1]s.id

	  dynamic "access_group" {
		for_each = [twingate_group.%%[2]s-group-1.id, twingate_group.%%[2]s-group-2.id]
		content {
			group_id = access_group.value
		}
      }
	}
	`, index)
}

func TestAccTwingateWithMultipleGroups(t *testing.T) {
	t.Parallel()

	prefix := "group" + acctest.RandString(5)
	group1 := acctests.TerraformGroup(prefix + "_1")
	group2 := acctests.TerraformGroup(prefix + "_2")
	groups, _ := genNewGroups(prefix, 15)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: strings.Join(groups, "\n"),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(group1),
					acctests.CheckTwingateResourceExists(group2),
				),
			},
		},
	})
}

func TestAccTwingateCreateResourceWithTags(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	remoteNetworkName := test.RandomName()

	theResource := acctests.TerraformResource(resourceName)

	const tag1 = "owner"
	const tag2 = "application"
	tags := map[string]string{
		tag1: "example_team",
		tag2: "example_application",
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithTags(resourceName, remoteNetworkName, tags),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Tags, tag1), tags[tag1]),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Tags, tag2), tags[tag2]),
				),
			},
		},
	})
}

func createResourceWithTags(resourceName, networkName string, tags map[string]string) string {
	tagsList := make([]string, 0, len(tags))
	for k, v := range tags {
		tagsList = append(tagsList, fmt.Sprintf(`%[1]s = "%[2]s"`, k, v))
	}

	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  tags = {
  	    %[3]s
	  }
	}
	`, networkName, resourceName, strings.Join(tagsList, ",\n\t"))
}

func TestAccTwingateCreateResourceWithTagsUpdateTags(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	remoteNetworkName := test.RandomName()

	theResource := acctests.TerraformResource(resourceName)

	const tag1 = "owner"
	const tag2 = "application"
	tags1 := map[string]string{
		tag1: "example_team",
	}
	tags2 := map[string]string{
		tag2: "example_application",
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
				),
			},
			{
				Config: createResourceWithTags(resourceName, remoteNetworkName, tags1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Tags, tag1), tags1[tag1]),
					sdk.TestCheckNoResourceAttr(theResource, attr.PathAttr(attr.Tags, tag2)),
				),
			},
			{
				Config: createResourceWithTags(resourceName, remoteNetworkName, tags2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckNoResourceAttr(theResource, attr.PathAttr(attr.Tags, tag1)),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Tags, tag2), tags2[tag2]),
				),
			},
		},
	})
}

func createResourceWithNullApprovalMode(remoteNetwork, resource, groupName string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}

	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  access_group {
		group_id = twingate_group.g21.id
      }
	}
	`, remoteNetwork, resource, groupName)
}

func createResourceWithApprovalMode(remoteNetwork, resource, groupName, approvalMode string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}

	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  access_group {
		group_id = twingate_group.g21.id
      }

      approval_mode = "%[4]s"
	}
	`, remoteNetwork, resource, groupName, approvalMode)
}

func createResourceWithApprovalModeInAccessGroup(remoteNetwork, resource, groupName, approvalMode string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}

	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  access_group {
		group_id = twingate_group.g21.id
        approval_mode = "%[4]s"
      }

	}
	`, remoteNetwork, resource, groupName, approvalMode)
}

func TestAccTwingateResourceWithApprovalMode(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithNullApprovalMode(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeManual),
				),
			},
			{
				Config: createResourceWithApprovalMode(remoteNetworkName, resourceName, groupName, model.ApprovalModeManual),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeManual),
				),
			},
			{
				Config: createResourceWithApprovalMode(remoteNetworkName, resourceName, groupName, model.ApprovalModeAutomatic),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeAutomatic),
				),
			},
			{
				Config: createResourceWithNullApprovalMode(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeAutomatic),
				),
			},
		},
	})
}

func TestAccTwingateResourceWithApprovalModeInAccessGroup(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithNullApprovalMode(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeManual),
				),
			},
			{
				Config: createResourceWithApprovalModeInAccessGroup(remoteNetworkName, resourceName, groupName, model.ApprovalModeManual),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Path(attr.AccessGroup, attr.ApprovalMode), model.ApprovalModeManual),
				),
			},
			{
				Config: createResourceWithApprovalMode(remoteNetworkName, resourceName, groupName, model.ApprovalModeAutomatic),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeAutomatic),
					sdk.TestCheckNoResourceAttr(theResource, attr.Path(attr.AccessGroup, attr.ApprovalMode)),
				),
			},
			{
				Config: createResourceWithNullApprovalMode(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeAutomatic),
					sdk.TestCheckNoResourceAttr(theResource, attr.Path(attr.AccessGroup, attr.ApprovalMode)),
				),
			},
		},
	})
}

func TestAccTwingateResourceWithAutomaticApprovalModeInAccessGroup(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	theResource := acctests.TerraformResource(resourceName)
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithNullApprovalMode(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeManual),
				),
			},
			{
				Config: createResourceWithApprovalModeInAccessGroup(remoteNetworkName, resourceName, groupName, model.ApprovalModeAutomatic),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Path(attr.AccessGroup, attr.ApprovalMode), model.ApprovalModeAutomatic),
					sdk.TestCheckResourceAttr(theResource, attr.ApprovalMode, model.ApprovalModeManual),
				),
			},
		},
	})
}

func TestAccTwingateCreateResourceWithUsageBasedAutolockDurationDays(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	remoteNetworkName := test.RandomName()

	theResource := acctests.TerraformResource(resourceName)
	autolockDays := rand.Intn(30) + 1

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResource(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, attr.UsageBasedAutolockDurationDays),
				),
			},
			{
				Config: createResourceWithUsageBasedAutolockDurationDays(remoteNetworkName, resourceName, autolockDays),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.UsageBasedAutolockDurationDays, fmt.Sprintf("%v", autolockDays)),
				),
			},
			{
				Config: createResource(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, attr.UsageBasedAutolockDurationDays),
				),
			},
		},
	})
}

func createResourceWithUsageBasedAutolockDurationDays(networkName, resourceName string, autolockDays int) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  usage_based_autolock_duration_days = %[3]d
	}
	`, networkName, resourceName, autolockDays)
}

func TestAccTwingateCreateResourceWithDefaultUsageBasedAutolockDurationDays(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	theResource := acctests.TerraformResource(resourceName)
	autolockDays := rand.Int63n(30) + 1
	autolockDays1 := autolockDays + 1

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithDefaultAutolock(remoteNetworkName, resourceName, groupName, autolockDays),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.UsageBasedAutolockDurationDays, fmt.Sprintf("%v", autolockDays)),
					acctests.CheckTwingateResourceUsageBasedDuration(theResource, autolockDays),
					acctests.CheckTwingateResourceUsageBasedIsNullOnGroupAccess(theResource),
				),
			},
			{
				Config: createResourceWithDefaultAutolock(remoteNetworkName, resourceName, groupName, autolockDays1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.UsageBasedAutolockDurationDays, fmt.Sprintf("%v", autolockDays1)),
					acctests.CheckTwingateResourceUsageBasedDuration(theResource, autolockDays1),
					acctests.CheckTwingateResourceUsageBasedIsNullOnGroupAccess(theResource),
				),
			},
			{
				Config: createResourceWithoutDefaultAutolock(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, attr.UsageBasedAutolockDurationDays),
					acctests.CheckTwingateResourceUsageBasedIsNullOnResource(theResource),
					acctests.CheckTwingateResourceUsageBasedIsNullOnGroupAccess(theResource),
				),
			},
		},
	})
}

func createResourceWithDefaultAutolock(remoteNetwork, resource, groupName string, autolockDays int64) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  usage_based_autolock_duration_days = %[4]d
	  
	  access_group {
		group_id = twingate_group.g21.id
      }
	}
	`, remoteNetwork, resource, groupName, autolockDays)
}

func createResourceWithoutDefaultAutolock(remoteNetwork, resource, groupName string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  
	  access_group {
		group_id = twingate_group.g21.id
      }
	}
	`, remoteNetwork, resource, groupName)
}

func TestAccTwingateCreateResourceWithDefaultUsageBasedAutolockDurationDaysAndGroupAutolock(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()

	theResource := acctests.TerraformResource(resourceName)
	autolockDays1 := rand.Int63n(30) + 1
	autolockDays2 := autolockDays1 + 2

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithDefaultAutolockAndGroupAutolock(remoteNetworkName, resourceName, groupName, autolockDays1, autolockDays2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.UsageBasedAutolockDurationDays, fmt.Sprintf("%v", autolockDays1)),
					acctests.CheckTwingateResourceUsageBasedOnGroupAccess(theResource, autolockDays2),
				),
			},
			{
				Config: createResourceWithDefaultAutolock(remoteNetworkName, resourceName, groupName, autolockDays1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.UsageBasedAutolockDurationDays, fmt.Sprintf("%v", autolockDays1)),
					acctests.CheckTwingateResourceUsageBasedDuration(theResource, autolockDays1),
					acctests.CheckTwingateResourceUsageBasedIsNullOnGroupAccess(theResource),
				),
			},
			{
				Config: createResourceWithoutDefaultAutolock(remoteNetworkName, resourceName, groupName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceUsageBasedIsNullOnGroupAccess(theResource),
				),
			},
		},
	})
}

func createResourceWithDefaultAutolockAndGroupAutolock(remoteNetwork, resource, groupName string, autolockDays1, autolockDays2 int64) string {
	return fmt.Sprintf(`
	resource "twingate_group" "g21" {
      name = "%[3]s"
    }
	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test-address.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  usage_based_autolock_duration_days = %[4]d
	  
	  access_group {
		group_id = twingate_group.g21.id
		usage_based_autolock_duration_days = %[5]d
      }
	}
	`, remoteNetwork, resource, groupName, autolockDays1, autolockDays2)
}

func TestAccTwingateCreateResourceWithDefaultTags(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomResourceName()
	remoteNetworkName := test.RandomName()

	theResource := acctests.TerraformResource(resourceName)

	const tagOwner = "owner"
	const tagApp = "application"
	const tagEnv = "env"
	tags := map[string]string{
		tagOwner: "example_team",
		tagApp:   "custom_application",
	}
	defaultTags := map[string]string{
		tagEnv: "stage",
		tagApp: "example_application",
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: createResourceWithTags(resourceName, remoteNetworkName, tags),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceTags(theResource, tagOwner, "example_team"),
					acctests.CheckTwingateResourceTags(theResource, tagApp, "custom_application"),
				),
			},
			{
				Config: createResourceWithDefaultTags(resourceName, remoteNetworkName, tags, defaultTags),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceTags(theResource, tagOwner, "example_team"),
					acctests.CheckTwingateResourceTags(theResource, tagApp, "custom_application"),
					acctests.CheckTwingateResourceTags(theResource, tagEnv, "stage"),
				),
			},
			{
				Config: createResourceWithDefaultTags(resourceName, remoteNetworkName, tags, map[string]string{tagEnv: "prod"}),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceTags(theResource, tagOwner, "example_team"),
					acctests.CheckTwingateResourceTags(theResource, tagApp, "custom_application"),
					acctests.CheckTwingateResourceTags(theResource, tagEnv, "prod"),
				),
			},
			{
				Config: createResourceWithDefaultTags(resourceName, remoteNetworkName, map[string]string{
					tagOwner: "new_team",
					tagApp:   "custom_application",
				}, map[string]string{tagEnv: "prod"}),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.CheckTwingateResourceTags(theResource, tagOwner, "new_team"),
					acctests.CheckTwingateResourceTags(theResource, tagApp, "custom_application"),
					acctests.CheckTwingateResourceTags(theResource, tagEnv, "prod"),
				),
			},
		},
	})
}

func createResourceWithDefaultTags(resourceName, networkName string, tags, defaultTags map[string]string) string {
	tagsList := make([]string, 0, len(tags))
	for k, v := range tags {
		tagsList = append(tagsList, fmt.Sprintf(`%[1]s = "%[2]s"`, k, v))
	}

	defaultTagsList := make([]string, 0, len(defaultTags))
	for k, v := range defaultTags {
		defaultTagsList = append(defaultTagsList, fmt.Sprintf(`%[1]s = "%[2]s"`, k, v))
	}

	return fmt.Sprintf(`
	provider "twingate" {
	  default_tags = {
	    tags = {
		  %[4]s
	    }
	  }
	}

	resource "twingate_remote_network" "%[1]s" {
	  name = "%[1]s"
	}
	resource "twingate_resource" "%[2]s" {
	  name = "%[2]s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.%[1]s.id
	  tags = {
  	    %[3]s
	  }
	}
	`, networkName, resourceName, strings.Join(tagsList, ",\n\t"), strings.Join(defaultTagsList, ",\n\t"))
}
