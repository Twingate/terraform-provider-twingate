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


// Key rotation using the [time](https://registry.terraform.io/providers/hashicorp/time/latest) provider**

resource "time_rotating" "key_rotation" {
  rotation_days = 30
}

resource "twingate_service_account_key" "github_key_with_rotation" {
  name = "Github Actions PROD key (automatically rotating)"
  service_account_id = twingate_service_account.github_actions_prod.id
  lifecycle {
    replace_triggered_by = [time_rotating.key_rotation.rfc3339]
  }
}