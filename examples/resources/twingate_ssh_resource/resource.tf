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
  address           = "10.0.0.1:8443"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
  ssh_ca_id         = twingate_ssh_certificate_authority.ssh.id
}

resource "twingate_ssh_resource" "ssh_server" {
  name       = "SSH Server"
  gateway_id = twingate_gateway.main.id
  alias      = "test.int"
  remote_network_id = twingate_remote_network.prod.id
  address    = "10.128.0.105"
  username   = "ubuntu"
}
