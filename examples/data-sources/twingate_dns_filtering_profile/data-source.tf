provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

data "twingate_dns_filtering_profile" "example" {
  id = "<your dns profile's id>"
}

