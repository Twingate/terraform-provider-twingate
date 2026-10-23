# --- TLS Certificate Resources ---

# Generate a self-signed X.509 CA for the gateway
resource "tls_private_key" "x509_ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "x509_ca" {
  private_key_pem = tls_private_key.x509_ca.private_key_pem

  subject {
    common_name = "GCE HTTPS Example Twingate Gateway CA"
  }

  validity_period_hours = 8760 # 1 year
  is_ca_certificate     = true

  allowed_uses = [
    "cert_signing",
    "digital_signature",
    "key_encipherment",
  ]
}

# Public CA certificate for clients to trust. Install it in your OS or browser trust store.
resource "local_file" "x509_ca_cert" {
  content         = tls_self_signed_cert.x509_ca.cert_pem
  filename        = "${path.module}/gce_example_https_ca.crt"
  file_permission = "0644"
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

  # Replace this with a wildcard to cover multiple aliases
  # For example: ["*.example.com"]
  dns_names = [var.resource_alias]
}

resource "tls_locally_signed_cert" "server" {
  cert_request_pem   = tls_cert_request.server.cert_request_pem
  ca_private_key_pem = tls_private_key.x509_ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.x509_ca.cert_pem

  validity_period_hours = 8760 # 365 days

  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "server_auth",
  ]
}

# --- Twingate Resources ---

resource "twingate_remote_network" "main" {
  name = "demo-gcp-https-network"
}

resource "twingate_x509_certificate_authority" "tls" {
  name        = "demo-gcp-https-x509-ca"
  certificate = tls_self_signed_cert.x509_ca.cert_pem
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.main.id
  address           = "${google_compute_address.gateway.address}:${local.gateway_port}"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
}

resource "twingate_connector" "main" {
  remote_network_id = twingate_remote_network.main.id
  name              = "demo-gcp-https-connector"
}

resource "twingate_connector_tokens" "main" {
  connector_id = twingate_connector.main.id
}

# Everyone has access so the example works immediately. Replace this with the groups
# that should reach the Resource.
data "twingate_groups" "everyone" {
  name = "Everyone"
}

resource "twingate_web_app_resource" "httpbin" {
  name              = "demo-gcp-httpbin"
  address           = google_compute_instance.httpbin.network_interface[0].network_ip
  remote_network_id = twingate_remote_network.main.id
  gateway_id        = twingate_gateway.main.id
  alias             = var.resource_alias

  downstream = {
    port     = 443
    tls_mode = "TLS13"
  }

  upstream = {
    port = local.httpbin_port
  }

  request_header_rewrites = {
    "X-Twingate-User"    = "{{username}}",
    "X-Twingate-Groups"  = "{{groups}}",
    "X-Twingate-City"    = "{{clientGeoCity}}",
    "X-Twingate-Country" = "{{clientGeoCountry}}",
  }

  access_group {
    group_id = data.twingate_groups.everyone.groups[0].id
  }
}
