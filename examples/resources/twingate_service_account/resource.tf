provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_service_account" "aws" {
  name = "aws_account"
}
