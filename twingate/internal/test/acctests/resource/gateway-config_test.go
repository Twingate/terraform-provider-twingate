package resource

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"gopkg.in/yaml.v3"
)

// checkYAMLContent parses the content attribute as YAML and runs the provided check against the document.
func checkYAMLContent(resourceName string, checkFn func(doc map[string]any) error) sdk.TestCheckFunc {
	return sdk.TestCheckResourceAttrWith(resourceName, attr.Content, func(value string) error {
		var doc map[string]any
		if err := yaml.Unmarshal([]byte(value), &doc); err != nil {
			return fmt.Errorf("content is not valid YAML: %w\n---\n%s", err, value)
		}
		return checkFn(doc)
	})
}

func gatewayConfigWithSSHOnly(tfName string, sshName, sshAddress, sshUsername string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  ssh_resources = [
	    {
	      name     = "%s"
	      address  = "%s"
	      username = "%s"
	    }
	  ]
	  kubernetes_resources = []
	}
	`, tfName, sshName, sshAddress, sshUsername)
}

func gatewayConfigWithK8sOnly(tfName string, k8sName, k8sAddress string, inCluster bool) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  ssh_resources = []
	  kubernetes_resources = [
	    {
	      name       = "%s"
	      address    = "%s"
	      in_cluster = %v
	    }
	  ]
	}
	`, tfName, k8sName, k8sAddress, inCluster)
}

func gatewayConfigWithBoth(tfName string, sshName, sshAddress, sshUsername, k8sName, k8sAddress string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  ssh_resources = [
	    {
	      name     = "%s"
	      address  = "%s"
	      username = "%s"
	    }
	  ]
	  kubernetes_resources = [
	    {
	      name       = "%s"
	      address    = "%s"
	      in_cluster = true
	    }
	  ]
	}
	`, tfName, sshName, sshAddress, sshUsername, k8sName, k8sAddress)
}

func gatewayConfigWithSshCA(tfName, sshName, sshAddress, sshUsername, vaultAddr string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  ssh_ca = {
	    vault_addr = "%s"
	  }
	  ssh_resources = [
	    {
	      name     = "%s"
	      address  = "%s"
	      username = "%s"
	    }
	  ]
	  kubernetes_resources = []
	}
	`, tfName, vaultAddr, sshName, sshAddress, sshUsername)
}

func gatewayConfigWithPrivateKeyCA(tfName, sshName, sshAddress, sshUsername, keyFile string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  ssh_ca = {
	    private_key_file = "%s"
	  }
	  ssh_resources = [
	    {
	      name     = "%s"
	      address  = "%s"
	      username = "%s"
	    }
	  ]
	  kubernetes_resources = []
	}
	`, tfName, keyFile, sshName, sshAddress, sshUsername)
}

func gatewayConfigWithConflictingCA(tfName, sshName, sshAddress, sshUsername string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  ssh_ca = {
	    vault_addr       = "https://vault.example.com"
	    private_key_file = "/etc/ssh/id_ed25519"
	  }
	  ssh_resources = [
	    {
	      name     = "%s"
	      address  = "%s"
	      username = "%s"
	    }
	  ]
	  kubernetes_resources = []
	}
	`, tfName, sshName, sshAddress, sshUsername)
}

func gatewayConfigWithSSHNoUsername(tfName string, sshName, sshAddress string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  ssh_resources = [
	    {
	      name    = "%s"
	      address = "%s"
	      username = ""
	    }
	  ]
	  kubernetes_resources = []
	}
	`, tfName, sshName, sshAddress)
}

func gatewayConfigBothEmpty(tfName string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  ssh_resources        = []
	  kubernetes_resources = []
	}
	`, tfName)
}

