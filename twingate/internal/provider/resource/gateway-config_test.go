package resource

import (
	"context"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/provider/providerdata"
	fwattr "github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var baseConfig = providerdata.Config{Network: "mynet", URL: "twingate.com"}

var tlsObjType = types.ObjectType{
	AttrTypes: map[string]fwattr.Type{
		"certificate_file": types.StringType,
		"private_key_file": types.StringType,
	},
}

func defaultTLS() types.Object {
	return types.ObjectValueMust(tlsObjType.AttrTypes, map[string]fwattr.Value{
		"certificate_file": types.StringValue(defaultTLSCertificateFile),
		"private_key_file": types.StringValue(defaultTLSPrivateKeyFile),
	})
}

func customTLS(certFile, keyFile string) types.Object {
	return types.ObjectValueMust(tlsObjType.AttrTypes, map[string]fwattr.Value{
		"certificate_file": types.StringValue(certFile),
		"private_key_file": types.StringValue(keyFile),
	})
}

var sshGatewayObjType = types.ObjectType{
	AttrTypes: map[string]fwattr.Type{
		"username":      types.StringType,
		"key_type":      types.StringType,
		"host_cert_ttl": types.StringType,
		"user_cert_ttl": types.StringType,
	},
}

func defaultSshGateway() types.Object {
	return types.ObjectValueMust(sshGatewayObjType.AttrTypes, map[string]fwattr.Value{
		"username":      types.StringValue(defaultSSHGatewayUsername),
		"key_type":      types.StringValue(defaultSSHGatewayKeyType),
		"host_cert_ttl": types.StringValue(defaultSSHGatewayHostCertTTL),
		"user_cert_ttl": types.StringValue(defaultSSHGatewayUserCertTTL),
	})
}

func customSshGateway(username, keyType, hostCertTTL, userCertTTL string) types.Object {
	return types.ObjectValueMust(sshGatewayObjType.AttrTypes, map[string]fwattr.Value{
		"username":      types.StringValue(username),
		"key_type":      types.StringValue(keyType),
		"host_cert_ttl": types.StringValue(hostCertTTL),
		"user_cert_ttl": types.StringValue(userCertTTL),
	})
}

var sshCAObjType = types.ObjectType{
	AttrTypes: map[string]fwattr.Type{
		"vault_addr":           types.StringType,
		"private_key_file":     types.StringType,
		"vault_ca_bundle_file": types.StringType,
		"vault_mount":          types.StringType,
		"vault_role":           types.StringType,
	},
}

func defaultSshCA() types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"vault_addr":           types.StringNull(),
		"private_key_file":     types.StringNull(),
		"vault_ca_bundle_file": types.StringValue(defaultVaultCABundleFile),
		"vault_mount":          types.StringValue(defaultVaultMount),
		"vault_role":           types.StringValue(defaultVaultRole),
	})
}

func sshCAWithVault(vaultAddr string) types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"vault_addr":           types.StringValue(vaultAddr),
		"private_key_file":     types.StringNull(),
		"vault_ca_bundle_file": types.StringValue(defaultVaultCABundleFile),
		"vault_mount":          types.StringValue(defaultVaultMount),
		"vault_role":           types.StringValue(defaultVaultRole),
	})
}

func sshCAWithPrivateKey(keyFile string) types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"vault_addr":           types.StringNull(),
		"private_key_file":     types.StringValue(keyFile),
		"vault_ca_bundle_file": types.StringValue(defaultVaultCABundleFile),
		"vault_mount":          types.StringValue(defaultVaultMount),
		"vault_role":           types.StringValue(defaultVaultRole),
	})
}

var (
	sshElemType = types.ObjectType{
		AttrTypes: map[string]fwattr.Type{
			"name":     types.StringType,
			"address":  types.StringType,
			"username": types.StringType,
		},
	}

	k8sElemType = types.ObjectType{
		AttrTypes: map[string]fwattr.Type{
			"name":       types.StringType,
			"address":    types.StringType,
			"in_cluster": types.BoolType,
		},
	}
)

func makeSshList(items ...map[string]fwattr.Value) types.List {
	elems := make([]fwattr.Value, 0, len(items))
	for _, item := range items {
		elems = append(elems, types.ObjectValueMust(sshElemType.AttrTypes, item))
	}
	return types.ListValueMust(sshElemType, elems)
}

