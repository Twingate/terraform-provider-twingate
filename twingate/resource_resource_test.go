package twingate

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

func TestAccTwingateResource_basic(t *testing.T) {
	remoteNetworkName := acctest.RandomWithPrefix(testPrefixName)
	groupName := acctest.RandomWithPrefix(testPrefixName + "-group")
	groupName2 := acctest.RandomWithPrefix(testPrefixName + "-group")
	resourceName := acctest.RandomWithPrefix(testPrefixName + "-resource")

	_, _ = groupName, groupName2

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test"),
					resource.TestCheckNoResourceAttr("twingate_resource.test", "group_ids.#"),
				),
			},
			//{
			//	Config: testTwingateResource_withProtocolsAndGroups(remoteNetworkName, groupName, groupName2, resourceName),
			//	Check: resource.ComposeTestCheckFunc(
			//		testAccCheckTwingateResourceExists("twingate_resource.test"),
			//		resource.TestCheckResourceAttr("twingate_resource.test", "address", "updated-acc-test.com"),
			//		resource.TestCheckResourceAttr("twingate_resource.test", "group_ids.#", "2"),
			//		resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.tcp.0.policy", policyRestricted),
			//		resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.tcp.0.ports.0", "80"),
			//	),
			//},
			//{
			//	Config: testTwingateResource_fullFlowCreation(remoteNetworkName, groupName, resourceName),
			//	Check: resource.ComposeTestCheckFunc(
			//		resource.TestCheckResourceAttr("twingate_remote_network.test", "name", remoteNetworkName),
			//		resource.TestCheckResourceAttr("twingate_resource.test", "name", resourceName),
			//		resource.TestMatchResourceAttr("twingate_connector_tokens.test_1", "access_token", regexp.MustCompile(".*")),
			//	),
			//},
			//{
			//	Config:      testTwingateResource_errorGroupId(remoteNetworkName, resourceName),
			//	ExpectError: regexp.MustCompile("Error: failed to update resource with id"),
			//},
			//{
			//	Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
			//	Check: resource.ComposeTestCheckFunc(
			//		testAccCheckTwingateResourceExists("twingate_resource.test"),
			//		resource.TestCheckNoResourceAttr("twingate_resource.test", "group_ids.#"),
			//		resource.TestCheckNoResourceAttr("twingate_resource.test", "protocols.0.tcp.0.ports.0"),
			//	),
			//},
			//{
			//	Config: testTwingateResource_withTcpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
			//	Check: resource.ComposeTestCheckFunc(
			//		testAccCheckTwingateResourceExists("twingate_resource.test"),
			//		resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.tcp.0.policy", policyRestricted),
			//	),
			//},
			//// expecting no changes - empty plan
			//{
			//	Config:   testTwingateResource_withTcpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
			//	PlanOnly: true,
			//},
			//{
			//	Config: testTwingateResource_withUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
			//	Check: resource.ComposeTestCheckFunc(
			//		testAccCheckTwingateResourceExists("twingate_resource.test"),
			//		resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.udp.0.policy", policyRestricted),
			//	),
			//},
			//// expecting no changes - empty plan
			//{
			//	Config:   testTwingateResource_withUdpDenyAllPolicy(remoteNetworkName, groupName, resourceName),
			//	PlanOnly: true,
			//},
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
	remoteNetworkName := acctest.RandomWithPrefix(testPrefixName)
	groupName := acctest.RandomWithPrefix(testPrefixName + "-group")
	resourceName := acctest.RandomWithPrefix(testPrefixName + "-resource")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_restrictedPolicyWithEmptyPortsList(remoteNetworkName, groupName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("twingate_resource.test", "name", resourceName),
					resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.tcp.0.policy", policyRestricted),
					resource.TestCheckNoResourceAttr("twingate_resource.test", "protocols.0.tcp.0.ports.#"),
					resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.udp.0.policy", policyRestricted),
					resource.TestCheckNoResourceAttr("twingate_resource.test", "protocols.0.udp.0.ports.#"),
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
	remoteNetworkName := acctest.RandomWithPrefix(testPrefixName)
	groupName := acctest.RandomWithPrefix(testPrefixName + "-group")
	resourceName := acctest.RandomWithPrefix(testPrefixName + "-resource")
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
	remoteNetworkName := acctest.RandomWithPrefix(testPrefixName)
	resourceName := acctest.RandomWithPrefix(testPrefixName + "-resource")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_withPortRange(remoteNetworkName, resourceName, `"82-83", "80"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test"),
					resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.tcp.0.ports.0", "80"),
					resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.udp.0.ports.0", "80"),
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
					testAccCheckTwingateResourceExists("twingate_resource.test"),
					resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.tcp.0.ports.0", "70"),
					resource.TestCheckResourceAttr("twingate_resource.test", "protocols.0.udp.0.ports.0", "70"),
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
	remoteNetworkName := acctest.RandomWithPrefix(testPrefixName)
	resourceName := acctest.RandomWithPrefix(testPrefixName + "-resource")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckTwingateResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceExists("twingate_resource.test"),
					deactivateTwingateResource("twingate_resource.test"),
					testAccCheckTwingateResourceActiveState("twingate_resource.test", false),
				),
			},
			{
				Config: testTwingateResource_Simple(remoteNetworkName, resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTwingateResourceActiveState("twingate_resource.test", true),
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
	const terraformResourceName = "twingate_resource.test"
	remoteNetworkName := acctest.RandomWithPrefix(testPrefixName)
	resourceName := acctest.RandomWithPrefix(testPrefixName + "-resource")

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
