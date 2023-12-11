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

resource "twingate_group" "devops" {
  name = "DevOps"
}

resource "twingate_service_account" "github_actions_prod" {
  name = "Github Actions PROD"
}

data "twingate_security_policy" "mfa" {
  name = "Default Policy"
}

data "twingate_security_policy" "test_policy" {
  name = "Test Policy"
}

resource "twingate_resource" "resource" {
  name = "network"
  address = "internal.int"
  remote_network_id = twingate_remote_network.aws_network.id

  security_policy_id = data.twingate_security_policy.test_policy.id

  protocols = {
    allow_icmp = true
    tcp = {
      policy = "RESTRICTED"
      ports = ["80", "82-83"]
    }
    udp = {
      policy = "ALLOW_ALL"
    }
  }

  dynamic "access" {
    for_each = [twingate_group.devops.id, twingate_group.aws.id]
    content {
      security_policy_id = data.twingate_security_policy.mfa.id
      group_id = access.value
    }
  }

  access {
    service_account_ids = [twingate_service_account.github_actions_prod.id]
  }
}

