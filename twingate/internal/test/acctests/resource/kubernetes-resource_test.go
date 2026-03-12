package resource

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func terraformResourceKubernetesResource(tfName, gatewayTFName, remoteNetworkTFName, name, address string) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName)
}

func terraformResourceKubernetesResourceWithInClusterField(tfName, gatewayTFName, remoteNetworkTFName, name, address string, inCluster bool) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  in_cluster        = %v
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, inCluster)
}

func TestAccTwingateKubernetesResource_InvalidAddress(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "10.0.0.1:9000"
	gatewayAddress := "10.0.0.1:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
				ExpectError: regexp.MustCompile(`address string must be a valid IP or FQDN`),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceCreate(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.1:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					sdk.TestCheckResourceAttr(theResource, attr.Name, resourceName),
					sdk.TestCheckResourceAttr(theResource, attr.Address, resourceAddress),
					sdk.TestCheckResourceAttrSet(theResource, attr.GatewayID),
					sdk.TestCheckResourceAttrSet(theResource, attr.RemoteNetworkID),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceUpdateName(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name1 := test.RandomName()
	name2 := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.2:8080"
	resourceID := new(string)

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, name1, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name1),
					acctests.GetTwingateResourceID(theResource, &resourceID),
				),
			},
			{
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, name2, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Name, name2),
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if *resourceID == "" {
							return errors.New("failed to fetch initial resource id")
						}

						if value != *resourceID {
							return errors.New("resource should not be re-created on name change")
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceUpdateInCluster(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.2:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResourceWithInClusterField(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, true),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.InCluster, "true"),
				),
			},
			{
				Config: prereqs + terraformResourceKubernetesResourceWithInClusterField(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, false),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.InCluster, "false"),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceDelete(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.5:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)
	config := prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  config,
				Destroy: true,
			},
			{
				Config: config,
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateKubernetesResourceReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.6:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)
	config := prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: config,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateKubernetesResource),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: config,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Address, resourceAddress),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceImport(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	resourceName := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.0.7:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, resourceName, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				ResourceName:      theResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func terraformResourceKubernetesResourceWithIsVisible(tfName, gatewayTFName, remoteNetworkTFName, name, address string, isVisible bool) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  is_visible        = %v
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, isVisible)
}

func terraformResourceKubernetesResourceWithAlias(tfName, gatewayTFName, remoteNetworkTFName, name, address, alias string) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  alias             = "%s"
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, alias)
}

func terraformResourceKubernetesResourceWithTags(tfName, gatewayTFName, remoteNetworkTFName, name, address string, tags map[string]string) string {
	tagLines := ""
	for k, v := range tags {
		tagLines += fmt.Sprintf(`    %s = "%s"`+"\n", k, v)
	}

	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  tags = {
%s  }
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, tagLines)
}

func terraformResourceKubernetesResourceWithProtocols(tfName, gatewayTFName, remoteNetworkTFName, name, address string) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  protocols = {
	    allow_icmp = true
	    tcp = {
	      policy = "RESTRICTED"
	      ports  = ["8080", "8443"]
	    }
	    udp = {
	      policy = "ALLOW_ALL"
	    }
	  }
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName)
}

func terraformResourceKubernetesResourceWithAccessGroup(tfName, gatewayTFName, remoteNetworkTFName, groupTFName, name, address string) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  access_group {
	    group_id = twingate_group.%s.id
	  }
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, groupTFName)
}

