## keeping connector up to date , once the image changes the helm deployment will be updated
data "docker_registry_image" "connector" {
  name = "twingate/connector:1"
}

resource "helm_release" "connector" {
  chart         = "connector"
  name          = "connector"
  repository    = "https://twingate.github.io/helm-charts"
  version       = "0.1.5"
  recreate_pods = true

  set {
    name  = "connector.network"
    value = "[NETWORK_NAME_HERE]"
  }

  set {
    name  = "image.tag"
    value = "1"
  }

  set {
    name  = "sha256"
    value = data.docker_registry_image.connector.sha256_digest
  }

  set {
    name  = "connector.accessToken"
    value = twingate_connector_tokens.aws_connector_tokens.access_token
  }

  set {
    name  = "connector.refreshToken"
    value = twingate_connector_tokens.aws_connector_tokens.refresh_token
  }

}