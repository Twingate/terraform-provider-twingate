provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_remote_network" "aws_network" {
  name = "aws_remote_network"
}

resource "twingate_group" "aws" {
  name = "aws_group"
}

data "twingate_security_policy" "test_policy" {
  name = "Test Policy"
}

resource "twingate_resource" "resource" {
  name              = "network"
  address           = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  is_authoritative = false
}

resource "twingate_resource_access" "aws" {
  resource_id        = twingate_resource.resource.id
  group_id           = twingate_group.aws.id
  security_policy_id = data.twingate_security_policy.test_policy.id

  access_policy {
    mode          = "AUTO_LOCK"
    approval_mode = "AUTOMATIC"
    duration      = "30d"
  }
}
