package twingate

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

const (
	testResource   = "twingate_resource.test"
	groupIdsNumber = "group_ids.#"
	tcpPolicy      = "protocols.0.tcp.0.policy"
	udpPolicy      = "protocols.0.udp.0.policy"
	firstTCPPort   = "protocols.0.tcp.0.ports.0"
	firstUDPPort   = "protocols.0.udp.0.ports.0"
	tcpPorts       = "protocols.0.tcp.0.ports.#"
	udpPorts       = "protocols.0.udp.0.ports.#"
)

func TestAccTwingateResourceCreate(t *testing.T) {
	remoteNetworkName := getRandomName()
	resourceName := getRandomResourceName()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceOnlyWithNetwork(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
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
	remoteNetworkName := getRandomName()
	groupName1 := getRandomGroupName()
	groupName2 := getRandomGroupName()
	resourceName := getRandomResourceName()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithProtocolsAndGroups(remoteNetworkName, groupName1, groupName2, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test2"),
					resource.TestCheckResourceAttr("twingate_resource.test2", "address", "new-acc-test.com"),
					resource.TestCheckResourceAttr("twingate_resource.test2", "group_ids.#", "2"),
					resource.TestCheckResourceAttr("twingate_resource.test2", "protocols.0.tcp.0.policy", policyRestricted),
					resource.TestCheckResourceAttr("twingate_resource.test2", "protocols.0.tcp.0.ports.0", "80"),
				),
				// group is updated with linked resource
				ExpectNonEmptyPlan: true,
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
	`, networkName, groupName1, groupName2, resourceName, policyRestricted, policyAllowAll)
}

func TestAccTwingateResourceFullCreationFlow(t *testing.T) {
	remoteNetworkName := getRandomName()
	groupName := getRandomGroupName()
	resourceName := getRandomResourceName()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceFullCreationFlow(remoteNetworkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("twingate_remote_network.test3", "name", remoteNetworkName),
					resource.TestCheckResourceAttr("twingate_resource.test3", "name", resourceName),
					resource.TestMatchResourceAttr("twingate_connector_tokens.test31", "access_token", regexp.MustCompile(".*")),
				),
				// group is updated with linked resource
				ExpectNonEmptyPlan: true,
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
	`, networkName, groupName, resourceName, policyRestricted, policyAllowAll)
}

func TestAccTwingateResourceWithInvalidGroupId(t *testing.T) {
	resourceName := getRandomResourceName()
	networkName := getRandomResourceName()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
	resourceName := getRandomResourceName()
	networkName := getRandomResourceName()
	groupName := getRandomResourceName()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithTcpDenyAllPolicy(networkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test5"),
					resource.TestCheckResourceAttr("twingate_resource.test5", "protocols.0.tcp.0.policy", policyRestricted),
				),
				// group is updated with linked resource
				ExpectNonEmptyPlan: true,
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
    `, networkName, groupName, resourceName, policyDenyAll, policyAllowAll)
}

func TestAccTwingateResourceWithUdpDenyAllPolicy(t *testing.T) {
	remoteNetworkName := getRandomName()
	groupName := getRandomGroupName()
	resourceName := getRandomResourceName()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test6"),
					resource.TestCheckResourceAttr("twingate_resource.test6", "protocols.0.udp.0.policy", policyRestricted),
				),
				// group is updated with linked resource
				ExpectNonEmptyPlan: true,
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
    `, networkName, groupName, resourceName, policyAllowAll, policyDenyAll)
}

func TestAccTwingateResourceWithRestrictedPolicyAndEmptyPortsList(t *testing.T) {
	remoteNetworkName := getRandomName()
	groupName := getRandomGroupName()
	resourceName := getRandomResourceName()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithRestrictedPolicyAndEmptyPortsList(remoteNetworkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("twingate_resource.test7", "name", resourceName),
					resource.TestCheckResourceAttr("twingate_resource.test7", "protocols.0.tcp.0.policy", policyRestricted),
					resource.TestCheckNoResourceAttr("twingate_resource.test7", "protocols.0.tcp.0.ports.#"),
					resource.TestCheckResourceAttr("twingate_resource.test7", "protocols.0.udp.0.policy", policyRestricted),
					resource.TestCheckNoResourceAttr("twingate_resource.test7", "protocols.0.udp.0.ports.#"),
				),
				// group is updated with linked resource
				ExpectNonEmptyPlan: true,
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
	`, networkName, groupName, resourceName, policyRestricted, policyRestricted)
}

func TestAccTwingateResourceWithInvalidPortRange(t *testing.T) {
	remoteNetworkName := getRandomName()
	resourceName := getRandomResourceName()
	expectedError := regexp.MustCompile("Error: failed to parse protocols port range")

	genConfig := func(portRange string) string {
		return createResourceWithRestrictedPolicyAndPortRange(remoteNetworkName, resourceName, portRange)
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
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
	`, networkName, resourceName, policyRestricted, portRange, policyAllowAll)
}

func testAccCheckTwingateResourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_resource" {
			continue
		}

		resourceId := rs.Primary.ID

		err := client.deleteResource(context.Background(), resourceId)
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

func TestResourceResourceReadDiagnosticsError(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Resource : Resource Read Diagnostics Error", func(t *testing.T) {
		groups := []*graphql.ID{}
		protocols := &ProtocolsInput{}

		res := &Resource{
			Name:            graphql.String(""),
			RemoteNetworkID: graphql.ID(""),
			Address:         graphql.String(""),
			GroupsIds:       groups,
			Protocols:       protocols,
		}
		d := &schema.ResourceData{}
		diags := resourceResourceReadDiagnostics(d, res)
		assert.True(t, diags.HasError())
	})
}

func TestAccTwingateResourcePortReorderingCreatesNoChanges(t *testing.T) {
	remoteNetworkName := getRandomName()
	resourceName := getRandomResourceName()

	const theResource = "twingate_resource.test9"

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResourceWithPortRange(remoteNetworkName, resourceName, `"82-83", "80"`),
				Check: resource.ComposeTestCheckFunc(
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
				Check: resource.ComposeTestCheckFunc(
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
	`, networkName, resourceName, policyRestricted, portRange, policyRestricted, portRange)
}

func TestAccTwingateResourceSetActiveStateOnUpdate(t *testing.T) {
	remoteNetworkName := getRandomName()
	resourceName := getRandomResourceName()

	const theResource = "twingate_resource.test10"

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResource10(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
					deactivateTwingateResource(theResource),
					testAccCheckTwingateResourceActiveState(theResource, false),
				),
			},
			{
				Config: createResource10(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
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
		client := testAccProvider.Meta().(*Client)

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		resourceId := rs.Primary.ID

		if resourceId == "" {
			return fmt.Errorf("No ResourceId set ")
		}

		err := client.updateResourceActiveState(context.Background(), &Resource{
			ID:       graphql.ID(resourceId),
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
		client := testAccProvider.Meta().(*Client)

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ResourceId set ")
		}

		resource, err := client.readResource(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read resource: %w", err)
		}

		if bool(resource.IsActive) != expectedActiveState {
			return fmt.Errorf("expected active state %v, got %v", expectedActiveState, resource.IsActive)
		}

		return nil
	}
}

func TestAccTwingateResourceReCreateAfterDeletion(t *testing.T) {
	const theResource = "twingate_resource.test11"
	remoteNetworkName := getRandomName()
	resourceName := getRandomResourceName()

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResource11(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
					deleteTwingateResource(theResource, resourceResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: createResource11(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
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
		client := testAccProvider.Meta().(*Client)

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
			err = client.deleteResource(context.Background(), resourceId)
		case remoteNetworkResourceName:
			err = client.deleteRemoteNetwork(context.Background(), resourceId)
		case groupResourceName:
			err = client.deleteGroup(context.Background(), resourceId)
		case connectorResourceName:
			err = client.deleteConnector(context.Background(), resourceId)
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
	remoteNetworkName := getRandomName()
	groupName := getRandomGroupName()
	groupName2 := getRandomGroupName()
	resourceName := getRandomResourceName()

	const theResource = "twingate_resource.test12"

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: createResource12(remoteNetworkName, groupName, groupName2, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(theResource),
				),
				// group is updated with linked resource
				ExpectNonEmptyPlan: true,
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
						{name: "protocols.0.tcp.0.policy", expected: policyRestricted},
						{name: "protocols.0.tcp.0.ports.#", expected: "2"},
						{name: "protocols.0.tcp.0.ports.0", expected: "80"},
						{name: "protocols.0.udp.0.policy", expected: policyAllowAll},
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
	`, networkName, groupName1, groupName2, resourceName, policyRestricted, policyAllowAll)
}
