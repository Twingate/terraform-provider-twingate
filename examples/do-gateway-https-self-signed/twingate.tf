# --- TLS Certificate Resources ---

resource "tls_private_key" "x509_ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "x509_ca" {
  private_key_pem = tls_private_key.x509_ca.private_key_pem

  subject {
    common_name = "DO HTTPS Example Twingate Gateway CA"
  }

  validity_period_hours = 8760 # 1 year
  is_ca_certificate     = true

  allowed_uses = [
    "digital_signature",
    "key_encipherment",
    "cert_signing",
    "crl_signing",
  ]
}

# Public CA certificate for clients to trust. Install it in your OS or browser trust store.
resource "local_file" "x509_ca_cert" {
  content         = tls_self_signed_cert.x509_ca.cert_pem
  filename        = "${path.module}/do_example_https_ca.crt"
  file_permission = "0644"
}

resource "tls_private_key" "server" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_cert_request" "server" {
  private_key_pem = tls_private_key.server.private_key_pem

  subject {
    common_name = "gateway"
  }

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
  name = "demo-test-https"
}

resource "twingate_x509_certificate_authority" "tls" {
  name        = "demo-gateway-x509-ca"
  certificate = tls_self_signed_cert.x509_ca.cert_pem
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.main.id
  address           = "${digitalocean_droplet.gateway.ipv4_address_private}:${local.gateway_port}"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
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

# SSH access to the Gateway droplet through the Connector, for troubleshooting.
resource "twingate_resource" "gateway_ssh" {
  name              = "demo-gateway-ssh"
  address           = digitalocean_droplet.gateway.ipv4_address_private
  remote_network_id = twingate_remote_network.main.id

  protocols = {
    allow_icmp = false
    tcp = {
      policy = "RESTRICTED"
      ports  = ["22"]
    }
    udp = {
      policy = "DENY_ALL"
    }
  }

  access_group {
    group_id = data.twingate_groups.everyone.groups[0].id
  }
}

resource "twingate_web_app_resource" "httpbin" {
  name              = "demo-httpbin"
  address           = digitalocean_droplet.httpbin.ipv4_address_private
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
