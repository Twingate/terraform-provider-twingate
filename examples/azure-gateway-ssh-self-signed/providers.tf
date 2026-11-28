terraform {
  required_providers {
    twingate = {
      source  = "Twingate/twingate"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
  url       = var.tg_url
}

# Credentials come from the Azure CLI (`az login`).
provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}
