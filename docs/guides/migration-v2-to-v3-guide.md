---
subcategory: "migration"
page_title: "v2 to v3 Migration Guide"
description: |-
This document covers how to migrate from v2 to v3 of the Twingate Terraform provider.
---

# Migration Guide
j
This guide covers how to migrate from v2.x.x to v3.0.0 of the Twingate Terraform provider. Migration needs to be done for the following objects:
- Resources
    - `twingate_resource`

## Migrating Resources

The `access` block `twingate_resource` has been separated into two blocks: `access_group` and `access_service`. Access for Groups and Service Accounts is now specified separately. This change is primarily to enable specifying a Security Policy ID for a Group's access. 

In v2.x.x, the following was valid:

```terraform
resource "twingate_resource" "resource" {
  name = "resource"
  address = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  access {
    group_ids = [twingate_group.aws.id]
    service_account_ids = [twingate_service_account.github_actions_prod.id]
  }
}
```

From v3.0.0 and onbward, access must be specified using the `access_group` and `access_service` blocks. Further, `access_group` can only be specified for a single group and no longer uses a list of group IDs.

```terraform
resource "twingate_resource" "resource" {
  name = "resource"
  address = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  // Group access is now assigned via the `access_group` block
  // Further, security policies may now (optionally) be specified within
  // an `access_group` block.
  access_group {
      security_policy_id = twingate_security_policy.no_mfa.id
      group_id = data.twingate_groups.devops.id
  }
  
  // To assign access to multiple groups, use a `dynamic` block
  dynamic access_group {
    for_each = toset([twingate_groups.infra.id, twingate_groups.security.id])
    content {
      security_policy_id = twingate_security_policy.no_mfa.id
      group_id = access.value.key
    }
  }
  
  // Service accounts are now assigned via the `service_access` block
  // Service accounts do not use policies and, as such, one cannot be specified
  access_service {
    service_account_id = twingate_service_account.github_actions_prod.id
  }
```

