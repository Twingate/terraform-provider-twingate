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

resource "twingate_connector" "connector" {
  remote_network_id = twingate_remote_network.network.id
}

resource "twingate_connector_tokens" "tokens" {
  connector_id = twingate_connector.connector.id
}