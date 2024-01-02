---
subcategory: "migration"
page_title: "v1 to v2 Migration Guide"
description: |-
This document covers how to migrate from v1 to v2 of the Twingate Terraform provider.
---

# Migration Guide
j
This guide covers how to migrate from v1 to v2 of the Twingate Terraform provider. Migration needs to be done for the following objects:
- Resources
    - `twingate_resource`
- Data sources
    - `twingate_user`
    - `twingate_users`

## Migrating Resources

The `protocols` attribute in the `twingate_resource` Resource has been changed from a block to an object.

In v1, the following was valid:

```terraform
resource "twingate_resource" "resource" {
  name = "resource"
  address = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  protocols {
    allow_icmp = true
    tcp {
      policy = "RESTRICTED"
      ports = ["80", "82-83"]
    }
    udp {
      policy = "ALLOW_ALL"
    }
  }
}
```

The `protocols`, `tcp` and `udp` attributes were blocks and not objects. In v2, these are now objects:

```
protocols {   ->   protocols = {
tcp {         ->   tcp = {
udp {         ->   udp = {
```

In v2, the above resource needs to be rewritten like this:

```terraform
resource "twingate_resource" "resource" {
  name = "resource"
  address = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  protocols = {
    allow_icmp = true
    tcp = {
      policy = "RESTRICTED"
      ports = ["80", "82-83"]
    }
    udp = {
      policy = "ALLOW_ALL"
    }
  }
}
```

## Migrating data sources

The attribute `is_admin` has been removed from the `twingate_user` and `twingate_users` data sources. Similar information is now available via the [`role` attribute](https://registry.terraform.io/providers/Twingate/twingate/latest/docs/data-sources/users#role).
