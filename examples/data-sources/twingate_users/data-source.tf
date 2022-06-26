provider "twingate" {}

data "twingate_users" "all" {}

output "my_users" {
  value = data.twingate_users.all
}