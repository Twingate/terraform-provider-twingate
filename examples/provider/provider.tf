provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "autoco"

  cache = {
    resource_enabled = false
    groups_enabled = true
  }
}

