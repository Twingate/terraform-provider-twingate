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

resource "twingate_group" "devops" {
  name = "DevOps"
}

resource "twingate_service_account" "ci" {
  name = "CI"
}

data "twingate_security_policy" "test_policy" {
  name = "Default Policy"
}

# Kubernetes resource accessed via in-cluster DNS
resource "twingate_kubernetes_resource" "prod_cluster" {
  name              = "Production K8s"
  gateway_id        = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id

  is_visible = true
  is_authoritative = true

  # Optional. Grant access to groups and service accounts.
  access_group {
    group_id = twingate_group.devops.id

    security_policy_id = data.twingate_security_policy.test_policy.id
    access_policy {
      mode          = "AUTO_LOCK"
      approval_mode = "AUTOMATIC"
      duration      = "30d"
    }
  }

  access_service {
    service_account_id = twingate_service_account.ci.id
  }

  security_policy_id = data.twingate_security_policy.test_policy.id

  access_policy {
    mode          = "AUTO_LOCK"
    approval_mode = "MANUAL"
    duration      = "12d"
  }

  tags = {
    environment = "dev"
    owner       = "devops"
    project     = "api"
  }
}

# Kubernetes resource accessed via external address
resource "twingate_kubernetes_resource" "external_cluster" {
  name              = "External K8s"
  address           = "k8s-api.example.com"
  gateway_id        = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
}