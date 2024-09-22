provider "twingate" {
#   api_token = "1234567890abcdef"
#   network   = "mynetwork"
}

resource "twingate_remote_network" "aws_network" {
  name = "aws_remote_network-2"
}

resource "twingate_group" "aws" {
  name = "aws_group"
}

data "twingate_security_policy" "test_policy" {
  name = "Test Policy"
}

data twingate_dlp_policy test {
  name = "Test"
}

resource "twingate_resource" "resource" {
  name              = "network"
  address           = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  security_policy_id = data.twingate_security_policy.test_policy.id
  dlp_policy_id = data.twingate_dlp_policy.test.id

  is_active = true
}