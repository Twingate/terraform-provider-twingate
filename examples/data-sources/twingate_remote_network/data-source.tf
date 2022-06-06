provider "twingate" {}

resource "twingate_remote_network" "net1" {
  name = "net1"
}

data "twingate_remote_network" "net2" {
  id = twingate_remote_network.net1.id
}

output "my_network_id" {
  value = data.twingate_remote_network.net2.id
}

output "my_network1_name" {
  value = data.twingate_remote_network.net2.name
}
