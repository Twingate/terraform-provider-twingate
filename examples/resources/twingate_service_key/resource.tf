provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_service" "github_actions_prod" {
  name = "Github Actions PROD"
}

resource "twingate_service_key" "github_key" {
  name = "Github Actions PROD key"
  service = twingate_service.github_actions_prod.id
}
