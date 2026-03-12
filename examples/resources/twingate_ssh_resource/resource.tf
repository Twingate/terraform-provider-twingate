provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_remote_network" "prod" {
  name = "Production Network"
}

resource "twingate_x509_certificate_authority" "tls" {
  name        = "My TLS CA"
  certificate = file("ca.pem")
}

resource "twingate_ssh_certificate_authority" "ssh" {
  name       = "My SSH CA"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIr+Aj3O8csUrFRNWS7wViafil3rMlZ0glQ/OZ0CjCti"
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.prod.id
  address           = "10.0.0.1:8001"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
  ssh_ca_id         = twingate_ssh_certificate_authority.ssh.id
}

# Kubernetes resource accessed via in-cluster DNS
resource "twingate_kubernetes_resource" "prod_cluster" {
  name       = "Production K8s"
  remote_network_id = twingate_remote_network.prod.id
  gateway_id = twingate_gateway.main.id
  address    = "kubernetes.default.svc.cluster.local"
  in_cluster = true
}

resource "twingate_ssh_resource" "bastion" {
  name       = "SSH Bastion"
  gateway_id = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
  address    = "10.128.0.105"
  username   = "ubuntu"
}

resource "twingate_ssh_resource" "bastion2" {
  name       = "SSH Bastion 2"
  gateway_id = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
  address    = "10.128.0.106"
  username   = "ubuntu-2"
}

resource "local_file" "config" {
  content  = templatefile("${path.module}/config-template.yaml", {
    ssh_resources = [twingate_ssh_resource.bastion, twingate_ssh_resource.bastion2]
    kubernetes_resources = [twingate_kubernetes_resource.prod_cluster]
  })
  filename = "${path.module}/generated/config.yaml"
}
```

### config-template.yaml
```yaml
twingate:
network: "my-network"
host: "twingate.com"

port: 8443
metricsPort: 9090

tls:
certificateFile: "/etc/gateway/tls.crt"
privateKeyFile: "/etc/gateway/tls.key"

kubernetes:
upstreams: %{ for item in kubernetes_resources }
- name: ${item.name}
address: ${item.address}
inCluster: ${item.in_cluster}%{ endfor }

ssh:
gateway:
username: "gateway"
key:
type: "ed25519"
hostCertificate:
ttl: "24h"
userCertificate:
ttl: "5m"

ca:
vault:
address: "<https://vault.example.com:8200>"
caBundleFile: "/etc/ssl/vault-ca.crt"
auth:
token: "<vault-token>"
mount: "ssh"
role: "gateway"

upstreams: %{ for item in ssh_resources }
- name: ${item.name}
address: ${item.address}
user: ${item.username}%{ endfor }
