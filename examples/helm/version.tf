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