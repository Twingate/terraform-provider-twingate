provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_remote_network" "gcp" {
  name = "GCP Network"
}

resource "twingate_x509_certificate_authority" "tls" {
  name        = "My TLS CA"
  certificate = file("ca.pem")
}

resource "twingate_ssh_certificate_authority" "ssh" {
  name       = "My SSH CA"
  public_key = trimspace(file("~/.ssh/id_ed25519.pub"))
}

# Gateway with both X.509 and SSH CAs
resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.gcp.id
  address           = "10.0.0.1:8001"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
  ssh_ca_id         = twingate_ssh_certificate_authority.ssh.id
}

# Gateway with only X.509 CA (ssh_ca_id is optional)
resource "twingate_gateway" "minimal" {
  remote_network_id = twingate_remote_network.gcp.id
  address           = "10.0.0.2:9001"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
}
