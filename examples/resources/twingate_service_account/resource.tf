provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_service_account" "github_actions_prod" {
  name = "Github Actions PROD"
}