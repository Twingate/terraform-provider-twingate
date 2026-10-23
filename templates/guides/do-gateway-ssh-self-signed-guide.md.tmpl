---
subcategory: "digitalocean"
page_title: "DigitalOcean SSH Gateway with local SSH CA"
description: "Deploy a Twingate SSH Access Gateway on DigitalOcean using a local SSH CA for SSH certificate signing."
---

# DigitalOcean SSH Gateway with local SSH CA

This guide walks through deploying a Twingate SSH Access Gateway on DigitalOcean using a local SSH CA. The Gateway holds the SSH CA private key and signs SSH certificates directly, enabling certificate-based authentication without an external signing service. For simplicity, the example also uses a self-signed X.509 CA for TLS.

This guide highlights the key sections. A complete, runnable example with full Terraform configurations, startup scripts, and usage instructions is available in the [`examples/do-gateway-ssh-self-signed`](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/do-gateway-ssh-self-signed) directory.

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

The setup deploys three DigitalOcean Droplets:

- **Connector Droplet** — Bridges the private VPC to the Twingate Client via a secure tunnel.
- **Gateway Droplet** — Runs the [Twingate Gateway](https://github.com/Twingate/gateway) binary and proxies SSH connections using certificate-based authentication.
- **SSH Server Droplet** — A target machine configured to trust SSH certificates signed by the Gateway's local CA.

## Before you begin

- A [Twingate](https://www.twingate.com) account with an [API key](https://docs.twingate.com/docs/api-overview) that has Read, Write, and Provision permissions.
- A DigitalOcean account with an [API token](https://docs.digitalocean.com/reference/api/create-personal-access-token/).

-> **Note:** The example provisions its own VPC. Droplets receive public IPs by default, so no NAT configuration is needed.

## Setting up the providers

```terraform
terraform {
  required_providers {
    twingate = {
      source  = "Twingate/twingate"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
}

provider "digitalocean" {
  token = var.do_token
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

A server certificate is then signed by this CA and used by the Gateway for TLS termination. See the full example for the server certificate configuration.

### Local SSH CA

```terraform
resource "tls_private_key" "ssh_ca" {
  algorithm = "ED25519"
}
```

The Gateway uses this key to sign SSH certificates on the fly.

## Creating the Twingate resources

Register both CAs with Twingate and create the Remote Network, Gateway, Connector, SSH Resource, and a debug Resource for SSH access to the Gateway Droplet:

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
  address           = "${digitalocean_droplet.gateway.ipv4_address_private}:${local.gateway_port}"
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
  address           = digitalocean_droplet.ssh_server.ipv4_address_private
  alias             = var.resource_alias != "" ? var.resource_alias : null
  remote_network_id = twingate_remote_network.main.id
  gateway_id        = twingate_gateway.main.id

  access_group {
    group_id = data.twingate_groups.everyone.groups[0].id
  }
}

# SSH access to the Gateway Droplet through the Connector, for troubleshooting.
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
```

The Gateway is registered with its private IP, so the Connector reaches it inside the VPC. The `demo-gateway-ssh` Resource is a regular Twingate Resource routed through the Connector, so you can SSH to the Gateway Droplet for troubleshooting even though the firewall blocks inbound SSH from the internet. Once your deployment is working as expected, this can be removed.

-> **Note:** Both Resources grant access to the `Everyone` group so the example works right away. Once it's running, give each Resource the groups that should use it. For example, keep `demo-ssh-server` for the teams that need the SSH server and limit the `demo-gateway-ssh` debug Resource to administrators.

The optional `alias` field lets users connect using a friendly name (e.g., `ssh-server.int`) instead of the raw IP address. When set, the alias is also added as a DNS SAN in the server's TLS certificate so the Gateway can verify the connection.

## Configuring the gateway

The Gateway reads its settings from a YAML configuration file. The TLS and SSH CA paths are hardcoded and must match where the startup script writes those files:

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

ssh:
  gateway:
    username: gateway
    key:
      type: ed25519
    hostCertificate:
      ttl: 24h
    userCertificate:
      ttl: 5m

  ca:
    local:
      privateKeyFile: /etc/gateway/ssh-ca.key
```

`templatefile()` fills in the network and port, and the result goes to the instance's startup script:

```terraform
locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port
  })
}
```

## Deploying the Droplets

The three Droplets are deployed on a dedicated VPC. Each Droplet uses cloud-init (`user_data`) with `templatefile()` to inject configuration at boot:

- **Connector Droplet** — Receives connector tokens via template variables and runs the Twingate connector setup script.
- **Gateway Droplet** — Downloads the Gateway binary, writes TLS/SSH keys from template variables, and starts a systemd service. Twingate reaches it on its private IP.
- **SSH Server Droplet** — Creates a `gateway` user and configures `sshd` to trust the SSH CA:

```bash
#!/bin/bash
set -e

# Create the gateway user account
useradd -m -s /bin/bash gateway

# Write the SSH CA public key
cat > /etc/ssh/twingate-ca.pub <<'PUBKEY'
${ssh_ca_public_key}
PUBKEY

# Configure sshd to trust certificates signed by our CA
echo "TrustedUserCAKeys /etc/ssh/twingate-ca.pub" >> /etc/ssh/sshd_config

# Only accept the "gateway" principal, and only for the "gateway" account
mkdir -p /etc/ssh/auth_principals
echo "gateway" > /etc/ssh/auth_principals/gateway
echo "AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u" >> /etc/ssh/sshd_config

systemctl restart sshd
```

The `AuthorizedPrincipalsFile` directive restricts authentication to certificates carrying the `gateway` principal, so a CA-signed certificate can only ever log in as the `gateway` user.

Keys and certificates are injected into Droplets via `templatefile()` in `user_data`.

## Restricting inbound traffic

A DigitalOcean cloud firewall allows inbound traffic to all three Droplets only from inside the VPC, and leaves outbound connections unrestricted:

```terraform
resource "digitalocean_tag" "vpc_only" {
  name = "demo-ssh-vpc-only"
}

resource "digitalocean_firewall" "vpc_only" {
  name = "demo-ssh-vpc-only"
  tags = [digitalocean_tag.vpc_only.id]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "1-65535"
    source_addresses = [data.digitalocean_vpc.main.ip_range]
  }

  # ... matching UDP and ICMP inbound rules, and outbound rules to 0.0.0.0/0 and ::/0
}
```
