provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_group" "admins" {
  name = "admins"
}

data "twingate_users" "all" {}

locals {
  admin_users = [for user in data.twingate_users.all.users : user.id if user.is_admin == true]
}

resource "twingate_users_group_assign" "admins" {
  user_ids = local.admin_users
  group_id = twingate_group.admins.id
}