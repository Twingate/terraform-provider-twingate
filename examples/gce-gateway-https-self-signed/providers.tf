terraform {
  required_providers {
    twingate = {
      source  = "Twingate/twingate"
      version = "~> 5.2"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
  url       = var.tg_url
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}
