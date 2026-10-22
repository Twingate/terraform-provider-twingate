locals {
  httpbin_port = 8080
}

resource "google_compute_instance" "httpbin" {
  name         = "demo-httpbin"
  machine_type = "e2-micro"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main.id
  }

  tags = ["iap-ssh"]

  metadata = {
    httpbin-port = local.httpbin_port
  }

  metadata_startup_script = file("${path.module}/scripts/httpbin-startup.sh")
}
