provider "twingate" {
  # api_token = "1234567890abcdef"
  # network   = "mynetwork"
}

resource "twingate_remote_network" "aws_network" {
  name = "aws_remote_network-test2"
}

resource "twingate_connector" "aws_connector" {
  remote_network_id = twingate_remote_network.aws_network.id
  status_updates_enabled = true
}