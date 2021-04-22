
resource "twingate_remote_network" "test_remote_network" {
  name = "hello_from_terraform"
}

resource "twingate_connector" "test_connector" {
  remote_network_id = twingate_remote_network.test_remote_network.id
}

resource "twingate_connector_tokens" "test_tokens" {
  connector_id = twingate_connector.test_connector.id
  keepers = {
    foo = "bar"
  }
}

resource "helm_release" "connector" {
  name       = replace(twingate_connector.test_connector.name, "_", "-")
  chart      = "connector"
  repository = "https://twingate.github.io/helm-charts"
  namespace  = "default"
  version    = "0.1.5"

  set {
    name  = "connector.url"
    value = var.url
  }

  set {
    name  = "connector.network"
    value = var.network
  }

  set_sensitive  {
    name  = "connector.accessToken"
    value = twingate_connector_tokens.test_tokens.access_token
  }

  set_sensitive  {
    name  = "connector.refreshToken"
    value = twingate_connector_tokens.test_tokens.refresh_token
  }

  depends_on = [twingate_connector_tokens.test_tokens]
}


