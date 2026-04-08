---
subcategory: "gcp"
page_title: "GCE SSH Gateway with Local SSH CA"
description: "Deploy a Twingate SSH Access Gateway on Google Compute Engine using a local SSH CA for SSH certificate signing."
---

# GCE SSH Gateway with local SSH CA

This guide walks through deploying a Twingate SSH Access Gateway on Google Compute Engine (GCE) using a local SSH CA. The Gateway holds the SSH CA private key and signs SSH certificates directly, enabling certificate-based authentication without an external signing service. For simplicity, the example also uses a self-signed X.509 CA for TLS.

This guide highlights the key sections. A complete, runnable example with full Terraform configurations, startup scripts, and usage instructions is available in the [`examples/gce-gateway-ssh-self-signed`](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/gce-gateway-ssh-self-signed) directory.

~> **Warning:** This example generates private keys and certificates that are stored unencrypted in the Terraform state. Use a [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration) to protect sensitive state data.

## Architecture

```
                                 ┌──────────────── Private VPC ─────────────────────────┐
                                 │                                                      │
SSH Client ─► Twingate Client ══════► Connector ─► Gateway (:8443) ─► SSH Server (:22)  │
                                 │                                                      │
                                 └──────────────────────────────────────────────────────┘
```

The SSH client opens an SSH connection that is routed to the Gateway. The Gateway terminates the SSH connection and opens a new upstream connection to the SSH Server using a signed SSH certificate.

The setup deploys three GCE VMs:

- **Connector VM** — Bridges the private VPC to the Twingate Client via a secure tunnel.
- **Gateway VM** — Runs the [Twingate Gateway](https://github.com/Twingate/gateway) binary and proxies SSH connections using certificate-based authentication.
- **SSH Server VM** — A target machine configured to trust SSH certificates signed by the Gateway's local CA.

## Before you begin

- A [Twingate](https://www.twingate.com) account with an [API key](https://docs.twingate.com/docs/api-overview) that has Read, Write, and Provision permissions.
- A GCP project with the Compute Engine API enabled and `gcloud` CLI authenticated.

-> **Note:** The example provisions its own VPC, private subnet, and [Cloud NAT](https://cloud.google.com/nat/docs/overview) so VMs can download binaries from the internet.

## Setting up the providers

```terraform
terraform {
  required_providers {
    twingate = {
      source  = "Twingate/twingate"
      version = "~> 4.1"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    google = {
      source  = "hashicorp/google"
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
```

~> We recommend using [environment variables](https://www.terraform.io/language/values/variables#environment-variables) for sensitive values like the API token.

## Creating the certificate authorities

The Gateway needs two CAs: a local SSH CA for signing SSH certificates, and an X.509 CA for TLS encryption (self-signed here for simplicity).

### X.509 CA (for TLS)

```terraform
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
```

A server certificate is then signed by this CA and used by the Gateway for TLS termination. See the [full example](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/gce-gateway-ssh-self-signed) for the server certificate configuration.

### Local SSH CA

```terraform
resource "tls_private_key" "ssh_ca" {
  algorithm = "ED25519"
}
```

The Gateway uses this key to sign SSH certificates on the fly.

## Creating the Twingate resources

Register both CAs with Twingate and create the Remote Network, Gateway, Connector, and SSH Resource:

```terraform
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
```

The optional `alias` field lets users connect using a friendly name (e.g., `ssh-server.int`) instead of the raw IP address. When set, the alias is also added as a DNS SAN in the server's TLS certificate so the Gateway can verify the connection.

## Configuring the Gateway

The `twingate_gateway_config` resource generates the Gateway's configuration file. It specifies the TLS certificate paths and SSH CA key path:

```terraform
resource "twingate_gateway_config" "config" {
  port = local.gateway_port

  tls = {
    certificate_file = "/opt/gateway/tls.crt"
    private_key_file = "/opt/gateway/tls.key"
  }

  ssh = {
    gateway = {
      username = "gateway"
    }

    ca = {
      private_key_file = "/opt/gateway/ssh-ca.key"
    }

    resources = [twingate_ssh_resource.ssh_server]
  }
}
```

## Deploying the GCE instances

The three VMs are deployed on a dedicated VPC with Cloud NAT for internet access. Each VM uses a startup script to configure itself:

- **Connector VM** — Retrieves connector tokens from metadata and runs the Twingate connector setup script.
- **Gateway VM** — Downloads the Gateway binary, writes TLS/SSH keys from instance metadata, and starts a systemd service.
- **SSH Server VM** — Creates a `gateway` user and configures `sshd` to trust the SSH CA:

```bash
#!/bin/bash
set -e

# Create the gateway user account
useradd -m -s /bin/bash gateway

# Get the SSH CA public key from instance metadata
CA_KEY=$(curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/ssh-ca-public-key)

# Configure sshd to trust certificates signed by our CA
echo "$CA_KEY" > /etc/ssh/twingate-ca.pub
echo "TrustedUserCAKeys /etc/ssh/twingate-ca.pub" >> /etc/ssh/sshd_config

systemctl restart sshd
```

The Gateway VM uses a reserved internal IP so its address is stable and can be registered with Twingate. Keys and certificates are passed to VMs via GCE instance metadata.
