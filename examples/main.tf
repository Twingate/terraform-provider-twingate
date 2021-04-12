variable "token" {
  type = string
  sensitive = true
}
variable "tenant" {
  type = string
}

provider "twingate" {
  token = var.token
  tenant = var.tenant
  url = "dev.opstg.com"
}

data "twingate_group" "test" {
  name = "Employees"
}

resource "twingate_remote_network" "test_remote_network" {
  name = "hello_from_terraform"
  is_active = true
}

output "test" {
  value = data.twingate_group.test.name
}

output "is_active" {
  value = twingate_remote_network.test_remote_network.is_active
}