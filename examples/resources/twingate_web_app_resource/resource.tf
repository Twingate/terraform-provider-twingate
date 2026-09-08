provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_remote_network" "prod" {
  name = "Production Network"
}

resource "twingate_x509_certificate_authority" "tls" {
  name        = "My TLS CA"
  certificate = file("ca.pem")
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.prod.id
  address           = "10.0.0.1:8443"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
}

resource "twingate_web_app_resource" "internal_app" {
  name              = "Internal App"
  gateway_id        = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
  address           = "internal.acme.com"
  alias             = "app.int"
  downstream = {
    port = 80
  }
  upstream = {
    port = 8080
  }
  request_header_rewrites = {
    "X-Twingate-User" = "{{username}}"
  }
}
