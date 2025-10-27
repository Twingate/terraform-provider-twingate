provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_remote_network" "aws_network" {
  name = "test-aws_remote_network"
}

resource "twingate_connector" "aws_connector" {
  remote_network_id = twingate_remote_network.aws_network.id
}

ephemeral "twingate_connector_tokens" "aws_connector_tokens" {
  connector_id = twingate_connector.aws_connector.id
}

resource "twingate_connector_tokens" "aws_connector_tokens" {
  connector_id = twingate_connector.aws_connector.id
}

locals {
  credentials = {
    # ephemeral
    ephemeral_access_token = ephemeral.twingate_connector_tokens.aws_connector_tokens.access_token

    # non-ephemeral
    access_token = resource.twingate_connector_tokens.aws_connector_tokens.access_token
  }
}
