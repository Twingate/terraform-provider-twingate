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
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "twingate" {
  api_token = var.tg_api_token
  network   = var.tg_network
}

provider "aws" {
  region = var.aws_region
}
