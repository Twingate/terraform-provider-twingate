# --- TLS Certificate Resources ---

locals {
  # The Twingate Client connects with the alias, so it must be covered by the wildcard
  # certificate. No public DNS record is needed for it, only the ACME challenge record.
  resource_alias = "httpbin.${var.domain}"
}

# ACME account with Let's Encrypt
resource "tls_private_key" "acme_account" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "acme_registration" "main" {
  account_key_pem = tls_private_key.acme_account.private_key_pem
  email_address   = var.acme_email
}

# Wildcard certificate issued by Let's Encrypt. The DNS-01 challenge writes a TXT record
# to the Cloud DNS zone for var.domain using your Application Default Credentials.
resource "acme_certificate" "gateway" {
  account_key_pem = acme_registration.main.account_key_pem
  common_name     = "*.${var.domain}"

  key_type = "2048"
  # Let's Encrypt offers more than one chain (for example, YR1 -> Root YR, cross-signed by
  # ISRG Root X1). Ask for the one that ends at the root registered with Twingate below.
  preferred_chain = "ISRG Root X1"

  # Reissued on the next `terraform apply` once fewer than 30 days remain (certificates last 90 days)
  min_days_remaining = 30

  dns_challenge {
    provider = "gcloud"

    config = {
      GCE_PROJECT = var.project_id
    }
  }
}

# --- Twingate Resources ---

resource "twingate_remote_network" "main" {
  name = "demo-gcp-https-public-ca-network"
}

# The Let's Encrypt root the Gateway's certificate chains to (ISRG Root X1). It has to be
# the root, since Let's Encrypt picks the intermediate per order.
resource "twingate_x509_certificate_authority" "letsencrypt" {
  name        = "demo-gcp-https-public-ca-letsencrypt-root"
  certificate = file("${path.module}/isrg-root-x1.pem")
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.main.id
  address           = "${google_compute_address.gateway.address}:${local.gateway_port}"
  x509_ca_id        = twingate_x509_certificate_authority.letsencrypt.id
}

resource "twingate_connector" "main" {
  remote_network_id = twingate_remote_network.main.id
  name              = "demo-gcp-https-public-ca-connector"
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
  name              = "demo-gcp-public-ca-httpbin"
  address           = google_compute_instance.httpbin.network_interface[0].network_ip
  remote_network_id = twingate_remote_network.main.id
  gateway_id        = twingate_gateway.main.id
  alias             = local.resource_alias

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
