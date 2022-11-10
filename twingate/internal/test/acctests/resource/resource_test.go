package resource

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/client"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	resourceResourceName = "resource"
	testResource         = "twingate_resource.test"
	groupIdsNumber       = "group_ids.#"
	tcpPolicy            = "protocols.0.tcp.0.policy"
	udpPolicy            = "protocols.0.udp.0.policy"
	firstTCPPort         = "protocols.0.tcp.0.ports.0"
	firstUDPPort         = "protocols.0.udp.0.ports.0"
	tcpPorts             = "protocols.0.tcp.0.ports.#"
	udpPorts             = "protocols.0.udp.0.ports.#"
)

func TestAccTwingateResourceCreate(t *testing.T) {
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceOnlyWithNetwork(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test1"),
					resource.TestCheckNoResourceAttr("twingate_resource.test1", "group_ids.#"),
				),
			},
		},
	})
}

func createResourceOnlyWithNetwork(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test1" {
	  name = "%s"
	}
	resource "twingate_resource" "test1" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test1.id
	}
	`, networkName, resourceName)
}

func TestAccTwingateResourceCreateWithProtocolsAndGroups(t *testing.T) {
	remoteNetworkName := test.RandomName()
	groupName1 := test.RandomGroupName()
	groupName2 := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithProtocolsAndGroups(remoteNetworkName, groupName1, groupName2, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test2"),
					resource.TestCheckResourceAttr("twingate_resource.test2", "address", "new-acc-test.com"),
					resource.TestCheckResourceAttr("twingate_resource.test2", "group_ids.#", "2"),
					resource.TestCheckResourceAttr("twingate_resource.test2", "protocols.0.tcp.0.policy", model.PolicyRestricted),
					resource.TestCheckResourceAttr("twingate_resource.test2", "protocols.0.tcp.0.ports.0", "80"),
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
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceFullCreationFlow(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("twingate_remote_network.test3", "name", remoteNetworkName),
					resource.TestCheckResourceAttr("twingate_resource.test3", "name", resourceName),
					resource.TestMatchResourceAttr("twingate_connector_tokens.test31", "access_token", regexp.MustCompile(".*")),
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

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
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
	resourceName := test.RandomResourceName()
	networkName := test.RandomResourceName()
	groupName := test.RandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithTcpDenyAllPolicy(networkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test5"),
					resource.TestCheckResourceAttr("twingate_resource.test5", "protocols.0.tcp.0.policy", model.PolicyRestricted),
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
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test6"),
					resource.TestCheckResourceAttr("twingate_resource.test6", "protocols.0.udp.0.policy", model.PolicyRestricted),
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
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithRestrictedPolicyAndEmptyPortsList(remoteNetworkName, groupName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("twingate_resource.test7", "name", resourceName),
					resource.TestCheckResourceAttr("twingate_resource.test7", "protocols.0.tcp.0.policy", model.PolicyRestricted),
					resource.TestCheckNoResourceAttr("twingate_resource.test7", "protocols.0.tcp.0.ports.#"),
					resource.TestCheckResourceAttr("twingate_resource.test7", "protocols.0.udp.0.policy", model.PolicyRestricted),
					resource.TestCheckNoResourceAttr("twingate_resource.test7", "protocols.0.udp.0.ports.#"),
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

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
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

func testAccCheckTwingateResourceDestroy(s *terraform.State) error {
	c := acctests.Provider.Meta().(*client.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_resource" {
			continue
		}

		resourceId := rs.Primary.ID

		err := c.DeleteResource(context.Background(), resourceId)
		// expecting error here , since the resource is already gone
		if err == nil {
			return fmt.Errorf("resource with ID %s still present : ", resourceId)
		}
	}

	return nil
}

func testAccCheckTwingateResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ResourceId set ")
		}

		return nil
	}
}

//func TestResourceResourceReadDiagnosticsError(t *testing.T) {
//	t.Parallel()
//	t.Run("Test Twingate Resource : Resource Read Diagnostics Error", func(t *testing.T) {
//		groups := []*graphql.ID{}
//		protocols := &transport.Protocols{}
//
//		res := &transport.Resource{
//			Name:            graphql.String(""),
//			RemoteNetworkID: graphql.ID(""),
//			Address:         graphql.String(""),
//			GroupsIds:       groups,
//			Protocols:       protocols,
//		}
//		d := &schema.ResourceData{}
//		diags := providerResource.readDiagnostics(d, res)
//		assert.True(t, diags.HasError())
//	})
//}

func TestAccTwingateResourcePortReorderingCreatesNoChanges(t *testing.T) {
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	const theResource = "twingate_resource.test9"

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"82-83", "80"`),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
					resource.TestCheckResourceAttr(theResource, "protocols.0.tcp.0.ports.0", "80"),
					resource.TestCheckResourceAttr(theResource, "protocols.0.udp.0.ports.0", "80"),
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
					testAccCheckTwingateResourceExists(theResource),
					resource.TestCheckResourceAttr(theResource, "protocols.0.tcp.0.ports.0", "70"),
					resource.TestCheckResourceAttr(theResource, "protocols.0.udp.0.ports.0", "70"),
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
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	const theResource = "twingate_resource.test10"

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResource10(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
					deactivateTwingateResource(theResource),
					acctests.WaitTestFunc(),
					testAccCheckTwingateResourceActiveState(theResource, false),
				),
			},
			{
				Config: createResource10(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceActiveState(theResource, true),
				),
			},
		},
	})
}

