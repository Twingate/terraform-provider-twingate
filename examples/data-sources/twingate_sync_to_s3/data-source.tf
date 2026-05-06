provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

data "twingate_sync_to_s3" "example" {}

output "oidc_url" {
  value = data.twingate_sync_to_s3.example.oidc_url
}

output "oidc_prefix" {
  value = data.twingate_sync_to_s3.example.oidc_prefix
}