func gatewayConfigBothOmitted(tfName string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	}
	`, tfName)
}

func gatewayConfigWithCustomPort(tfName string, port, metricsPort int, sshName, sshAddress, sshUsername string) string {
	return fmt.Sprintf(`
	resource "twingate_gateway_config" "%s" {
	  port         = %d
	  metrics_port = %d
	  ssh_resources = [
	    {
	      name     = "%s"
	      address  = "%s"
	      username = "%s"
	    }
	  ]
	  kubernetes_resources = []
	}
	`, tfName, port, metricsPort, sshName, sshAddress, sshUsername)
}

func TestAccTwingateGatewayConfigCreate_WithSSHOnly(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")
	theResource := acctests.TerraformGatewayConfig(tfName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: gatewayConfigWithSSHOnly(tfName, "web", "10.0.0.1", "ubuntu"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						ssh, ok := doc["ssh"].(map[string]any)
						if !ok {
							return fmt.Errorf("expected ssh block to be present")
						}
						if _, exists := doc["kubernetes"]; exists {
							return fmt.Errorf("expected kubernetes block to be absent when kubernetes_resources is empty")
						}
						upstreams, ok := ssh["upstreams"].([]any)
						if !ok || len(upstreams) != 1 {
							return fmt.Errorf("expected 1 ssh upstream, got %v", ssh["upstreams"])
						}
						u := upstreams[0].(map[string]any)
						if u["name"] != "web" {
							return fmt.Errorf("expected upstream name 'web', got %v", u["name"])
						}
						if u["address"] != "10.0.0.1" {
							return fmt.Errorf("expected upstream address '10.0.0.1', got %v", u["address"])
						}
						if u["user"] != "ubuntu" {
							return fmt.Errorf("expected upstream user 'ubuntu', got %v", u["user"])
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayConfigCreate_WithSSHNoUsername(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")
	theResource := acctests.TerraformGatewayConfig(tfName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: gatewayConfigWithSSHNoUsername(tfName, "web", "10.0.0.1"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						ssh, ok := doc["ssh"].(map[string]any)
						if !ok {
							return fmt.Errorf("expected ssh block to be present")
						}
						upstreams, ok := ssh["upstreams"].([]any)
						if !ok || len(upstreams) != 1 {
							return fmt.Errorf("expected 1 ssh upstream, got %v", ssh["upstreams"])
						}
						u := upstreams[0].(map[string]any)
						if _, exists := u["user"]; exists {
							return fmt.Errorf("expected user field to be absent when username is not set, got %v", u["user"])
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayConfigCreate_WithKubernetesOnly(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")
	theResource := acctests.TerraformGatewayConfig(tfName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: gatewayConfigWithK8sOnly(tfName, "prod-cluster", "10.0.0.2:6443", true),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						k8s, ok := doc["kubernetes"].(map[string]any)
						if !ok {
							return fmt.Errorf("expected kubernetes block to be present")
						}
						if _, exists := doc["ssh"]; exists {
							return fmt.Errorf("expected ssh block to be absent when ssh_resources is empty")
						}
						upstreams, ok := k8s["upstreams"].([]any)
						if !ok || len(upstreams) != 1 {
							return fmt.Errorf("expected 1 kubernetes upstream, got %v", k8s["upstreams"])
						}
						u := upstreams[0].(map[string]any)
						if u["name"] != "prod-cluster" {
							return fmt.Errorf("expected upstream name 'prod-cluster', got %v", u["name"])
						}
						if u["address"] != "10.0.0.2:6443" {
							return fmt.Errorf("expected upstream address '10.0.0.2:6443', got %v", u["address"])
						}
						if u["inCluster"] != true {
							return fmt.Errorf("expected inCluster true, got %v", u["inCluster"])
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayConfigCreate_WithBoth(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")
	theResource := acctests.TerraformGatewayConfig(tfName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: gatewayConfigWithBoth(tfName, "web", "10.0.0.1", "ubuntu", "prod-cluster", "10.0.0.2:6443"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						if _, ok := doc["ssh"].(map[string]any); !ok {
							return fmt.Errorf("expected ssh block to be present")
						}
						if _, ok := doc["kubernetes"].(map[string]any); !ok {
							return fmt.Errorf("expected kubernetes block to be present")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayConfig_BothResourcesEmpty(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      gatewayConfigBothEmpty(tfName),
				ExpectError: regexp.MustCompile(`At least one of "ssh_resources" or "kubernetes_resources" must contain`),
			},
		},
	})
}

func TestAccTwingateGatewayConfig_BothResourcesOmitted(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      gatewayConfigBothOmitted(tfName),
				ExpectError: regexp.MustCompile(`At least one of "ssh_resources" or "kubernetes_resources" must contain`),
			},
		},
	})
}

func TestAccTwingateGatewayConfig_SshCAConflictsWith(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config:      gatewayConfigWithConflictingCA(tfName, "web", "10.0.0.1", "ubuntu"),
				ExpectError: regexp.MustCompile(`Attribute "ssh_ca.private_key_file" cannot be specified`),
			},
		},
	})
}

func TestAccTwingateGatewayConfig_SshCAWithVault(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")
	theResource := acctests.TerraformGatewayConfig(tfName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: gatewayConfigWithSshCA(tfName, "web", "10.0.0.1", "ubuntu", "https://vault.example.com"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						ssh := doc["ssh"].(map[string]any)
						ca, ok := ssh["ca"].(map[string]any)
						if !ok {
							return fmt.Errorf("expected ssh.ca block to be present")
						}
						vault, ok := ca["vault"].(map[string]any)
						if !ok {
							return fmt.Errorf("expected ssh.ca.vault block to be present")
						}
						if vault["address"] != "https://vault.example.com" {
							return fmt.Errorf("expected vault address 'https://vault.example.com', got %v", vault["address"])
						}
						if vault["caBundleFile"] != "/etc/ssl/vault-ca.crt" {
							return fmt.Errorf("expected caBundleFile '/etc/ssl/vault-ca.crt', got %v", vault["caBundleFile"])
						}
						if vault["mount"] != "ssh" {
							return fmt.Errorf("expected vault mount 'ssh', got %v", vault["mount"])
						}
						if vault["role"] != "gateway" {
							return fmt.Errorf("expected vault role 'gateway', got %v", vault["role"])
						}
						if _, exists := ca["manual"]; exists {
							return fmt.Errorf("expected manual block to be absent when vault_addr is set")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayConfig_SshCAWithPrivateKey(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")
	theResource := acctests.TerraformGatewayConfig(tfName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: gatewayConfigWithPrivateKeyCA(tfName, "web", "10.0.0.1", "ubuntu", "/etc/ssh/id_ed25519"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						ssh := doc["ssh"].(map[string]any)
						ca, ok := ssh["ca"].(map[string]any)
						if !ok {
							return fmt.Errorf("expected ssh.ca block to be present")
						}
						manual, ok := ca["manual"].(map[string]any)
						if !ok {
							return fmt.Errorf("expected ssh.ca.manual block to be present")
						}
						if manual["privateKeyFile"] != "/etc/ssh/id_ed25519" {
							return fmt.Errorf("expected privateKeyFile '/etc/ssh/id_ed25519', got %v", manual["privateKeyFile"])
						}
						if _, exists := ca["vault"]; exists {
							return fmt.Errorf("expected vault block to be absent when private_key_file is set")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayConfig_CustomPortAndMetrics(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")
	theResource := acctests.TerraformGatewayConfig(tfName)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: gatewayConfigWithCustomPort(tfName, 9443, 9091, "web", "10.0.0.1", "ubuntu"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttr(theResource, attr.Port, "9443"),
					sdk.TestCheckResourceAttr(theResource, attr.MetricsPort, "9091"),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						if port, ok := doc["port"].(int); !ok || port != 9443 {
							return fmt.Errorf("expected port 9443 in YAML, got %v", doc["port"])
						}
						if mp, ok := doc["metricsPort"].(int); !ok || mp != 9091 {
							return fmt.Errorf("expected metricsPort 9091 in YAML, got %v", doc["metricsPort"])
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccTwingateGatewayConfigUpdate_ContentChanges(t *testing.T) {
	t.Parallel()

	tfName := test.TerraformRandName("test_gw_cfg")
	theResource := acctests.TerraformGatewayConfig(tfName)
	firstID := new(string)

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []sdk.TestStep{
			{
				Config: gatewayConfigWithSSHOnly(tfName, "web", "10.0.0.1", "ubuntu"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrSet(theResource, attr.ID),
					acctests.GetTwingateResourceID(theResource, &firstID),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						ssh := doc["ssh"].(map[string]any)
						upstreams := ssh["upstreams"].([]any)
						u := upstreams[0].(map[string]any)
						if u["name"] != "web" {
							return fmt.Errorf("expected upstream name 'web', got %v", u["name"])
						}
						return nil
					}),
				),
			},
			{
				Config: gatewayConfigWithSSHOnly(tfName, "db", "10.0.0.2", "postgres"),
				Check: acctests.ComposeTestCheckFunc(
					sdk.TestCheckResourceAttrWith(theResource, attr.ID, func(value string) error {
						if value == *firstID {
							return fmt.Errorf("expected ID to change after content update, got same ID %s", value)
						}
						return nil
					}),
					checkYAMLContent(theResource, func(doc map[string]any) error {
						ssh := doc["ssh"].(map[string]any)
						upstreams := ssh["upstreams"].([]any)
						u := upstreams[0].(map[string]any)
						if u["name"] != "db" {
							return fmt.Errorf("expected upstream name 'db', got %v", u["name"])
						}
						if u["user"] != "postgres" {
							return fmt.Errorf("expected upstream user 'postgres', got %v", u["user"])
						}
						return nil
					}),
				),
			},
		},
	})
}