func createResource10(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test10" {
	  name = "%s"
	}
	resource "twingate_resource" "test10" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test10.id
	}
	`, networkName, resourceName)
}

func deactivateTwingateResource(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := acctests.Provider.Meta().(*client.Client)

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		resourceId := rs.Primary.ID

		if resourceId == "" {
			return fmt.Errorf("No ResourceId set ")
		}

		err := c.UpdateResourceActiveState(context.Background(), &model.Resource{
			ID:       resourceId,
			IsActive: false,
		})

		if err != nil {
			return fmt.Errorf("resource with ID %s still active: %w", resourceId, err)
		}

		return nil
	}
}

func testAccCheckTwingateResourceActiveState(resourceName string, expectedActiveState bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := acctests.Provider.Meta().(*client.Client)

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ResourceId set ")
		}

		res, err := c.ReadResource(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read resource: %w", err)
		}

		if res.IsActive != expectedActiveState {
			return fmt.Errorf("expected active state %v, got %v", expectedActiveState, res.IsActive)
		}

		return nil
	}
}

func TestAccTwingateResourceReCreateAfterDeletion(t *testing.T) {
	const theResource = "twingate_resource.test11"

	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()
	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResource11(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
					deleteTwingateResource(theResource, resourceResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource11(remoteNetworkName, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func createResource11(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test11" {
	  name = "%s"
	}
	resource "twingate_resource" "test11" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test11.id
	}
	`, networkName, resourceName)
}

func deleteTwingateResource(resourceName, resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := acctests.Provider.Meta().(*client.Client)

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		resourceId := rs.Primary.ID

		if resourceId == "" {
			return fmt.Errorf("No ResourceId set ")
		}

		var err error
		switch resourceType {
		case resourceResourceName:
			err = c.DeleteResource(context.Background(), resourceId)
		case remoteNetworkResourceName:
			err = c.DeleteRemoteNetwork(context.Background(), resourceId)
		case groupResourceName:
			err = c.DeleteGroup(context.Background(), resourceId)
		case connectorResourceName:
			err = c.DeleteConnector(context.Background(), resourceId)
		default:
			return fmt.Errorf("%s unknown resource type", resourceType)
		}

		if err != nil {
			return fmt.Errorf("%s with ID %s still active: %w", resourceType, resourceId, err)
		}

		return nil
	}
}

func TestAccTwingateResourceImport(t *testing.T) {
	remoteNetworkName := test.RandomName()
	groupName := test.RandomGroupName()
	groupName2 := test.RandomGroupName()
	resourceName := test.RandomResourceName()

	const theResource = "twingate_resource.test12"

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResource12(remoteNetworkName, groupName, groupName2, resourceName),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
				),
			},
			{
				ImportState:  true,
				ResourceName: theResource,
				ImportStateCheck: func(data []*terraform.InstanceState) error {
					if len(data) != 1 {
						return fmt.Errorf("expected 1 resource, got %d", len(data))
					}

					attributes := []struct {
						name     string
						expected string
					}{
						{name: "address", expected: "acc-test.com.12"},
						{name: "protocols.0.tcp.0.policy", expected: model.PolicyRestricted},
						{name: "protocols.0.tcp.0.ports.#", expected: "2"},
						{name: "protocols.0.tcp.0.ports.0", expected: "80"},
						{name: "protocols.0.udp.0.policy", expected: model.PolicyAllowAll},
						{name: "protocols.0.udp.0.ports.#", expected: "0"},
					}

					res := data[0]
					for _, attr := range attributes {
						if res.Attributes[attr.name] != attr.expected {
							return fmt.Errorf("attribute %s doesn't match, expected: %s, got: %s", attr.name, attr.expected, res.Attributes[attr.name])
						}
					}

					return nil
				},
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
	remoteNetworkName := test.RandomName()
	resourceName := test.RandomResourceName()

	const theResource = "twingate_resource.test13"

	groups, groupsID := genNewGroups("g13", 111)

	resource.Test(t, resource.TestCase{
		ProviderFactories: acctests.ProviderFactories,
		PreCheck:          func() { acctests.PreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResource13(remoteNetworkName, resourceName, groups, groupsID),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
					resource.TestCheckResourceAttr(theResource, "group_ids.#", "111"),
				),
			},
			{
				Config: createResource13(remoteNetworkName, resourceName, groups[:75], groupsID[:75]),
				Check: acctests.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
					resource.TestCheckResourceAttr(theResource, "group_ids.#", "75"),
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
