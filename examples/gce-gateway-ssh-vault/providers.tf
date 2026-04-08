terraform {
  required_providers {
    twingate = {
      source  = "Twingate/twingate"
      version = "~> 4.1"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "~> 5.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

provider "vault" {
  address = var.vault_addr
  token   = var.vault_token
  # Running Terraform from a local machine; skip TLS verification to avoid adding the self-signed CA
  skip_tls_verify = true
}
