# Generate a self-signed X.509 CA for the gateway
resource "tls_private_key" "x509_ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "x509_ca" {
  private_key_pem = tls_private_key.x509_ca.private_key_pem

  subject {
    common_name = "Twingate Gateway CA"
  }

  validity_period_hours = 8760 # 1 year
  is_ca_certificate     = true

  allowed_uses = [
    "cert_signing",
    "digital_signature",
    "key_encipherment",
  ]
}

# Generate a server TLS certificate signed by the CA
resource "tls_private_key" "server" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_cert_request" "server" {
  private_key_pem = tls_private_key.server.private_key_pem

  subject {
    common_name = "Twingate Gateway"
  }

  dns_names    = var.resource_alias != "" ? [var.resource_alias] : []
  ip_addresses = [google_compute_instance.ssh_server.network_interface[0].network_ip]
}

resource "tls_locally_signed_cert" "server" {
  cert_request_pem   = tls_cert_request.server.cert_request_pem
  ca_private_key_pem = tls_private_key.x509_ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.x509_ca.cert_pem

  validity_period_hours = 720 # 30 days

  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "server_auth",
  ]
}

# Generate a fixed SSH CA key pair for the gateway
resource "tls_private_key" "ssh_ca" {
  algorithm = "ED25519"
}

resource "twingate_remote_network" "main" {
  name = "demo-test-ssh"
}

resource "twingate_ssh_certificate_authority" "ssh" {
  name       = "demo-ssh-ca"
  public_key = tls_private_key.ssh_ca.public_key_openssh
}

resource "twingate_x509_certificate_authority" "tls" {
  name        = "demo-gateway-x509-ca"
  certificate = tls_self_signed_cert.x509_ca.cert_pem
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.main.id
  address           = "${google_compute_address.gateway.address}:${local.gateway_port}"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
  ssh_ca_id         = twingate_ssh_certificate_authority.ssh.id
}

resource "twingate_connector" "main" {
  remote_network_id = twingate_remote_network.main.id
  name              = "demo-connector"
}

resource "twingate_connector_tokens" "main" {
  connector_id = twingate_connector.main.id
}

data "twingate_groups" "everyone" {
  name = "Everyone"
}

resource "twingate_ssh_resource" "ssh_server" {
  name              = "demo-ssh-server"
  address           = google_compute_instance.ssh_server.network_interface[0].network_ip
  alias             = var.resource_alias != "" ? var.resource_alias : null
  remote_network_id = twingate_remote_network.main.id
  gateway_id        = twingate_gateway.main.id

  access_group {
    group_id = data.twingate_groups.everyone.groups[0].id
  }
}
