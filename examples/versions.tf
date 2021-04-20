terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "3.64.0"
    }
    twingate = {
      version = "0.1"
      source = "twingate/twingate"
    }
    helm = {
      source = "hashicorp/helm"
      version = "2.1.0"
    }
  }
  required_version = "= 0.15.0"
}
