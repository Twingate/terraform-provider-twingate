---
subcategory: "k8s"
page_title: "Deploy connect with helm - Twingate Provider"
description: |-
An example of how to deploy a connector with helm and keep it up to date
---

## Deploy a connector with a helm provider

providers used in the example

```terraform
terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "2.1.0"
    }
    twingate = {
      source  = "Twingate/twingate"
      version = "0.0.4"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "2.11.0"
    }
  }
}
```

Helm deployment

```terraform
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
```