func makeK8sList(items ...map[string]fwattr.Value) types.List {
	elems := make([]fwattr.Value, 0, len(items))
	for _, item := range items {
		elems = append(elems, types.ObjectValueMust(k8sElemType.AttrTypes, item))
	}
	return types.ListValueMust(k8sElemType, elems)
}

func sshItem(name, address, username string) map[string]fwattr.Value {
	return map[string]fwattr.Value{
		"name":     types.StringValue(name),
		"address":  types.StringValue(address),
		"username": types.StringValue(username),
	}
}

func k8sItem(name, address string, inCluster bool) map[string]fwattr.Value {
	return map[string]fwattr.Value{
		"name":       types.StringValue(name),
		"address":    types.StringValue(address),
		"in_cluster": types.BoolValue(inCluster),
	}
}

func TestGatewayConfigGenerateContent(t *testing.T) {
	ctx := context.Background()

	baseSsh := makeSshList(sshItem("ssh-1", "10.0.0.1:22", "admin"))
	baseK8s := makeK8sList(k8sItem("k8s-1", "10.0.0.2:6443", true))

	cases := []struct {
		name      string
		config    providerdata.Config
		model     gatewayConfigModel
		checkYAML func(t *testing.T, doc map[string]any)
	}{
		{
			name: "vault addr set — ca uses vault block",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               sshCAWithVault("https://vault.example.com"),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				ca := ssh["ca"].(map[string]any)
				assert.Contains(t, ca, "vault", "expected vault key inside ca")
				assert.NotContains(t, ca, "manual")
				vault := ca["vault"].(map[string]any)
				assert.Equal(t, "https://vault.example.com", vault["address"])
				assert.Equal(t, "/etc/ssl/vault-ca.crt", vault["caBundleFile"])
				assert.Equal(t, "ssh", vault["mount"])
				assert.Equal(t, "gateway", vault["role"])
			},
		},
		{
			name: "private key file set — ca uses manual block",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               sshCAWithPrivateKey("/etc/ssh/id_ed25519"),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				ca := ssh["ca"].(map[string]any)
				assert.Contains(t, ca, "manual", "expected manual key inside ca")
				assert.NotContains(t, ca, "vault")
				manual := ca["manual"].(map[string]any)
				assert.Equal(t, "/etc/ssh/id_ed25519", manual["privateKeyFile"])
			},
		},
		{
			name: "both vault addr and private key empty — ca is empty",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               defaultSshCA(),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				ca := ssh["ca"]
				// ca should be nil/absent or an empty map
				if ca != nil {
					assert.Empty(t, ca.(map[string]any), "expected ca to be empty")
				}
			},
		},
		{
			name:   "twingate network is rendered",
			config: providerdata.Config{Network: "acme-corp", URL: "twingate.com"},
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               defaultSshCA(),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				twingate := doc["twingate"].(map[string]any)
				assert.Equal(t, "acme-corp", twingate["network"])
			},
		},
		{
			name: "ssh upstreams rendered correctly",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SshCA:       defaultSshCA(),
				SSHResources: makeSshList(
					sshItem("web", "192.168.1.10:22", "root"),
					sshItem("db", "192.168.1.11:22", "postgres"),
				),
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				upstreams := ssh["upstreams"].([]any)
				assert.Len(t, upstreams, 2)
				u0 := upstreams[0].(map[string]any)
				assert.Equal(t, "web", u0["name"])
				assert.Equal(t, "192.168.1.10:22", u0["address"])
				assert.Equal(t, "root", u0["user"])
				u1 := upstreams[1].(map[string]any)
				assert.Equal(t, "db", u1["name"])
				assert.Equal(t, "postgres", u1["user"])
			},
		},
		{
			name: "kubernetes upstreams rendered correctly",
			model: gatewayConfigModel{
				Port:         types.Int64Value(defaultPort),
				MetricsPort:  types.Int64Value(defaultMetricsPort),
				SshCA:        defaultSshCA(),
				SSHResources: baseSsh,
				KubernetesResources: makeK8sList(
					k8sItem("prod-cluster", "10.1.0.1:6443", true),
					k8sItem("dev-cluster", "10.2.0.1:6443", false),
				),
				TLS:        defaultTLS(),
				SshGateway: defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				k8s := doc["kubernetes"].(map[string]any)
				upstreams := k8s["upstreams"].([]any)
				assert.Len(t, upstreams, 2)
				u0 := upstreams[0].(map[string]any)
				assert.Equal(t, "prod-cluster", u0["name"])
				assert.Equal(t, "10.1.0.1:6443", u0["address"])
				assert.Equal(t, true, u0["inCluster"])
				u1 := upstreams[1].(map[string]any)
				assert.Equal(t, "dev-cluster", u1["name"])
				assert.Equal(t, false, u1["inCluster"])
			},
		},
		{
			name: "custom ssh_gateway values are rendered",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               defaultSshCA(),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          customSshGateway("ops", "rsa", "12h", "30m"),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				gw := ssh["gateway"].(map[string]any)
				assert.Equal(t, "ops", gw["username"])
				assert.Equal(t, "rsa", gw["key"].(map[string]any)["type"])
				assert.Equal(t, "12h", gw["hostCertificate"].(map[string]any)["ttl"])
				assert.Equal(t, "30m", gw["userCertificate"].(map[string]any)["ttl"])
			},
		},
		{
			name: "default ssh_gateway values are rendered",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               defaultSshCA(),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				gw := ssh["gateway"].(map[string]any)
				assert.Equal(t, defaultSSHGatewayUsername, gw["username"])
				assert.Equal(t, defaultSSHGatewayKeyType, gw["key"].(map[string]any)["type"])
				assert.Equal(t, defaultSSHGatewayHostCertTTL, gw["hostCertificate"].(map[string]any)["ttl"])
				assert.Equal(t, defaultSSHGatewayUserCertTTL, gw["userCertificate"].(map[string]any)["ttl"])
			},
		},
		{
			name: "custom port and metrics_port are rendered",
			model: gatewayConfigModel{
				Port:                types.Int64Value(9443),
				MetricsPort:         types.Int64Value(9091),
				SshCA:               defaultSshCA(),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				assert.EqualValues(t, 9443, doc["port"])
				assert.EqualValues(t, 9091, doc["metricsPort"])
			},
		},
		{
			name: "custom tls certificate and key files are rendered",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               defaultSshCA(),
				TLS:                 customTLS("/custom/tls.crt", "/custom/tls.key"),
				SshGateway:          defaultSshGateway(),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				tls := doc["tls"].(map[string]any)
				assert.Equal(t, "/custom/tls.crt", tls["certificateFile"])
				assert.Equal(t, "/custom/tls.key", tls["privateKeyFile"])
			},
		},
		{
			name: "only ssh resources — kubernetes block absent",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               defaultSshCA(),
				SSHResources:        baseSsh,
				KubernetesResources: makeK8sList(),
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				assert.Contains(t, doc, "ssh", "expected ssh block to be present")
				assert.NotContains(t, doc, "kubernetes", "expected kubernetes block to be absent")
			},
		},
		{
			name: "only kubernetes resources — ssh block absent",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               defaultSshCA(),
				SSHResources:        makeSshList(),
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				assert.Contains(t, doc, "kubernetes", "expected kubernetes block to be present")
				assert.NotContains(t, doc, "ssh", "expected ssh block to be absent")
			},
		},
		{
			name: "fixed top-level fields are present",
			model: gatewayConfigModel{
				Port:                types.Int64Value(defaultPort),
				MetricsPort:         types.Int64Value(defaultMetricsPort),
				SshCA:               defaultSshCA(),
				SSHResources:        baseSsh,
				KubernetesResources: baseK8s,
				TLS:                 defaultTLS(),
				SshGateway:          defaultSshGateway(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				assert.EqualValues(t, 8443, doc["port"])
				assert.EqualValues(t, 9090, doc["metricsPort"])
				tls := doc["tls"].(map[string]any)
				assert.Equal(t, "/etc/gateway/tls.crt", tls["certificateFile"])
				assert.Equal(t, "/etc/gateway/tls.key", tls["privateKeyFile"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.config
			if cfg == (providerdata.Config{}) {
				cfg = baseConfig
			}
			content, err := tc.model.generateContent(ctx, cfg)
			assert.NoError(t, err)
			assert.NotEmpty(t, content)

			var doc map[string]any
			err = yaml.Unmarshal([]byte(content), &doc)
			assert.NoError(t, err, "generated content must be valid YAML:\n%s", content)

			tc.checkYAML(t, doc)
		})
	}
}
