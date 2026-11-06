locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port
  })

  gateway_metadata = {
    # Leaf followed by the Let's Encrypt intermediate, so clients can build the chain to the root
    tls-cert       = "${acme_certificate.gateway.certificate_pem}${acme_certificate.gateway.issuer_pem}"
    tls-key        = acme_certificate.gateway.private_key_pem
    gateway-config = local.gateway_config
  }
}

resource "google_compute_address" "gateway" {
  name         = "demo-https-public-ca-gateway-ip"
  subnetwork   = google_compute_subnetwork.main.id
  address_type = "INTERNAL"
}

resource "terraform_data" "gateway_metadata" {
  input = local.gateway_metadata
}

resource "google_compute_instance" "gateway" {
  name         = "demo-https-public-ca-gateway"
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

  # The startup script downloads the Gateway binary, so NAT must exist before boot
  depends_on = [google_compute_router_nat.main]

  lifecycle {
    replace_triggered_by = [terraform_data.gateway_metadata]
  }
}
