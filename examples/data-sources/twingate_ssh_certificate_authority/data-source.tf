provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

data "twingate_ssh_certificate_authority" "example" {
  id = "<your ssh certificate authority's id>"
}
