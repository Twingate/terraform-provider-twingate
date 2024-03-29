---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "twingate_groups Data Source - terraform-provider-twingate"
subcategory: ""
description: |-
  Groups are how users are authorized to access Resources. For more information, see Twingate's documentation https://docs.twingate.com/docs/groups.
---

# twingate_groups (Data Source)

Groups are how users are authorized to access Resources. For more information, see Twingate's [documentation](https://docs.twingate.com/docs/groups).

## Example Usage

```terraform
data "twingate_groups" "foo" {
  name = "<your group's name>"
  #  name_regexp = "<regular expression of group name>"
  #  name_contains = "<a string in the group name>"
  #  name_exclude = "<your group's name to exclude>"
  #  name_prefix = "<prefix of resource name>"
  #  name_suffix = "<suffix of resource name>"
}

# Group names are not constrained to be unique within Twingate,
# so it is possible that this data source will return multiple list items.
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `is_active` (Boolean) Returns only Groups matching the specified state.
- `name` (String) Returns only groups that exactly match this name. If no options are passed it will return all resources. Only one option can be used at a time.
- `name_contains` (String) Match when the value exist in the name of the group.
- `name_exclude` (String) Match when the exact value does not exist in the name of the group.
- `name_prefix` (String) The name of the group must start with the value.
- `name_regexp` (String) The regular expression match of the name of the group.
- `name_suffix` (String) The name of the group must end with the value.
- `types` (Set of String) Returns groups that match a list of types. valid types: `MANUAL`, `SYNCED`, `SYSTEM`.

### Read-Only

- `groups` (Attributes List) List of Groups (see [below for nested schema](#nestedatt--groups))
- `id` (String) The ID of this resource.

<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Read-Only:

- `id` (String) The ID of the Group
- `is_active` (Boolean) Indicates if the Group is active
- `name` (String) The name of the Group
- `security_policy_id` (String) The Security Policy assigned to the Group.
- `type` (String) The type of the Group
