---
subcategory: "gcp"
page_title: "Gateway for HTTPS on GCE with Let's Encrypt"
description: "Deploy a Twingate Gateway for an HTTPS Web App on Google Compute Engine using a publicly trusted certificate from Let's Encrypt."
---

# Gateway for HTTPS on GCE with Let's Encrypt

This guide walks through deploying a Twingate Gateway on Google Compute Engine (GCE) in front of a private Web App. The Gateway terminates TLS for the Resource using a wildcard certificate issued by [Let's Encrypt](https://letsencrypt.org/), then forwards each request to the upstream app with identity headers injected. The example uses [httpbin](https://github.com/mccutchen/go-httpbin) as the Web App so you can see the injected headers.

Terraform requests the certificate with the [ACME provider](https://registry.terraform.io/providers/vancluever/acme/latest/docs) and proves domain ownership with a DNS-01 challenge against Cloud DNS. The certificate chains to a publicly trusted root, so clients do not need to install a CA certificate.

This guide highlights the key sections. A complete, runnable example with full Terraform configurations, startup scripts, and usage instructions is available in the [`examples/gce-gateway-https-public-ca`](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/gce-gateway-https-public-ca) directory.

~> **Warning:** This example stores the Gateway's private key and the ACME account key unencrypted in the Terraform state. Use a [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration) to protect sensitive state data.

## Architecture

```
                                 ┌──────────────── Private VPC ──────────────────────┐
                                 │                                                   │
Browser ─► Twingate Client ════════► Connector ─► Gateway (:8443) ─► httpbin (:8080) │
                                 │                                                   │
                                 └───────────────────────────────────────────────────┘
```

The browser opens an HTTPS connection to the Gateway, using the Resource alias as the hostname. The Gateway terminates TLS using its Let's Encrypt certificate, adds request headers, and forwards the request to httpbin over HTTP.

The setup deploys three GCE VMs:

- The **Connector VM** bridges the private VPC to the Twingate Client through a secure tunnel.
- The **Gateway VM** runs the [Twingate Gateway](https://github.com/Twingate/gateway) binary, terminates TLS, and proxies requests to the Web App.
- The **httpbin VM** runs httpbin in Docker. httpbin echoes back the request headers it receives.

## Before you begin

- Terraform 1.11 or later. The `certificate` argument of `twingate_x509_certificate_authority` is [write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).
- A [Twingate](https://www.twingate.com) account with an [API key](https://docs.twingate.com/docs/api-overview) that has Read, Write, and Provision permissions.
- A GCP project with the Compute Engine and Cloud DNS APIs enabled.
- A domain you control, delegated to a **public** Cloud DNS zone in that project.
- The `gcloud` CLI, authenticated with Application Default Credentials so the Google and ACME providers can reach GCP:

  ```bash
  gcloud auth login
  gcloud auth application-default login
  ```

  The ACME provider uses these credentials to create the challenge TXT record, so they need permission to change records in the zone (for example, `roles/dns.admin`).

-> **Note:** The example provisions its own VPC, private subnet, and [Cloud NAT](https://cloud.google.com/nat/docs/overview) so VMs can download binaries and container images from the internet.

## Setting up the providers

The `acme` provider requests the certificate from Let's Encrypt. The `tls` provider generates the ACME account key.

```terraform
terraform {
  required_providers {
    twingate = {
      source  = "Twingate/twingate"
      version = "~> 5.2"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    acme = {
      source  = "vancluever/acme"
      version = "~> 2.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
  url       = var.tg_url
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

provider "acme" {
  server_url = "https://acme-v02.api.letsencrypt.org/directory"
}
```

~> For sensitive values like the API token, we recommend using [environment variables](https://www.terraform.io/language/values/variables#environment-variables).

## Requesting the certificate

The certificate covers `*.<domain>`, and the Resource alias is `httpbin.<domain>`. The alias does not need a public DNS record because the Twingate Client resolves it. The only record written to the zone is the temporary `_acme-challenge` TXT record.

```terraform
locals {
  resource_alias = "httpbin.${var.domain}"
}
```

Register an ACME account, then request the wildcard certificate with a DNS-01 challenge against Cloud DNS:

```terraform
resource "tls_private_key" "acme_account" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "acme_registration" "main" {
  account_key_pem = tls_private_key.acme_account.private_key_pem
  email_address   = var.acme_email
}

resource "acme_certificate" "gateway" {
  account_key_pem = acme_registration.main.account_key_pem
  common_name     = "*.${var.domain}"

  key_type = "2048"
  # Let's Encrypt offers more than one chain. Ask for the one that ends at the
  # root registered with Twingate below.
  preferred_chain = "ISRG Root X1"

  # Reissued on the next `terraform apply` once fewer than 30 days remain
  min_days_remaining = 30

  dns_challenge {
    provider = "gcloud"

    config = {
      GCE_PROJECT = var.project_id
    }
  }
}
```

Let's Encrypt certificates are valid for 90 days, and nothing renews them in the background. Run `terraform apply` within 30 days of expiry to issue a new certificate. The apply also replaces the Gateway VM.

## Creating the Twingate Resources

Register the Let's Encrypt root with Twingate and create the Remote Network, Gateway, Connector, and Web App Resource. Twingate needs the root rather than an intermediate, because Let's Encrypt can choose a different intermediate for each certificate it issues. The example ships the root as `isrg-root-x1.pem`.

```terraform
resource "twingate_remote_network" "main" {
  name = "demo-gcp-https-public-ca-network"
}

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
```

`downstream` is the side the Twingate Client connects to: HTTPS on port 443 at the Resource alias. `upstream` is the port the Gateway forwards to on the httpbin VM. `request_header_rewrites` adds headers to every request for this Resource. The example injects a Twingate-signed JWT at the Gateway level and adds user details for the httpbin Resource only.

The `Everyone` group gives all users access so the example works immediately. Replace it with the groups that should reach the Resource.

## Configuring the Gateway

The Gateway reads its settings from a YAML configuration file. The TLS paths are hardcoded and must match where the startup script writes those files:

```yaml
twingate:
  network: ${twingate_network}
  host: ${twingate_host}

port: ${port}
metricsPort: 9090

tls:
  certificates:
    files:
      - certificateFile: /etc/gateway/tls.crt
        privateKeyFile: /etc/gateway/tls.key

webApp:
  requestHeaders:
    # Injects into every request
    Authorization: "Bearer {{jwt}}"
```

`webApp.requestHeaders` adds headers to every Web App request the Gateway handles, in addition to the per-Resource `request_header_rewrites`. Here it forwards a JWT the upstream app can verify.

`templatefile()` fills in the network and port. The certificate file contains the leaf followed by the Let's Encrypt intermediate (`issuer_pem`), so clients can build the chain to the root. The certificate, key, and config all go to the instance metadata:

```terraform
locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port
  })

  gateway_metadata = {
    tls-cert       = "${acme_certificate.gateway.certificate_pem}${acme_certificate.gateway.issuer_pem}"
    tls-key        = acme_certificate.gateway.private_key_pem
    gateway-config = local.gateway_config
  }
}
```

## Deploying the GCE instances

The example deploys the three VMs on a dedicated VPC, with Cloud NAT for internet access. Each VM depends on the NAT so its startup script can reach the internet on first boot:

- The **Connector VM** reads its Connector tokens from metadata and runs the Twingate Connector setup script.
- The **Gateway VM** downloads the Gateway binary, writes its TLS files and config from instance metadata, and starts a systemd service.
- The **httpbin VM** installs Docker and runs httpbin.

The Gateway VM uses a reserved internal IP, so its address stays stable and can be registered with Twingate. When the certificate or config changes, for example on renewal, `replace_triggered_by` replaces the VM so the startup script runs again with the new values:

```terraform
resource "terraform_data" "gateway_metadata" {
  input = local.gateway_metadata
}

resource "google_compute_instance" "gateway" {
  # ...

  metadata                = local.gateway_metadata
  metadata_startup_script = file("${path.module}/scripts/gateway-startup.sh")

  depends_on = [google_compute_router_nat.main]

  lifecycle {
    replace_triggered_by = [terraform_data.gateway_metadata]
  }
}
```

## Testing

With the Twingate Client connected, request the `/headers` endpoint:

```bash
curl https://httpbin.<domain>/headers
```

httpbin returns the request headers it received, including `Authorization` and the `X-Twingate-*` headers added by the Gateway.

See the [example README](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/gce-gateway-https-public-ca) for renewal and troubleshooting.
