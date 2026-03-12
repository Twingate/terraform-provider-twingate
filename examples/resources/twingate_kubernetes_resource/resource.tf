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
  public_key = trimspace(file("~/.ssh/id_ed25519.pub"))
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.prod.id
  address           = "10.0.0.1:8001"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
  ssh_ca_id         = twingate_ssh_certificate_authority.ssh.id
}

# Kubernetes resource accessed via in-cluster DNS
resource "twingate_kubernetes_resource" "prod_cluster" {
  name              = "Production K8s"
  address           = "kubernetes.default.svc.cluster.local"
  gateway_id        = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
  in_cluster        = true
}

# Kubernetes resource accessed via external address
resource "twingate_kubernetes_resource" "external_cluster" {
  name              = "External K8s"
  address           = "k8s-api.example.com"
  gateway_id        = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
}