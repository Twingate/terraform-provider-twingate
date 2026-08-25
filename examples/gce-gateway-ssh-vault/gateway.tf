resource "vault_pki_secret_backend_cert" "gateway" {
  backend = vault_mount.pki.path
  name    = vault_pki_secret_backend_role.gateway.name

  common_name = "demo-gateway"
  alt_names   = [twingate_ssh_resource.ssh_server.alias]
  ip_sans     = [google_compute_instance.ssh_server.network_interface[0].network_ip]
  ttl         = "8736h" # ~364 days (must fit within the 1-year root CA)
}

resource "google_service_account" "gateway" {
  account_id   = "demo-gateway"
  display_name = "Demo Gateway"
}

resource "google_compute_address" "gateway" {
  name         = "demo-gateway-ip"
  subnetwork   = data.terraform_remote_state.vault.outputs.subnetwork_id
  address_type = "INTERNAL"
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
    subnetwork = data.terraform_remote_state.vault.outputs.subnetwork_id
    network_ip = google_compute_address.gateway.address
  }

  tags = ["iap-ssh"]

  service_account {
    email  = google_service_account.gateway.email
    scopes = ["cloud-platform"]
  }

  metadata_startup_script = templatefile("${path.module}/scripts/gateway-startup.sh", {
    tls-cert       = vault_pki_secret_backend_cert.gateway.certificate
    tls-key        = vault_pki_secret_backend_cert.gateway.private_key
    vault_ca_cert  = data.terraform_remote_state.vault.outputs.vault_tls_cert
    gateway-config = local.gateway_config
  })
}

locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port

    vault_address   = "https://${data.terraform_remote_state.vault.outputs.vault_internal_ip}:8200"
    vault_ssh_mount = vault_mount.ssh.path
    vault_ssh_role  = vault_ssh_secret_backend_role.gateway.name
    vault_gcp_mount = vault_auth_backend.gcp.path
    vault_gcp_role  = vault_gcp_auth_backend_role.gateway.role
  })
}
