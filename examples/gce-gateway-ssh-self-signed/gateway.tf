locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port
  })

  gateway_metadata = {
    tls-cert       = tls_locally_signed_cert.server.cert_pem
    tls-key        = tls_private_key.server.private_key_pem
    ssh-ca-key     = tls_private_key.ssh_ca.private_key_openssh
    gateway-config = local.gateway_config
  }
}

resource "google_compute_address" "gateway" {
  name         = "demo-gateway-ip"
  subnetwork   = google_compute_subnetwork.main.id
  address_type = "INTERNAL"
}

resource "terraform_data" "gateway_metadata" {
  input = local.gateway_metadata
}

resource "google_compute_instance" "gateway" {
  name         = "demo-gateway"
  machine_type = "e2-micro"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main.id
    network_ip = google_compute_address.gateway.address
  }

  tags = ["iap-ssh"]

  metadata = local.gateway_metadata

  metadata_startup_script = file("${path.module}/scripts/gateway-startup.sh")

  lifecycle {
    replace_triggered_by = [terraform_data.gateway_metadata]
  }
}
