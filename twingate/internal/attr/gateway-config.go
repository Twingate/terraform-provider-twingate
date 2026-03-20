package attr

const (
	TwingateNetwork     = "twingate_network"
	VaultAddr           = "vault_addr"
	PrivateKeyFile      = "private_key_file"
	SSHResources        = "ssh_resources"
	KubernetesResources = "kubernetes_resources"
	Content             = "content"
	SshGateway          = "ssh_gateway"
	KeyType             = "key_type"
	HostCertTTL         = "host_cert_ttl"
	UserCertTTL         = "user_cert_ttl"
	Port                = "port"
	MetricsPort         = "metrics_port"
	TLS                 = "tls"
	CertificateFile     = "certificate_file"
	SshCA               = "ssh_ca"
	VaultCABundleFile   = "vault_ca_bundle_file"
	VaultMount          = "vault_mount"
	VaultRole           = "vault_role"
	VaultAuthToken      = "vault_auth_token" //nolint:gosec
)
