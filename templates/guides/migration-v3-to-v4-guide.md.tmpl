---
subcategory: "migration"
page_title: "v3 to v4 Migration Guide"
description: "This document covers how to migrate from v3 to v4 of the Twingate Terraform provider."
---

# Migration Guide

This guide covers how to migrate from v3.x.x to v4.0.0 of the Twingate Terraform provider. Migration needs to be done for the following objects:
- Resources
    - `twingate_resource`

## Migrating Resources

The `approval_mode` and `usage_based_autolock_duration_days` attributes in the `twingate_resource` top-level block and within the `access_group` block are deprecated.

Access controls should now be configured using the new `access_policy` block. This block is available both at the resource level (setting the default policy for the resource) and inside `access_group` blocks (setting specific policies for a group).

### Access Policy Configuration

The `access_policy` block supports the following arguments:

*   `mode` (Required) - The access mode. Valid values are `MANUAL`, `AUTO_LOCK`, or `ACCESS_REQUEST`.
*   `duration` (Optional) - The duration of the access (e.g., "1h", "48h").
    *   Required if `mode` is `AUTO_LOCK` (minimum "24h").
    *   Optional if `mode` is `ACCESS_REQUEST` (minimum "1h").
*   `approval_mode` (Optional) - The approval mode. Valid values are `MANUAL` or `AUTOMATIC`.
    *   Required if `mode` is `AUTO_LOCK` or `ACCESS_REQUEST`.


In v3.x.x, the following was valid:

```terraform
resource "twingate_resource" "resource" {
  name = "resource"
  address = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  approval_mode = "MANUAL"
  usage_based_autolock_duration_days = 2
  
  access_group {
    group_id = data.twingate_groups.devops.id
    security_policy_id = twingate_security_policy.no_mfa.id
    approval_mode = "MANUAL"
    usage_based_autolock_duration_days = 2
  }
}
```

From v4.0.0 and onward, access must be specified using the `access_policy` block (top-level block and within the `access_group` block).

```terraform
resource "twingate_resource" "resource" {
  name              = "resource"
  address           = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  access_policy {
    mode     = "MANUAL"
    duration = "48h"
  }

  access_group {
    group_id           = data.twingate_groups.devops.id
    security_policy_id = twingate_security_policy.no_mfa.id
    access_policy {
      mode          = "AUTO_LOCK"
      approval_mode = "MANUAL"
      duration      = "48h"
    }
  }
}
```
