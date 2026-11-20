---
subcategory: "azure"
page_title: "Azure SSH Gateway with local SSH CA"
description: "Deploy a Twingate Gateway on Azure using a local SSH CA for SSH certificate signing."
---

# Azure SSH Gateway with local SSH CA

This guide walks through deploying a Twingate Gateway on Azure using a local SSH CA. The Gateway holds the SSH CA private key and signs SSH certificates itself, so you don't need an external signing service. To keep things simple, the example also uses a self-signed X.509 CA for TLS.

This guide covers the key sections. The complete, runnable example, including startup scripts and usage instructions, is in the [`examples/azure-gateway-ssh-self-signed`](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/azure-gateway-ssh-self-signed) directory.

~> **Warning:** This example generates private keys and certificates that are stored unencrypted in the Terraform state. Use a [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration) to protect sensitive state data.

## Architecture

```
                                 ┌──────────────── Private VNet ────────────────────────┐
                                 │                                                      │
SSH Client ─► Twingate Client ══════► Connector ─► Gateway (:8443) ─► SSH Server (:22)  │
                                 │                                                      │
                                 └──────────────────────────────────────────────────────┘
```

The SSH client opens an SSH connection that is routed to the Gateway. The Gateway terminates the SSH connection and opens a new upstream connection to the SSH Server using a signed SSH certificate.

The setup deploys three Azure VMs:

