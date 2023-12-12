---
subcategory: "migration"
page_title: "Migrate Guide from v1 to v2"
description: |-
This document walks you through a migration process from v1 to v2
---

# Migration Guide

Migration needs to be done only for Twingate Resource objects in Terraform code. 

Changes affected schema of Resource `protocols` section, it was changed from block to object attribute.

For example, you may have Resource like this:

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

Main changes are about:
`protocols {` -> `protocols = {`
`tcp {` -> `tcp = {`
`udp {` -> `udp = {`

After changes, your Resource will look like this:

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