provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

data "twingate_gateway" "example" {
  id = "<your gateway's id>"
}