func terraformResourceKubernetesResourceWithAccessPolicy(tfName, gatewayTFName, remoteNetworkTFName, name, address, mode, duration, approvalMode string) string {
	return fmt.Sprintf(`
	resource "twingate_kubernetes_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  access_policy {
	    mode          = "%s"
	    duration      = "%s"
	    approval_mode = "%s"
	  }
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, mode, duration, approvalMode)
}

func TestAccTwingateKubernetesResourceIsVisible(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.1.1:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResourceWithIsVisible(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, true),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.IsVisible, "true"),
				),
			},
			{
				Config: prereqs + terraformResourceKubernetesResourceWithIsVisible(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, false),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.IsVisible, "false"),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceAlias(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.1.2:8080"
	alias := "k8s-alias.internal"
	newAlias := "k8s-alias.internal.new"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResourceWithAlias(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, alias),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Alias, alias),
				),
			},
			{
				Config: prereqs + terraformResourceKubernetesResourceWithAlias(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, newAlias),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Alias, newAlias),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceTags(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.1.3:8080"
	tags1 := map[string]string{"env": "staging", "team": "platform"}
	tags2 := map[string]string{"env": "production"}

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResourceWithTags(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, tags1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Tags, "env"), "staging"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Tags, "team"), "platform"),
				),
			},
			{
				Config: prereqs + terraformResourceKubernetesResourceWithTags(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, tags2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Tags, "env"), "production"),
					sdk.TestCheckNoResourceAttr(theResource, attr.PathAttr(attr.Tags, "team")),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceProtocols(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.1.4:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResourceWithProtocols(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Protocols, attr.AllowIcmp), "true"),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Protocols, attr.TCP, attr.Policy), model.PolicyRestricted),
					sdk.TestCheckResourceAttr(theResource, attr.PathAttr(attr.Protocols, attr.UDP, attr.Policy), model.PolicyAllowAll),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceAccessGroup(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	groupTFName := test.TerraformRandName("test_group")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.1.5:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)
	group := fmt.Sprintf(`resource "twingate_group" "%s" { name = "%s" }`, groupTFName, test.RandomName())

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + group + terraformResourceKubernetesResourceWithAccessGroup(k8sResTFName, gatewayTFName, remoteNetworkTFName, groupTFName, name, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Len(attr.AccessGroup), "1"),
					acctests.CheckResourceGroupsLen(theResource, 1),
				),
			},
			{
				// Remove access group
				Config: prereqs + group + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Len(attr.AccessGroup), "0"),
					acctests.CheckResourceGroupsLen(theResource, 0),
				),
			},
		},
	})
}

func TestAccTwingateKubernetesResourceAccessPolicy(t *testing.T) {
	t.Parallel()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	sshCATFName := test.TerraformRandName("test_ssh_ca")
	gatewayTFName := test.TerraformRandName("test_gw")
	k8sResTFName := test.TerraformRandName("test_k8s_res")
	theResource := acctests.TerraformKubernetesResource(k8sResTFName)
	certPEM := acctests.GenerateCACertPEM(t)
	publicKey := acctests.GenerateSSHPublicKey(t)
	name := test.RandomName()
	resourceAddress := "kubernetes.default.svc.cluster.local"
	gatewayAddress := "10.0.1.6:8080"

	prereqs := sshResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName, certPEM, sshCATFName, publicKey, gatewayTFName, gatewayAddress)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateKubernetesResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: prereqs + terraformResourceKubernetesResourceWithAccessPolicy(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress, model.AccessPolicyModeAutoLock, "2d", model.ApprovalModeManual),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Path(attr.AccessPolicy, attr.Mode), model.AccessPolicyModeAutoLock),
					sdk.TestCheckResourceAttr(theResource, attr.Path(attr.AccessPolicy, attr.Duration), "2d"),
					sdk.TestCheckResourceAttr(theResource, attr.Path(attr.AccessPolicy, attr.ApprovalMode), model.ApprovalModeManual),
				),
			},
			{
				// Update to MANUAL mode (no duration/approvalMode)
				Config: prereqs + terraformResourceKubernetesResource(k8sResTFName, gatewayTFName, remoteNetworkTFName, name, resourceAddress),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckNoResourceAttr(theResource, attr.Path(attr.AccessPolicy, attr.Mode)),
					sdk.TestCheckNoResourceAttr(theResource, attr.Path(attr.AccessPolicy, attr.Duration)),
					sdk.TestCheckNoResourceAttr(theResource, attr.Path(attr.AccessPolicy, attr.ApprovalMode)),
				),
			},
		},
	})
}
