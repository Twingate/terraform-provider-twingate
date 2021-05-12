provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

variable "twingate_network" {
  description = "Name of twingate remote network"
  default     = "my_test_network"
}

resource "twingate_remote_network" "network" {
  name = var.twingate_network
}

resource "twingate_resource" "resource" {
  name = "network"
  address = "internal.int"
  remote_network_id = twingate_remote_network.network.id
  group_ids = ["group1"]
  protocols {
    allow_icmp = true
    tcp  {
      policy = "RESTRICTED"
      ports = ["80", "82-83"]
    }
    udp {
      policy = "ALLOW_ALL"
    }
  }
}