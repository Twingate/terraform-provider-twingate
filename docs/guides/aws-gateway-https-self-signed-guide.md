---
subcategory: "aws"
page_title: "Gateway for HTTPS on AWS with Self-Signed CA"
description: "Deploy a Twingate Gateway for an HTTPS Web App on AWS using a self-signed X.509 CA."
---

# Gateway for HTTPS on AWS with self-signed CA

This guide walks through deploying a Twingate Gateway on AWS in front of a private Web App. The Gateway terminates TLS for the Resource using a certificate signed by a self-signed X.509 CA, then forwards each request to the upstream app with identity headers injected. The example uses [httpbin](https://github.com/mccutchen/go-httpbin) as the Web App so you can see the injected headers.

This guide highlights the key sections. A complete, runnable example with full Terraform configurations, startup scripts, and usage instructions is available in the [`examples/aws-gateway-https-self-signed`](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/aws-gateway-https-self-signed) directory.

~> **Warning:** This example generates private keys and certificates that are stored unencrypted in the Terraform state. Use a [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration) to protect sensitive state data.

## Architecture

```
                                 ┌──────────────── Private VPC ──────────────────────┐
                                 │                                                   │
Browser ─► Twingate Client ════════► Connector ─► Gateway (:8443) ─► httpbin (:8080) │
                                 │                                                   │
                                 └───────────────────────────────────────────────────┘
```

The browser opens an HTTPS connection to the Gateway, using the Resource alias as the hostname. The Gateway terminates TLS using its server certificate, adds request headers, and forwards the request to httpbin over HTTP.

The setup deploys three EC2 instances:

- The **Connector instance** bridges the private VPC to the Twingate Client through a secure tunnel.
- The **Gateway instance** runs the [Twingate Gateway](https://github.com/Twingate/gateway) binary, terminates TLS, and proxies requests to the Web App.
- The **httpbin instance** runs httpbin in Docker. httpbin echoes back the request headers it receives.

## Before you begin

- A [Twingate](https://www.twingate.com) account with an [API key](https://docs.twingate.com/docs/api-overview) that has Read, Write, and Provision permissions.
- An AWS account with credentials configured (`aws configure` or environment variables).
- Terraform 1.11 or later. The X.509 CA `certificate` attribute is write-only, which requires Terraform 1.11.

-> **Note:** The example provisions its own VPC with a private subnet for the instances and a public subnet that only hosts a [NAT gateway](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-nat-gateway.html). The instances have no public IPs and reach the internet through the NAT gateway to download binaries and container images.

## Setting up the providers

The `local` provider writes the CA certificate to disk so you can add it to your trust store.

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
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
  url       = var.tg_url
}

provider "aws" {
  region = var.aws_region
}
```

~> For sensitive values like the API token, we recommend using [environment variables](https://www.terraform.io/language/values/variables#environment-variables).

## Creating the certificate authority

The Gateway serves a TLS certificate for the Resource alias. Browsers only accept it if they trust the CA that signed it, so the example creates a self-signed CA and writes its certificate to disk:

```terraform
resource "tls_private_key" "x509_ca" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "x509_ca" {
  private_key_pem = tls_private_key.x509_ca.private_key_pem

  subject {
    common_name = "AWS HTTPS Example Twingate Gateway CA"
  }

  validity_period_hours = 8760 # 1 year
  is_ca_certificate     = true

  allowed_uses = [
    "cert_signing",
    "digital_signature",
    "key_encipherment",
  ]
}

resource "local_file" "x509_ca_cert" {
  content         = tls_self_signed_cert.x509_ca.cert_pem
  filename        = "${path.module}/aws_example_https_ca.crt"
  file_permission = "0644"
}
```

The server certificate uses the Resource alias as its only DNS SAN. Use a wildcard such as `*.example.com` to cover multiple aliases without reissuing the certificate:

```terraform
resource "tls_private_key" "server" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_cert_request" "server" {
  private_key_pem = tls_private_key.server.private_key_pem

  subject {
    common_name = "Twingate Gateway"
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
```

## Creating the Twingate Resources

Register the CA with Twingate and create the Remote Network, Gateway, Connector, and Web App Resource:

```terraform
resource "twingate_remote_network" "main" {
  name = "demo-aws-https-network"
}

resource "twingate_x509_certificate_authority" "tls" {
  # CA names must be unique, so the suffix lets the old and new CA coexist during
  # rotation. Since the certificate is write-only, the name change is also what
  # tells Terraform to replace this CA when the certificate changes.
  name        = "demo-aws-https-x509-ca-${substr(sha256(tls_self_signed_cert.x509_ca.cert_pem), 0, 8)}"
  certificate = tls_self_signed_cert.x509_ca.cert_pem

  # The Gateway references this CA, so the API refuses to delete it while in use.
  # Create the new CA and repoint the Gateway before deleting the old one.
  lifecycle {
    create_before_destroy = true
  }
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.main.id
  address           = "${aws_instance.gateway.private_ip}:${local.gateway_port}"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
}

resource "twingate_connector" "main" {
  remote_network_id = twingate_remote_network.main.id
  name              = "demo-aws-https-connector"
}

resource "twingate_connector_tokens" "main" {
  connector_id = twingate_connector.main.id
}

data "twingate_groups" "everyone" {
  name = "Everyone"
}

resource "twingate_web_app_resource" "httpbin" {
  name              = "demo-aws-httpbin"
  address           = aws_instance.httpbin.private_ip
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
```

`downstream` is the side the Twingate Client connects to: HTTPS on port 443 at the Resource alias. `upstream` is the port the Gateway forwards to on the httpbin instance. `request_header_rewrites` adds headers to every request for this Resource. The example injects a Twingate-signed JWT at the Gateway level and adds user details for the httpbin Resource only.

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

`templatefile()` fills in the network and port, and the result goes to the instance's `user_data`:

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

The example deploys the three instances on a private subnet. A security group allows TCP traffic between them, so the Connector reaches the Gateway and the Gateway reaches httpbin on their private IPs. Each instance uses cloud-init (`user_data`) with `templatefile()` to inject configuration at boot:

- The **Connector instance** receives its Connector tokens through template variables and runs the Twingate Connector setup script.
- The **Gateway instance** downloads the Gateway binary, writes its TLS files and config from template variables, and starts a systemd service.
- The **httpbin instance** installs Docker and runs httpbin:

```bash
#!/bin/bash
set -eo pipefail

# Retry because outbound access through the NAT gateway can lag behind boot.
APT_OPTS="-o Acquire::Retries=10"

# Install Docker from the Ubuntu repositories
apt-get $APT_OPTS update
apt-get $APT_OPTS install -y docker.io

# --restart always brings the container back after a reboot.
docker run -d --name httpbin --restart always -p "${httpbin_port}:8080" mccutchen/go-httpbin
```

cloud-init only runs `user_data` on first boot, so each instance sets `user_data_replace_on_change`. When a certificate, Connector token, or config value changes, Terraform replaces the instance instead of updating it in place:

```terraform
resource "aws_instance" "gateway" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.main.id
  vpc_security_group_ids = [aws_security_group.internal.id]

  user_data = templatefile("${path.module}/scripts/gateway-startup.sh", {
    tls_cert       = tls_locally_signed_cert.server.cert_pem
    tls_key        = tls_private_key.server.private_key_pem
    gateway_config = local.gateway_config
  })

  user_data_replace_on_change = true

  root_block_device {
    encrypted = true
  }

  depends_on = [
    aws_route_table_association.private,
    aws_route_table_association.public,
  ]
}
```

The startup scripts download software on first boot, so each instance depends on the route table associations. This ensures the full outbound path (private subnet, NAT gateway, internet gateway) exists before the instance launches.

## Trusting the CA and testing

`terraform apply` writes the CA certificate to `aws_example_https_ca.crt`. Add it to your trust store. On macOS:

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain aws_example_https_ca.crt
```

With the Twingate Client connected, request the `/headers` endpoint:

```bash
curl https://aws-http-bin.demo.int/headers
```

httpbin returns the request headers it received, including `Authorization` and the `X-Twingate-*` headers added by the Gateway.

See the [example README](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/aws-gateway-https-self-signed) for Linux trust store commands, CA rotation, troubleshooting through the EC2 Instance Connect Endpoint, and clean up.
