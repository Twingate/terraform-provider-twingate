provider "twingate" {
#   # api_token = "1234567890abcdef"
#   # network   = "mynetwork"
#
  default_tags = {
    tags = {
      managed-by = "Terraform"
      owner = "DevOps"
      cloud = "AWS"
    }
  }

  cache = {
    groups_enabled = false
    resources_filter = {
      name_prefix = "tf-acc-"
    #   tags = {
    #     cloud = "AWS"
    #   }
    }
  }
}

data "twingate_resources" "foo" {
  # name = "<your resource's name>"
   name_regexp = "tf-acc-"
  #  name_contains = "<a string in the resource name>"
  #  name_exclude = "<your resource's name to exclude>"
  #  name_prefix = "<prefix of resource name>"
  #  name_suffix = "<suffix of resource name>"

  # tags = {
  #   environment = "dev"
  # }
}

output "resources_count" {
  value = length(data.twingate_resources.foo.resources)
}

resource "twingate_remote_network" "aws_network" {
  name = "aws_remote_network-monday"
}

# resource "twingate_group" "aws" {
#   name = "aws_group"
# }

# data "twingate_group" "security" {
#   id = "securityGroupID"
# }
#
# data "twingate_groups" "devops" {
#   name_contains = "DevOps"
# }
#
# data "twingate_groups" "sre" {
#   name_contains = "SRE"
# }
#
# resource "twingate_service_account" "github_actions_prod" {
#   name = "Github Actions PROD"
# }
#
# data "twingate_security_policy" "test_policy" {
#   name = "Test Policy"
# }

resource "twingate_resource" "resource" {
  name              = "network-monday"
  address           = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  # security_policy_id = data.twingate_security_policy.test_policy.id
  # approval_mode      = "MANUAL"
  # usage_based_autolock_duration_days = 15

  # protocols = {
  #   allow_icmp = true
  #   tcp = {
  #     policy = "RESTRICTED"
  #     ports  = ["80", "82-83"]
  #   }
  #   udp = {
  #     policy = "ALLOW_ALL"
  #   }
  # }
  #
  # // Adding a single group via `access_group`
  # access_group {
  #   group_id                           = twingate_group.aws.id
  #   security_policy_id                 = data.twingate_security_policy.test_policy.id
  #   usage_based_autolock_duration_days = 30
  #   approval_mode                      = "AUTOMATIC"
  # }
  #
  # // Adding multiple groups by individual ID
  # dynamic "access_group" {
  #   for_each = toset([twingate_group.aws.id, data.twingate_group.security.id])
  #   content {
  #     group_id                           = access_group.value
  #     security_policy_id                 = data.twingate_security_policy.test_policy.id
  #     usage_based_autolock_duration_days = 30
  #   }
  # }
  #
  # // Adding multiple groups from twingate_groups data sources
  # dynamic "access_group" {
  #   for_each = setunion(
  #     data.twingate_groups.devops.groups[*].id,
  #     data.twingate_groups.sre.groups[*].id,
  #     // Single IDs can be added by wrapping them in a set
  #     toset([data.twingate_group.security.id])
  #   )
  #   content {
  #     group_id                           = access_group.value
  #     security_policy_id                 = data.twingate_security_policy.test_policy.id
  #     usage_based_autolock_duration_days = 30
  #
  #   }
  # }
  #
  # // Service account access is specified similarly
  # // A `for_each` block may be used like above to assign access to multiple
  # // service accounts in a single configuration block.
  # access_service {
  #   content {
  #     service_account_id = twingate_service_account.github_actions_prod.id
  #   }
  # }
  #
  # is_active = true
  #
  # tags = {
  #   environment = "dev"
  #   owner       = "devops"
  #   project     = "api"
  # }
}

