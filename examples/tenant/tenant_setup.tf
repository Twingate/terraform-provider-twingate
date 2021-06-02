resource "twingate_remote_network" "aws_remote_network" {
  name = "aws-remote-network"
}

resource "twingate_connector" "aws_connector" {
  remote_network_id = twingate_remote_network.aws_remote_network.id
}

resource "twingate_connector_tokens" "aws_connector_tokens" {
  connector_id = twingate_connector.aws_connector.id
}

