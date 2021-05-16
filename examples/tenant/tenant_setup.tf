variable "tenant_name" {
}

resource "twingate_remote_network" "network" {
  name = "${var.tenant_name}-network"
}

resource "twingate_connector" "connector" {
  remote_network_id = twingate_remote_network.network.id
}

resource "twingate_connector_tokens" "connector_tokens" {
  connector_id = twingate_connector.connector.id
}

