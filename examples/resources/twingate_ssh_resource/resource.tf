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

resource "twingate_ssh_resource" "bastion" {
  name              = "Production Bastion"
  address           = "bastion.internal.example.com"
  gateway_id        = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
}