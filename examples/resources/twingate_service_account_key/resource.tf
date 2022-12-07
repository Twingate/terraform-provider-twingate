provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_service_account" "github_actions_prod" {
  name = "Github Actions PROD"
}

resource "twingate_service_account_key" "github_key" {
  name = "Github Actions PROD key"
  service_account_id = twingate_service_account.github_actions_prod.id
}
