---
subcategory: "gcp"
page_title: "Gateway for SSH on GCE with Vault SSH CA"
description: "Deploy a Twingate Gateway for SSH on Google Compute Engine using HashiCorp Vault for SSH certificate management."
---

# Gateway for SSH on GCE with Vault SSH CA

This guide walks through deploying a Twingate Gateway for SSH on Google Compute Engine (GCE) using HashiCorp Vault as the certificate authority. The Gateway delegates SSH certificate signing to Vault via its SSH secrets engine. VMs authenticate to Vault using GCP identity tokens.

The setup requires a running Vault instance with the SSH secrets engine enabled. The included `vault/` subdirectory provides a sample Vault deployment on GCE for convenience, but you can use any Vault instance — self-hosted, [HCP Vault](https://cloud.hashicorp.com/products/vault), or otherwise.

This guide highlights the key sections. A complete, runnable example with full Terraform configurations, startup scripts, and usage instructions is available in the [`examples/gce-gateway-ssh-vault`](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/gce-gateway-ssh-vault) directory.

~> **Warning:** The Vault root token is stored unencrypted in the Terraform state. Use a [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration) to protect sensitive state data.

## Architecture

```
                                 ┌──────────────── Private VPC ─────────────────────────┐
                                 │                                                      │
SSH Client ─► Twingate Client ══════► Connector ─► Gateway (:8443) ─► SSH Server (:22)  │
                                 │                      │                    │          │
                                 │                      └──► Vault (:8200) ◄─┘          │
                                 │                                                      │
                                 └──────────────────────────────────────────────────────┘
```

The SSH client opens an SSH connection that is routed to the Gateway. The Gateway terminates the SSH connection and requests a signed SSH certificate from Vault, then opens a new upstream connection to the SSH Server.

The setup deploys four GCE VMs:

- **Vault VM** — Runs [HashiCorp Vault](https://www.vaultproject.io/) with the SSH secrets engine (and optionally PKI for TLS). Deployed in a separate Terraform step so it is running before the root module configures it.
- **Connector VM** — Bridges the private VPC to the Twingate Client via a secure tunnel.
- **Gateway VM** — Runs the [Twingate Gateway](https://github.com/Twingate/gateway) binary. Authenticates to Vault via GCP auth to sign SSH certificates on demand.
- **SSH Server VM** — A target machine that authenticates to Vault at startup to obtain a signed host certificate, and trusts the Vault SSH CA for client authentication.

## Before you begin

- A [Twingate](https://www.twingate.com) account with an [API key](https://docs.twingate.com/docs/api-overview) that has Read, Write, and Provision permissions.
- A GCP project with the Compute Engine API enabled and `gcloud` CLI authenticated.

-> **Note:** This is a two-step deploy. The `vault/` subdirectory provisions the VPC, private subnet, Cloud NAT, and Vault server. The root module reads those outputs via `terraform_remote_state` and requires an IAP tunnel to Vault.

## Setting up the providers

The root module uses three providers: Twingate, Google Cloud, and Vault. The Vault provider connects through an IAP tunnel to the Vault VM.

```terraform
terraform {
  required_providers {
    twingate = {
      source  = "Twingate/twingate"
      version = "~> 4.1"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "~> 5.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

provider "vault" {
  address         = var.vault_addr
  token           = var.vault_token
  skip_tls_verify = true
}
```

~> We recommend using [environment variables](https://www.terraform.io/language/values/variables#environment-variables) for sensitive values like the API token and Vault token.

## Configuring the Vault backends

The SSH secrets engine is the core of this setup — it signs SSH certificates on demand.

### SSH backend

```terraform
resource "vault_mount" "ssh" {
  path = "ssh"
  type = "ssh"
}

resource "vault_ssh_secret_backend_ca" "ssh" {
  backend              = vault_mount.ssh.path
  generate_signing_key = true
  key_type             = "ssh-ed25519"
}
```

Vault generates and manages the SSH CA key pair. Signing roles define what types of certificates Vault can issue. The Gateway role allows both user and host certificates:

```terraform
resource "vault_ssh_secret_backend_role" "gateway" {
  name                    = "gateway"
  backend                 = vault_mount.ssh.path

  key_type                = "ca"
  ttl                     = "720h"  # 30 days
  max_ttl                 = "8760h" # 365 days

  allow_empty_principals  = true
  allow_host_certificates = true
  allow_user_certificates = true
  allowed_domains         = "*"
  allowed_users           = "gateway"
  allowed_extensions      = "permit-X11-forwarding,permit-agent-forwarding,permit-port-forwarding,permit-pty,permit-user-rc"
}
```

A Vault policy grants the Gateway permission to call the signing endpoint:

```terraform
resource "vault_policy" "gateway" {
  name = "gateway-signing"

  policy = <<-EOT
    path "${vault_mount.ssh.path}/sign/${vault_ssh_secret_backend_role.gateway.name}" {
      capabilities = ["create", "update"]
    }
    path "${vault_mount.ssh.path}/config/ca" {
      capabilities = ["read"]
    }
  EOT
}
```

The SSH server has a similar role (`ssh-server`) but restricted to host certificates only. See the [full example](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/gce-gateway-ssh-vault) for its configuration.

### PKI backend (for Gateway TLS)

This example also uses Vault's PKI engine to issue the Gateway's TLS certificate. Any TLS certificate source will work — Vault PKI is shown here as a convenient option.

```terraform
resource "vault_mount" "pki" {
  path                      = "pki"
  type                      = "pki"
  description               = "PKI backend for X.509 certificates"
  default_lease_ttl_seconds = 3600
  max_lease_ttl_seconds     = 31536000 # 1 year
}

resource "vault_pki_secret_backend_root_cert" "root" {
  backend     = vault_mount.pki.path
  type        = "internal"
  common_name = "Demo Root CA"
  ttl         = "8760h" # 1 year
}
```

A server certificate is issued from this CA at apply time and injected into the Gateway VM. See the [full example](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/gce-gateway-ssh-vault) for the `vault_pki_secret_backend_cert` configuration.

## Vault authentication

VMs authenticate to Vault using GCP identity tokens. Each VM role has scoped policies:

```terraform
resource "vault_auth_backend" "gcp" {
  type = "gcp"
}

resource "vault_gcp_auth_backend_role" "gateway" {
  backend                = vault_auth_backend.gcp.path
  role                   = "gateway-role"
  type                   = "gce"
  token_policies         = [vault_policy.gateway.name]
  token_ttl              = 86400 # 24 hours
  token_max_ttl          = 86400
  bound_projects         = [var.project_id]
  bound_zones            = [var.zone]
  bound_service_accounts = [google_service_account.gateway.email]
}
```

## Creating the Twingate resources

Register the Vault-managed CAs with Twingate and create the Remote Network, Gateway, Connector, and SSH Resource:

```terraform
resource "twingate_ssh_certificate_authority" "vault" {
  name       = "Vault SSH CA"
  public_key = vault_ssh_secret_backend_ca.ssh.public_key
}

resource "twingate_x509_certificate_authority" "vault" {
  name        = "demo-gcp-vault-x509-ca"
  certificate = vault_pki_secret_backend_root_cert.root.certificate
}

resource "twingate_gateway" "vault" {
  remote_network_id = twingate_remote_network.main.id
  address           = "${google_compute_address.gateway.address}:${local.gateway_port}"
  x509_ca_id        = twingate_x509_certificate_authority.vault.id
  ssh_ca_id         = twingate_ssh_certificate_authority.vault.id
}

resource "twingate_ssh_resource" "ssh_server" {
  name              = "demo-gcp-vault-ssh-server"
  address           = google_compute_instance.ssh_server.network_interface[0].network_ip
  alias             = var.resource_alias != "" ? var.resource_alias : null
  remote_network_id = twingate_remote_network.main.id
  gateway_id        = twingate_gateway.vault.id

  access_group {
    group_id = data.twingate_groups.everyone.groups[0].id
  }
}
```

## Configuring the Gateway

The `ssh.ca.vault` block configures the Gateway to delegate SSH certificate signing to Vault:

- **`address`** and **`ca_bundle_file`** — The Vault server URL and the CA bundle used to verify its TLS certificate.
- **`mount`** and **`role`** — The SSH secrets engine mount path and the signing role that defines certificate parameters (TTL, extensions, etc.).
- **`auth.gcp`** — The Gateway authenticates to Vault at runtime using its GCE instance identity token. The `mount` and `role` reference the GCP auth backend configured in Vault.

```terraform
resource "twingate_gateway_config" "config" {
  port = local.gateway_port

  tls = {
    certificate_file = "/etc/gateway/tls.crt"
    private_key_file = "/etc/gateway/tls.key"
  }

  ssh = {
    gateway = {
      username = "gateway"
    }

    ca = {
      vault = {
        address        = "https://${data.terraform_remote_state.vault.outputs.vault_internal_ip}:8200"
        ca_bundle_file = "/etc/ssl/vault-ca.crt"

        mount          = vault_mount.ssh.path
        role           = vault_ssh_secret_backend_role.gateway.name

        auth = {
          gcp = {
            type  = "gce"
            mount = vault_auth_backend.gcp.path
            role  = vault_gcp_auth_backend_role.gateway.role
          }
        }
      }
    }

    resources = [twingate_ssh_resource.ssh_server]
  }
}
```

## Deploying the GCE instances

The four VMs are deployed on a dedicated VPC with Cloud NAT for internet access. The Vault VM is deployed first in the `vault/` subdirectory. The remaining VMs use startup scripts to configure themselves:

- **Connector VM** — Retrieves connector tokens from metadata and runs the Twingate connector setup script.
- **Gateway VM** — Downloads the Gateway binary, writes Vault-issued TLS certificates and the Vault CA bundle from `templatefile` variables, and starts a systemd service.
- **SSH Server VM** — Authenticates to Vault via GCP auth to obtain a signed host certificate:

```bash
#!/bin/bash
set -euo pipefail

apt-get update -qq && apt-get install -y -qq jq

# Create the gateway user
useradd -m -s /bin/bash gateway

# Authenticate to Vault via GCP auth
JWT=$(curl -sf -H "Metadata-Flavor: Google" \
  "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=vault/${vault_gcp_role}&format=full")

LOGIN_RESPONSE=$(curl -s --fail-with-body \
  --cacert /etc/ssl/vault-ca.crt \
  -X POST \
  --data "{\"role\": \"${vault_gcp_role}\", \"jwt\": \"$JWT\"}" \
  "${vault_addr}/v1/auth/${vault_gcp_mount}/login")

VAULT_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.auth.client_token')

# Sign the host key with the Vault SSH CA
HOST_PUB_KEY=$(cat /etc/ssh/ssh_host_ed25519_key.pub)

RESPONSE=$(curl -s --fail-with-body \
  --cacert /etc/ssl/vault-ca.crt \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -X POST \
  --data "{\"public_key\": \"$HOST_PUB_KEY\", \"cert_type\": \"host\", \"ttl\": \"8760h\"}" \
  "${vault_addr}/v1/${vault_mount}/sign/${vault_role}")

echo "$RESPONSE" | jq -r '.data.signed_key' > /etc/ssh/ssh_host_ed25519_key-cert.pub

# Configure sshd to trust CA and present host certificate
echo "TrustedUserCAKeys /etc/ssh/vault-ssh-ca.pub" >> /etc/ssh/sshd_config
echo "HostCertificate /etc/ssh/ssh_host_ed25519_key-cert.pub" >> /etc/ssh/sshd_config

systemctl restart sshd
```

The Gateway and SSH Server VMs use `templatefile()` to inject Vault-specific variables (CA certificate, Vault address, role names) into their startup scripts. The Gateway VM uses a reserved internal IP so its address is stable and can be registered with Twingate.
