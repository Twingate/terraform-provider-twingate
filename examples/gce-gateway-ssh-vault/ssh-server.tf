resource "google_service_account" "vm" {
  account_id   = "demo-vm"
  display_name = "Demo VM"
}

resource "google_compute_instance" "ssh_server" {
  name         = "demo-ssh-server"
  machine_type = "e2-micro"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  network_interface {
    subnetwork = data.terraform_remote_state.vault.outputs.subnetwork_id
  }

  service_account {
    email  = google_service_account.vm.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    enable-oslogin = "FALSE"
  }

  metadata_startup_script = templatefile("${path.module}/scripts/ssh-server-startup.sh", {
    ssh-ca-public-key = vault_ssh_secret_backend_ca.ssh.public_key
    vault_ca_cert     = data.terraform_remote_state.vault.outputs.vault_tls_cert
    vault_addr        = "https://${data.terraform_remote_state.vault.outputs.vault_internal_ip}:8200"
    vault_mount       = vault_mount.ssh.path
    vault_role        = vault_ssh_secret_backend_role.ssh_server.name
    vault_gcp_role    = vault_gcp_auth_backend_role.vm.role
    vault_gcp_mount   = vault_auth_backend.gcp.path
  })
}
