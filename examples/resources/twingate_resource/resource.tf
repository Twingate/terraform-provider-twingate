provider "twingate" {
#  api_token = "1234567890abcdef"
#  network   = "mynetwork"
}

#resource "twingate_remote_network" "aws_network" {
#  name = "aws_remote_network"
#}
#
#resource "twingate_group" "aws" {
#  name = "aws_group"
#}

resource "twingate_resource" "resource" {
  name = "network"
  address = "internal.int"
  remote_network_id = "UmVtb3RlTmV0d29yazo0MDEzOQ=="
  group_ids = ["R3JvdXA6MzUxNjU="]
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