- **Connector VM** — Bridges the private VNet to the Twingate Client via a secure tunnel.
- **Gateway VM** — Runs the [Twingate Gateway](https://github.com/Twingate/gateway) binary and proxies SSH connections using certificate-based authentication.
- **SSH Server VM** — A target machine configured to trust SSH certificates signed by the Gateway's local CA.

## Before you begin

- A [Twingate](https://www.twingate.com) account with an [API key](https://docs.twingate.com/docs/api-overview) that has Read, Write, and Provision permissions.
- An Azure subscription and the `az` CLI authenticated (`az login`).
- An Azure region that offers the [Azure Bastion Developer SKU](https://learn.microsoft.com/en-us/azure/bastion/quickstart-developer). The example defaults to `eastus2`.

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
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
}

# Credentials come from the Azure CLI (`az login`).
provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
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

A server certificate is then signed by this CA and used by the Gateway for TLS termination. See the [full example](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/azure-gateway-ssh-self-signed) for the server certificate configuration.

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
  name = "demo-azure-ssh-network"
}

resource "twingate_ssh_certificate_authority" "ssh" {
  name       = "demo-azure-ssh-ca"
  public_key = tls_private_key.ssh_ca.public_key_openssh
}

resource "twingate_x509_certificate_authority" "tls" {
  # CA names must be unique, so the suffix lets the old and new CA coexist during
  # rotation. Since the certificate is write-only, the name change is also what
  # tells Terraform to replace this CA when the certificate changes.
  name        = "demo-azure-x509-ca-${substr(sha256(tls_self_signed_cert.x509_ca.cert_pem), 0, 8)}"
  certificate = tls_self_signed_cert.x509_ca.cert_pem

  # The Gateway references this CA, so the API refuses to delete it while in use.
  # Create the new CA and repoint the Gateway before deleting the old one.
  lifecycle {
    create_before_destroy = true
  }
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.main.id
  address           = "${azurerm_network_interface.gateway.private_ip_address}:${local.gateway_port}"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
  ssh_ca_id         = twingate_ssh_certificate_authority.ssh.id
}

resource "twingate_connector" "main" {
  remote_network_id = twingate_remote_network.main.id
  name              = "demo-azure-connector"
}

resource "twingate_connector_tokens" "main" {
  connector_id = twingate_connector.main.id
}

data "twingate_groups" "everyone" {
  name = "Everyone"
}

resource "twingate_ssh_resource" "ssh_server" {
  name              = "demo-azure-ssh-server"
  address           = azurerm_network_interface.ssh_server.private_ip_address
  alias             = var.resource_alias != "" ? var.resource_alias : null
  remote_network_id = twingate_remote_network.main.id
  gateway_id        = twingate_gateway.main.id

  access_group {
    group_id = data.twingate_groups.everyone.groups[0].id
  }
}
```

The Gateway is registered with its private IP, so the Connector reaches it inside the VNet. The Gateway and SSH server addresses come from their network interfaces, which exist before the VMs.

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

`templatefile()` fills in the network and port, and the result goes to the VM's startup script:

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

## Deploying the Azure VMs

All three VMs run in a single subnet with no public IPs. Each one gets its keys, certificates and configuration at boot from cloud-init (`custom_data`), rendered with `templatefile()`. Changing `custom_data` forces a new VM, so a new certificate, key or config replaces the VM instead of updating it in place.

- **Connector VM** — Receives connector tokens via template variables and runs the Twingate connector setup script.
- **Gateway VM** — Downloads the Gateway binary, writes TLS/SSH keys from template variables, and starts a systemd service. Twingate reaches it on its private IP.
- **SSH Server VM** — Creates a `gateway` user and configures `sshd` to trust the SSH CA:

```bash
#!/bin/bash
set -e

# Create the gateway user account
useradd -m -s /bin/bash gateway

# Write the SSH CA public key
cat > /etc/ssh/twingate-ca.pub <<'PUBKEY'
${ssh_ca_public_key}
PUBKEY

# Only accept the "gateway" principal, and only for the "gateway" account
mkdir -p /etc/ssh/auth_principals
echo "gateway" > /etc/ssh/auth_principals/gateway

# Configure sshd to trust certificates signed by our CA
cat > /etc/ssh/sshd_config.d/00-twingate.conf <<'SSHD'
TrustedUserCAKeys /etc/ssh/twingate-ca.pub
AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u
SSHD

systemctl restart ssh
```

The `AuthorizedPrincipalsFile` directive restricts authentication to certificates carrying the `gateway` principal, so a CA-signed certificate can only ever log in as the `gateway` user.

New Azure subnets have no default outbound access, so the example attaches a NAT gateway to the subnet:

```terraform
resource "azurerm_public_ip" "nat" {
  name                = "demo-nat-ip"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Standard"
  allocation_method   = "Static"
}

resource "azurerm_nat_gateway" "main" {
  name                = "demo-nat"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku_name            = "Standard"
}

resource "azurerm_nat_gateway_public_ip_association" "main" {
  nat_gateway_id       = azurerm_nat_gateway.main.id
  public_ip_address_id = azurerm_public_ip.nat.id
}

resource "azurerm_subnet_nat_gateway_association" "main" {
  subnet_id      = azurerm_subnet.main.id
  nat_gateway_id = azurerm_nat_gateway.main.id
}
```

The Connector and Gateway startup scripts download software at boot. Both VMs `depends_on` the two NAT gateway associations. The downloads also retry for a short time, because outbound access through the NAT gateway can lag behind the VM boot.

## Restricting inbound traffic

The VMs have no public IPs and no ports open to the internet. The network security group relies on Azure's default rules: `AllowVnetInBound` allows traffic between the three VMs, and `DenyAllInBound` blocks inbound traffic from the internet. One extra rule allows SSH (port 22) from Azure Bastion for troubleshooting. Outbound traffic is unrestricted. All access is authorized through Twingate or Azure.

```terraform
resource "azurerm_network_security_group" "main" {
  name                = "demo-nsg"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  # The Bastion Developer SKU connects from 168.63.129.16.
  security_rule {
    name                       = "AllowBastionSsh"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "168.63.129.16"
    destination_address_prefix = "*"
  }
}

resource "azurerm_subnet_network_security_group_association" "main" {
  subnet_id                 = azurerm_subnet.main.id
  network_security_group_id = azurerm_network_security_group.main.id
}
```

## Troubleshooting

The VMs have no public IPs, so the example adds an Azure Bastion host on the free Developer SKU to reach them for troubleshooting. The Developer SKU needs no dedicated subnet or public IP:

```terraform
resource "azurerm_bastion_host" "main" {
  name                = "demo-bastion"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Developer"
  virtual_network_id  = azurerm_virtual_network.main.id
}
```

Azure requires an admin user on every Linux VM. The example generates an SSH key pair for the `azureadmin` user and exposes the private key as the sensitive `vm_admin_private_key` output. Bastion connections are made from the Azure portal (**Connect > Bastion**) and authorized by your Azure role on the VM. The Developer SKU doesn't support `az network bastion ssh`.

See the [example README](https://github.com/Twingate/terraform-provider-twingate/tree/main/examples/azure-gateway-ssh-self-signed#troubleshooting) for the connection steps and how to view the Gateway logs.
