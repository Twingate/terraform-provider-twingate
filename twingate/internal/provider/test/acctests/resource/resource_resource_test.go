package resource

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

func TestAccTwingateResource_basic(t *testing.T) {
	remoteNetworkName := getRandomName()
	groupName := getRandomGroupName()
	groupName2 := getRandomGroupName()
	resourceName := getRandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
					resource.TestCheckNoResourceAttr(testResource, groupIdsNumber),
				),
			},
			{
				Config: testTwingateResource_withProtocolsAndGroups(remoteNetworkName, groupName, groupName2, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
					resource.TestCheckResourceAttr(testResource, "address", "updated-acc-test.com"),
					resource.TestCheckResourceAttr(testResource, groupIdsNumber, "2"),
					resource.TestCheckResourceAttr(testResource, tcpPolicy, policyRestricted),
					resource.TestCheckResourceAttr(testResource, firstTCPPort, "80"),
				),
			},
			{
				Config: testTwingateResource_fullFlowCreation(remoteNetworkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(remoteNetworkResource, nameAttr, remoteNetworkName),
					resource.TestCheckResourceAttr(testResource, nameAttr, resourceName),
					resource.TestMatchResourceAttr("twingate_connector_tokens.test_1", "access_token", regexp.MustCompile(".*")),
				),
			},
			{
				Config:      testTwingateResource_errorGroupId(remoteNetworkName, resourceName),
				ExpectError: regexp.MustCompile("Error: failed to update resource with id"),
			},
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
					resource.TestCheckNoResourceAttr(testResource, groupIdsNumber),
				),
			},
			{
				Config: testTwingateResource_withTcpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
					resource.TestCheckResourceAttr(testResource, tcpPolicy, policyRestricted),
				),
			},
			// expecting no changes - empty plan
			{
				Config:   testTwingateResource_withTcpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				PlanOnly: true,
			},
			{
				Config: testTwingateResource_withUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
					resource.TestCheckResourceAttr(testResource, udpPolicy, policyRestricted),
				),
			},
			// expecting no changes - empty plan
			{
				Config:   testTwingateResource_withUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
				PlanOnly: true,
			},
		},
	})
}

