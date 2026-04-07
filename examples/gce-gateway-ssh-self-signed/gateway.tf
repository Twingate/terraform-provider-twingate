locals {
  gateway_port = 8443

  gateway_metadata = {
    tls-cert       = tls_locally_signed_cert.server.cert_pem
    tls-key        = tls_private_key.server.private_key_pem
    ssh-ca-key     = tls_private_key.ssh_ca.private_key_openssh
    gateway-config = twingate_gateway_config.config.content
  }
}

resource "google_compute_address" "gateway" {
  name         = "demo-gateway-ip"
  subnetwork   = google_compute_subnetwork.main.id
  address_type = "INTERNAL"
  region       = var.region
}

resource "twingate_gateway_config" "config" {
  port = local.gateway_port

  tls = {
    certificate_file = "/opt/gateway/tls.crt"
    private_key_file = "/opt/gateway/tls.key"
  }

  ssh = {
    gateway = {
      username = "gateway"
    }

    ca = {
      private_key_file = "/opt/gateway/ssh-ca.key"
    }

    resources = [twingate_ssh_resource.ssh_server]
  }
}

resource "terraform_data" "gateway_metadata" {
  input = local.gateway_metadata
}

resource "google_compute_instance" "gateway" {
  name         = "demo-gateway"
  machine_type = "e2-micro"
  zone         = var.zone

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main.id
    network_ip = google_compute_address.gateway.address
  }

  metadata = local.gateway_metadata

  metadata_startup_script = file("${path.module}/scripts/gateway-startup.sh")

  lifecycle {
    replace_triggered_by = [terraform_data.gateway_metadata]
  }
}
