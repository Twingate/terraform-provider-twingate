resource "google_compute_instance" "ssh_server" {
  name         = "demo-ssh-server"
  machine_type = "e2-micro"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main.id
  }

  metadata = {
    enable-oslogin    = "FALSE" # Disabled because the Gateway uses SSH CA-based authentication
    ssh-ca-public-key = tls_private_key.ssh_ca.public_key_openssh
  }

  metadata_startup_script = file("${path.module}/scripts/ssh-server-startup.sh")

}
