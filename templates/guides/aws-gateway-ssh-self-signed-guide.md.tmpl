---
subcategory: "aws"
page_title: "AWS SSH Gateway with local SSH CA"
description: "Deploy a Twingate Gateway on AWS using a local SSH CA for SSH certificate signing."
---

# AWS SSH Gateway with local SSH CA

This guide walks through deploying a Twingate Gateway on AWS using a local SSH CA. The Gateway holds the SSH CA private key and signs SSH certificates itself, so you don't need an external signing service. To keep things simple, the example also uses a self-signed X.509 CA for TLS.

This guide covers the key sections. The complete, runnable example, including startup scripts and usage instructions, is in the [`examples/aws-gateway-ssh-self-signed`](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/aws-gateway-ssh-self-signed) directory.

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

The setup deploys three EC2 instances:

- **Connector Instance** — Bridges the private VPC to the Twingate Client via a secure tunnel.
- **Gateway Instance** — Runs the [Twingate Gateway](https://github.com/Twingate/gateway) binary and proxies SSH connections using certificate-based authentication.
- **SSH Server Instance** — A target machine configured to trust SSH certificates signed by the Gateway's local CA.

## Before you begin

- A [Twingate](https://www.twingate.com) account with an [API key](https://docs.twingate.com/docs/api-overview) that has Read, Write, and Provision permissions.
- An AWS account with credentials configured (`aws configure` or environment variables).

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
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
}

provider "aws" {
  region = var.aws_region
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

A server certificate is then signed by this CA and used by the Gateway for TLS termination. See the [full example](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/aws-gateway-ssh-self-signed) for the server certificate configuration.

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
  address           = "${aws_instance.gateway.private_ip}:${local.gateway_port}"
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
  address           = aws_instance.ssh_server.private_ip
  remote_network_id = twingate_remote_network.main.id
  gateway_id        = twingate_gateway.main.id
  alias             = var.resource_alias != "" ? var.resource_alias : null

  access_group {
    group_id = data.twingate_groups.everyone.groups[0].id
  }
}
```

The Gateway is registered with its private IP, so the Connector reaches it inside the VPC.

-> **Note:** The SSH Resource grants access to the `Everyone` group so the example works right away. Once it's running, replace it with the groups that should reach the SSH server.

The optional `alias` field lets users connect using a friendly name (e.g., `ssh-server.int`) instead of the raw IP address. When set, the alias is also added as a DNS SAN in the server's TLS certificate so the Gateway can verify the connection.

## Configuring the Gateway

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

## Deploying the EC2 instances

All three instances run in a private subnet with no public IPs. Each one gets its keys, certificates and configuration at boot from cloud-init (`user_data`), rendered with `templatefile()`:

- **Connector Instance** — Receives connector tokens via template variables and runs the Twingate connector setup script.
- **Gateway Instance** — Downloads the Gateway binary, writes TLS/SSH keys from template variables, and starts a systemd service. Twingate reaches it on its private IP.
- **SSH Server Instance** — Creates a `gateway` user and configures `sshd` to trust the SSH CA:

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

The Connector and Gateway startup scripts download software at boot. Both instances `depends_on` the private and public route associations, which in turn depend on the NAT gateway. The downloads also retry for a short time, because outbound access through the NAT gateway can lag behind the instance boot.

## Restricting inbound traffic

The instances have no public IPs and no ports open to the internet. A security group allows inbound TCP traffic between the three instances, plus SSH (port 22) from an EC2 Instance Connect Endpoint for troubleshooting. Outbound traffic is unrestricted. All access is authorized through Twingate or AWS IAM.

```terraform
resource "aws_security_group" "internal" {
  name   = "demo-internal"
  vpc_id = aws_vpc.main.id

  ingress {
    protocol  = "tcp"
    from_port = 0
    to_port   = 65535
    self      = true
  }

  ingress {
    protocol        = "tcp"
    from_port       = 22
    to_port         = 22
    security_groups = [aws_security_group.eic.id]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

## Troubleshooting

The instances have no public IPs, so the example adds an EC2 Instance Connect Endpoint to reach them for troubleshooting:

```terraform
resource "aws_security_group" "eic" {
  name   = "demo-eic-endpoint"
  vpc_id = aws_vpc.main.id

  egress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = [aws_subnet.main.cidr_block]
  }
}

resource "aws_ec2_instance_connect_endpoint" "main" {
  subnet_id          = aws_subnet.main.id
  security_group_ids = [aws_security_group.eic.id]
}
```

The EC2 console and `aws ec2-instance-connect ssh` both connect through the endpoint. Each connection is authorized by IAM: the caller needs `ec2-instance-connect:OpenTunnel` on the endpoint and `ec2-instance-connect:SendSSHPublicKey` on the instance. AWS then pushes a short-lived SSH key to the instance, so no key pair is needed.

See the [example README](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/aws-gateway-ssh-self-signed#troubleshooting) for the connection commands and how to view the Gateway logs.
