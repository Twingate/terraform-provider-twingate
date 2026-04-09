terraform {
  required_providers {
    twingate = {
      source  = "Twingate/twingate"
      version = "~> 4.1"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
}

provider "digitalocean" {
  token = var.do_token
}
