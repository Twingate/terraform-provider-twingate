
resource "twingate_remote_network" "test_remote_network" {
  name = "hello_from_terraform"
}

resource "twingate_connector" "test_connector" {
  remote_network_id = twingate_remote_network.test_remote_network.id
}

resource "helm_release" "connector" {
  name       = "connector"
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
    value = twingate_connector.test_connector.access_token
  }

  set_sensitive  {
    name  = "connector.refreshToken"
    value = twingate_connector.test_connector.refresh_token
  }

  depends_on = [twingate_connector.test_connector]
}


