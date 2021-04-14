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
  api_token = var.token
  network = var.network
  url = var.url
}

resource "twingate_remote_network" "test_remote_network" {
  name = "hello_from_terraform"
}

output "network_id" {
  value = twingate_remote_network.test_remote_network.id
}

