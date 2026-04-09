resource "google_compute_instance" "connector" {
  name         = "demo-connector"
  machine_type = "e2-micro"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    subnetwork = data.terraform_remote_state.vault.outputs.subnetwork_id
  }

  metadata = {
    connector-access-token  = twingate_connector_tokens.main.access_token
    connector-refresh-token = twingate_connector_tokens.main.refresh_token
    connector-url           = "https://${var.tg_network}.twingate.com"
  }

  metadata_startup_script = file("${path.module}/scripts/connector-startup.sh")
}
