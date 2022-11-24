provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_service_account" "aws" {
  name = "aws_account"
}

resource "twingate_service_account_key" "aws" {
  name = "aws_account_key"
  service_account_id = twingate_service_account.aws.id
}
