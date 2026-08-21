---
subcategory: "migration"
page_title: "v4 to v5 Migration Guide"
description: "This document covers how to migrate from v4 to v5 of the Twingate Terraform provider."
---

# Migration Guide

This guide covers how to migrate from v4.x.x to v5.0.0 of the Twingate Terraform provider. Migration needs to be done for the following objects:
- Resources
    - `twingate_ssh_resource`
    - `twingate_kubernetes_resource`
    - `twingate_gateway_config`

## Migrating SSH and Kubernetes Resources

The `protocols` attribute has been removed from the `twingate_ssh_resource` and `twingate_kubernetes_resource` resources. Port restrictions do not apply to SSH and Kubernetes Resources, so the attribute never had any effect.

In v4.x.x, the following was valid:

```terraform
resource "twingate_ssh_resource" "ssh_server" {
  name              = "SSH Server"
  address           = "10.128.0.105"
  remote_network_id = twingate_remote_network.prod.id
  gateway_id        = twingate_gateway.main.id
  username          = "ubuntu"

  protocols = {
    allow_icmp = true
    tcp = {
      policy = "RESTRICTED"
      ports  = ["22"]
    }
    udp = {
      policy = "DENY_ALL"
    }
  }
}
```

From v5.0.0 and onward, the `protocols` attribute must be removed:

```terraform
resource "twingate_ssh_resource" "ssh_server" {
  name              = "SSH Server"
  address           = "10.128.0.105"
  remote_network_id = twingate_remote_network.prod.id
  gateway_id        = twingate_gateway.main.id
  username          = "ubuntu"
}
```

The same applies to the `twingate_kubernetes_resource` resource. A configuration that still sets `protocols` on either resource fails with `Unsupported argument`.

## Migrating Gateway Configuration

The `twingate_gateway_config` resource has been removed. The Gateway no longer reads upstreams from its configuration file, so the resource only ever rendered a static YAML document and needed nothing from Terraform state to do it. The built-in [`templatefile()`](https://developer.hashicorp.com/terraform/language/functions/templatefile) function covers the same ground.

In v4.x.x, the following was valid:

```terraform
resource "twingate_gateway_config" "config" {
  port = local.gateway_port

  tls = {
    certificate_file = "/etc/gateway/tls.crt"
    private_key_file = "/etc/gateway/tls.key"
  }

  ssh = {
    gateway = { username = "gateway" }
    ca      = { private_key_file = "/etc/gateway/ssh-ca.key" }

    resources = [
      twingate_ssh_resource.ssh_server,
    ]
  }
}
```

From v5.0.0 and onward, the configuration must be rendered from a template file. Store the document as `config.yaml.tftpl`:

```yaml
twingate:
  network: ${twingate_network}
  host: ${twingate_host}

port: ${port}
metricsPort: 9090

tls:
  certificateFile: /etc/gateway/tls.crt
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
    manual:
      privateKeyFile: /etc/gateway/ssh-ca.key
```

Then render it with `templatefile()`:

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

Any reference to `twingate_gateway_config.config.content` becomes `local.gateway_config`. Since `replace_triggered_by` only accepts resource references, a `lifecycle` block that pointed at `twingate_gateway_config.config` needs the rendered config wrapped in a `terraform_data` resource:

```terraform
resource "terraform_data" "gateway_config" {
  input = local.gateway_config
}
```

Finally, drop the resource from state:

```bash
terraform state rm twingate_gateway_config.<name>
```

The resource was purely local and never created a remote object, so nothing else is required. The SSH Gateway guides show the complete configuration.