func testTwingateResource_Simple(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
	}
	`, networkName, resourceName)
}

func testTwingateResource_withProtocolsAndGroups(networkName, groupName1, groupName2, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}

    resource "twingate_group" "g1" {
      name = "%s"
    }

    resource "twingate_group" "g2" {
      name = "%s"
    }

	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "updated-acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
	  group_ids = [twingate_group.g1.id, twingate_group.g2.id]
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

func testTwingateResource_withTcpDenyAllPolicy(networkName, groupName, resourceName string) string {
	return fmt.Sprintf(`
    resource "twingate_remote_network" "test" {
      name = "%s"
    }

    resource "twingate_group" "g" {
      name = "%s"
    }

    resource "twingate_resource" "test" {
      name = "%s"
      address = "updated-acc-test.com"
      remote_network_id = twingate_remote_network.test.id
      group_ids = [twingate_group.g.id]
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

func testTwingateResource_withUdpDenyAllPolicy(networkName, groupName, resourceName string) string {
	return fmt.Sprintf(`
    resource "twingate_remote_network" "test" {
      name = "%s"
    }

    resource "twingate_group" "g" {
      name = "%s"
    }

    resource "twingate_resource" "test" {
      name = "%s"
      address = "updated-acc-test.com"
      remote_network_id = twingate_remote_network.test.id
      group_ids = [twingate_group.g.id]
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

func testTwingateResource_errorGroupId(networkName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}
	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "updated-acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
	  group_ids = ["foo", "bar"]
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
	`, networkName, resourceName, policyRestricted, policyAllowAll)
}

func testTwingateResource_fullFlowCreation(networkName, groupName, resourceName string) string {
	return fmt.Sprintf(`
    resource "twingate_remote_network" "test" {
      name = "%s"
    }
	
    resource "twingate_connector" "test_1" {
      remote_network_id = twingate_remote_network.test.id
    }

    resource "twingate_connector_tokens" "test_1" {
      connector_id = twingate_connector.test_1.id
    }

    resource "twingate_connector" "test_2" {
      remote_network_id = twingate_remote_network.test.id
    }
	
    resource "twingate_connector_tokens" "test_2" {
      connector_id = twingate_connector.test_2.id
    }

    resource "twingate_group" "test_res" {
      name = "%s"
    }

    resource "twingate_resource" "test" {
      name = "%s"
      address = "updated-acc-test.com"
      remote_network_id = twingate_remote_network.test.id
      group_ids = [twingate_group.test_res.id]
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

func TestAccTwingateResource_restrictedPolicyWithEmptyPortsList(t *testing.T) {
	remoteNetworkName := getRandomName()
	groupName := getRandomGroupName()
	resourceName := getRandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_restrictedPolicyWithEmptyPortsList(remoteNetworkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResource, nameAttr, resourceName),
					resource.TestCheckResourceAttr(testResource, tcpPolicy, policyRestricted),
					resource.TestCheckNoResourceAttr(testResource, tcpPorts),
					resource.TestCheckResourceAttr(testResource, udpPolicy, policyRestricted),
					resource.TestCheckNoResourceAttr(testResource, udpPorts),
				),
			},
		},
	})
}

func testTwingateResource_restrictedPolicyWithEmptyPortsList(networkName, groupName, resourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}

    resource "twingate_group" "test_res" {
      name = "%s"
    }

	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "updated-acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
	  group_ids = [twingate_group.test_res.id]
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

func TestAccTwingateResource_withInvalidPortRange(t *testing.T) {
	remoteNetworkName := getRandomName()
	groupName := getRandomGroupName()
	resourceName := getRandomResourceName()
	expectedError := regexp.MustCompile("Error: failed to parse protocols port range")

	genConfig := func(portRange string) string {
		return testTwingateResource_restrictedWithPortRange(remoteNetworkName, groupName, resourceName, portRange)
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
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

func testTwingateResource_restrictedWithPortRange(networkName, groupName, resourceName, portRange string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}

    resource "twingate_group" "test_res" {
      name = "%s"
    }

	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "updated-acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
	  group_ids = [twingate_group.test_res.id]
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
	`, networkName, groupName, resourceName, policyRestricted, portRange, policyAllowAll)
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

func TestAccTwingateResource_portReorderingCreatesNoChanges(t *testing.T) {
	remoteNetworkName := getRandomName()
	resourceName := getRandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_withPortRange(remoteNetworkName, resourceName, `"82-83", "80"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
					resource.TestCheckResourceAttr(testResource, firstTCPPort, "80"),
					resource.TestCheckResourceAttr(testResource, firstUDPPort, "80"),
				),
			},
			// no changes
			{
				Config:   testTwingateResource_withPortRange(remoteNetworkName, resourceName, `"82-83", "80"`),
				PlanOnly: true,
			},
			// no changes
			{
				Config:   testTwingateResource_withPortRange(remoteNetworkName, resourceName, `"82", "83", "80"`),
				PlanOnly: true,
			},
			// new changes applied
			{
				Config: testTwingateResource_withPortRange(remoteNetworkName, resourceName, `"82-83", "70"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
					resource.TestCheckResourceAttr(testResource, "protocols.0.tcp.0.ports.0", "70"),
					resource.TestCheckResourceAttr(testResource, "protocols.0.udp.0.ports.0", "70"),
				),
			},
		},
	})
}

func testTwingateResource_withPortRange(networkName, resourceName, portRange string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "test" {
	  name = "%s"
	}

	resource "twingate_resource" "test" {
	  name = "%s"
	  address = "acc-test.com"
	  remote_network_id = twingate_remote_network.test.id
	  protocols {
	    allow_icmp = true
	    tcp {
	      policy = "RESTRICTED"
	      ports = [%s]
	    }
	    udp {
	      policy = "RESTRICTED"
	      ports = [%s]
	    }
	  }
	}
	`, networkName, resourceName, portRange, portRange)
}

func TestAccTwingateResource_setActiveState(t *testing.T) {
	remoteNetworkName := getRandomName()
	resourceName := getRandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
					deactivateTwingateResource(testResource),
					testAccCheckTwingateResourceActiveState(testResource, false),
				),
			},
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceActiveState(testResource, true),
				),
			},
		},
	})
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

func TestAccTwingateResource_createAfterDeletion(t *testing.T) {
	const terraformResourceName = testResource
	remoteNetworkName := getRandomName()
	resourceName := getRandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(terraformResourceName),
					deleteTwingateResource(terraformResourceName, resourceResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(terraformResourceName),
				),
			},
		},
	})
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

func TestAccTwingateResource_import(t *testing.T) {
	remoteNetworkName := getRandomName()
	groupName := getRandomGroupName()
	groupName2 := getRandomGroupName()
	resourceName := getRandomResourceName()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_withProtocolsAndGroups(remoteNetworkName, groupName, groupName2, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists(testResource),
				),
			},
			{
				ImportState:  true,
				ResourceName: testResource,
				ImportStateCheck: func(data []*terraform.InstanceState) error {
					if len(data) != 1 {
						return fmt.Errorf("expected 1 resource, got %d", len(data))
					}

					attributes := []struct {
						name     string
						expected string
					}{
						{name: "address", expected: "updated-acc-test.com"},
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
