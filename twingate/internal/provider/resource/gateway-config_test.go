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

var vaultObjType = types.ObjectType{
	AttrTypes: map[string]fwattr.Type{
		"address":        types.StringType,
		"ca_bundle_file": types.StringType,
		"mount":          types.StringType,
		"role":           types.StringType,
		"auth":           authObjType,
	},
}

var gcpObjType = types.ObjectType{
	AttrTypes: map[string]fwattr.Type{
		"role":                  types.StringType,
		"type":                  types.StringType,
		"mount":                 types.StringType,
		"service_account_email": types.StringType,
	},
}

var authObjType = types.ObjectType{
	AttrTypes: map[string]fwattr.Type{
		"token": types.StringType,
		"gcp":   gcpObjType,
	},
}

var sshCAObjType = types.ObjectType{
	AttrTypes: map[string]fwattr.Type{
		"private_key_file": types.StringType,
		"vault":            vaultObjType,
	},
}

func defaultVaultObj() types.Object {
	return types.ObjectValueMust(vaultObjType.AttrTypes, map[string]fwattr.Value{
		"address":        types.StringNull(),
		"ca_bundle_file": types.StringValue(defaultVaultCABundleFile),
		"mount":          types.StringValue(defaultVaultMount),
		"role":           types.StringValue(defaultVaultRole),
		"auth":           defaultAuthObj(),
	})
}

func defaultGCPObj() types.Object {
	return types.ObjectValueMust(gcpObjType.AttrTypes, map[string]fwattr.Value{
		"role":                  types.StringNull(),
		"type":                  types.StringNull(),
		"mount":                 types.StringValue(defaultGCPMount),
		"service_account_email": types.StringNull(),
	})
}

func defaultAuthObj() types.Object {
	return types.ObjectValueMust(authObjType.AttrTypes, map[string]fwattr.Value{
		"token": types.StringNull(),
		"gcp":   defaultGCPObj(),
	})
}

func defaultSshCA() types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"private_key_file": types.StringNull(),
		"vault":            defaultVaultObj(),
	})
}

func sshCAWithVault(vaultAddr string) types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"private_key_file": types.StringNull(),
		"vault": types.ObjectValueMust(vaultObjType.AttrTypes, map[string]fwattr.Value{
			"address":        types.StringValue(vaultAddr),
			"ca_bundle_file": types.StringValue(defaultVaultCABundleFile),
			"mount":          types.StringValue(defaultVaultMount),
			"role":           types.StringValue(defaultVaultRole),
			"auth":           defaultAuthObj(),
		}),
	})
}

func sshCAWithVaultAndToken(vaultAddr, authToken string) types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"private_key_file": types.StringNull(),
		"vault": types.ObjectValueMust(vaultObjType.AttrTypes, map[string]fwattr.Value{
			"address":        types.StringValue(vaultAddr),
			"ca_bundle_file": types.StringValue(defaultVaultCABundleFile),
			"mount":          types.StringValue(defaultVaultMount),
			"role":           types.StringValue(defaultVaultRole),
			"auth": types.ObjectValueMust(authObjType.AttrTypes, map[string]fwattr.Value{
				"token": types.StringValue(authToken),
				"gcp":   defaultGCPObj(),
			}),
		}),
	})
}

func sshCAWithVaultAndGCP(vaultAddr, gcpRole, gcpType string) types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"private_key_file": types.StringNull(),
		"vault": types.ObjectValueMust(vaultObjType.AttrTypes, map[string]fwattr.Value{
			"address":        types.StringValue(vaultAddr),
			"ca_bundle_file": types.StringValue(defaultVaultCABundleFile),
			"mount":          types.StringValue(defaultVaultMount),
			"role":           types.StringValue(defaultVaultRole),
			"auth": types.ObjectValueMust(authObjType.AttrTypes, map[string]fwattr.Value{
				"token": types.StringNull(),
				"gcp": types.ObjectValueMust(gcpObjType.AttrTypes, map[string]fwattr.Value{
					"role":                  types.StringValue(gcpRole),
					"type":                  types.StringValue(gcpType),
					"mount":                 types.StringValue(defaultGCPMount),
					"service_account_email": types.StringNull(),
				}),
			}),
		}),
	})
}

