provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

data "twingate_x509_certificate_authority" "example" {
  id = "<your x509 certificate authority's id>"
}
