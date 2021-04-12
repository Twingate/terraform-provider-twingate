variable "token" {
  type = string
  sensitive = true
}
variable "network" {
  type = string
}

variable "url" {
  type = string
}

provider "twingate" {
  token = var.token
  network = var.network
  url = var.url
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