func sshCAWithVaultAndGCPFull(vaultAddr, gcpRole, gcpType, gcpMount, serviceAccountEmail string) types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"private_key_file": types.StringNull(),
		"vault": types.ObjectValueMust(vaultObjType.AttrTypes, map[string]fwattr.Value{
			"address":        types.StringValue(vaultAddr),
			"ca_bundle_file": types.StringValue(defaultVaultCABundleFile),
			"mount":          types.StringValue(defaultVaultMount),
			"role":           types.StringValue(defaultVaultRole),
			"auth": types.ObjectValueMust(authObjType.AttrTypes, map[string]fwattr.Value{
				"token": types.StringNull(),
				"gcp": types.ObjectValueMust(gcpObjType.AttrTypes, map[string]fwattr.Value{
					"role":                  types.StringValue(gcpRole),
					"type":                  types.StringValue(gcpType),
					"mount":                 types.StringValue(gcpMount),
					"service_account_email": types.StringValue(serviceAccountEmail),
				}),
			}),
		}),
	})
}

func sshCAWithPrivateKey(keyFile string) types.Object {
	return types.ObjectValueMust(sshCAObjType.AttrTypes, map[string]fwattr.Value{
		"private_key_file": types.StringValue(keyFile),
		"vault":            defaultVaultObj(),
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

	sshObjType = types.ObjectType{
		AttrTypes: map[string]fwattr.Type{
			"gateway":   sshGatewayObjType,
			"ca":        sshCAObjType,
			"resources": types.ListType{ElemType: sshElemType},
		},
	}

	k8sObjType = types.ObjectType{
		AttrTypes: map[string]fwattr.Type{
			"resources": types.ListType{ElemType: k8sElemType},
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

func makeSshObj(gateway, ca types.Object, resources types.List) types.Object {
	return types.ObjectValueMust(sshObjType.AttrTypes, map[string]fwattr.Value{
		"gateway":   gateway,
		"ca":        ca,
		"resources": resources,
	})
}

func defaultSshObj(resources types.List) types.Object {
	return makeSshObj(defaultSshGateway(), defaultSshCA(), resources)
}

func makeK8sObj(resources types.List) types.Object {
	return types.ObjectValueMust(k8sObjType.AttrTypes, map[string]fwattr.Value{
		"resources": resources,
	})
}

func defaultK8sObj() types.Object {
	return makeK8sObj(makeK8sList())
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

	baseSsh := makeSshList(sshItem("ssh-1", "10.0.0.1", "admin"))
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
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         makeSshObj(defaultSshGateway(), sshCAWithVault("https://vault.example.com"), baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
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
			name: "vault addr and auth token set — ca vault block includes auth",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         makeSshObj(defaultSshGateway(), sshCAWithVaultAndToken("https://vault.example.com", "s.mytoken"), baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				ca := ssh["ca"].(map[string]any)
				vault := ca["vault"].(map[string]any)
				auth := vault["auth"].(map[string]any)
				assert.Equal(t, "s.mytoken", auth["token"])
			},
		},
		{
			name: "vault addr and gcp auth set — ca vault block includes gcp auth",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         makeSshObj(defaultSshGateway(), sshCAWithVaultAndGCP("https://vault.example.com", "vm-role", "gce"), baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				ca := ssh["ca"].(map[string]any)
				vault := ca["vault"].(map[string]any)
				auth := vault["auth"].(map[string]any)
				gcp := auth["gcp"].(map[string]any)
				assert.Equal(t, "vm-role", gcp["role"])
				assert.Equal(t, "gce", gcp["type"])
			},
		},
		{
			name: "gcp auth with custom mount rendered",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         makeSshObj(defaultSshGateway(), sshCAWithVaultAndGCPFull("https://vault.example.com", "vm-role", "gce", "custom-gcp", ""), baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				gcp := doc["ssh"].(map[string]any)["ca"].(map[string]any)["vault"].(map[string]any)["auth"].(map[string]any)["gcp"].(map[string]any)
				assert.Equal(t, "custom-gcp", gcp["mount"])
				assert.NotContains(t, gcp, "serviceAccountEmail")
			},
		},
		{
			name: "gcp auth with service account email rendered for iam type",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         makeSshObj(defaultSshGateway(), sshCAWithVaultAndGCPFull("https://vault.example.com", "vm-role", "iam", defaultGCPMount, "sa@project.iam.gserviceaccount.com"), baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				gcp := doc["ssh"].(map[string]any)["ca"].(map[string]any)["vault"].(map[string]any)["auth"].(map[string]any)["gcp"].(map[string]any)
				assert.Equal(t, defaultGCPMount, gcp["mount"])
				assert.Equal(t, "sa@project.iam.gserviceaccount.com", gcp["serviceAccountEmail"])
			},
		},
		{
			name: "vault addr set without auth token — no auth block",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         makeSshObj(defaultSshGateway(), sshCAWithVault("https://vault.example.com"), baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				ssh := doc["ssh"].(map[string]any)
				ca := ssh["ca"].(map[string]any)
				vault := ca["vault"].(map[string]any)
				assert.NotContains(t, vault, "auth")
			},
		},
		{
			name: "private key file set — ca uses manual block",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         makeSshObj(defaultSshGateway(), sshCAWithPrivateKey("/etc/ssh/id_ed25519"), baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
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
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         defaultSshObj(baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
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
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         defaultSshObj(baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
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
				SSH: defaultSshObj(makeSshList(
					sshItem("web", "192.168.1.10", "root"),
					sshItem("db", "192.168.1.11", "postgres"),
				)),
				Kubernetes: makeK8sObj(baseK8s),
				TLS:        defaultTLS(),
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
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         defaultSshObj(baseSsh),
				Kubernetes: makeK8sObj(makeK8sList(
					k8sItem("prod-cluster", "10.1.0.1:6443", true),
					k8sItem("dev-cluster", "10.2.0.1:6443", false),
				)),
				TLS: defaultTLS(),
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
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         makeSshObj(customSshGateway("ops", "rsa", "12h", "30m"), defaultSshCA(), baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
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
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         defaultSshObj(baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
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
				Port:        types.Int64Value(9443),
				MetricsPort: types.Int64Value(9091),
				SSH:         defaultSshObj(baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				assert.EqualValues(t, 9443, doc["port"])
				assert.EqualValues(t, 9091, doc["metricsPort"])
			},
		},
		{
			name: "custom tls certificate and key files are rendered",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         defaultSshObj(baseSsh),
				TLS:         customTLS("/custom/tls.crt", "/custom/tls.key"),
				Kubernetes:  makeK8sObj(baseK8s),
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
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         defaultSshObj(baseSsh),
				Kubernetes:  makeK8sObj(makeK8sList()),
				TLS:         defaultTLS(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				assert.Contains(t, doc, "ssh", "expected ssh block to be present")
				assert.NotContains(t, doc, "kubernetes", "expected kubernetes block to be absent")
			},
		},
		{
			name: "only kubernetes resources — ssh block absent",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         defaultSshObj(makeSshList()),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
			},
			checkYAML: func(t *testing.T, doc map[string]any) {
				assert.Contains(t, doc, "kubernetes", "expected kubernetes block to be present")
				assert.NotContains(t, doc, "ssh", "expected ssh block to be absent")
			},
		},
		{
			name: "fixed top-level fields are present",
			model: gatewayConfigModel{
				Port:        types.Int64Value(defaultPort),
				MetricsPort: types.Int64Value(defaultMetricsPort),
				SSH:         defaultSshObj(baseSsh),
				Kubernetes:  makeK8sObj(baseK8s),
				TLS:         defaultTLS(),
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
