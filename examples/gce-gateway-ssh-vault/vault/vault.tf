# --- Enable required APIs ---

resource "google_project_service" "compute" {
  service            = "compute.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "iam" {
  service            = "iam.googleapis.com"
  disable_on_destroy = false
}

# --- Service account for Vault VM ---

resource "google_service_account" "vault" {
  account_id   = "demo-vault"
  display_name = "Demo Vault"
  depends_on   = [google_project_service.iam]
}

# Vault needs these permissions to verify GCE instance identity for GCP auth
resource "google_project_iam_member" "vault_compute_viewer" {
  project = var.project_id
  role    = "roles/compute.viewer"
  member  = "serviceAccount:${google_service_account.vault.email}"
}

resource "google_project_iam_member" "vault_sa_viewer" {
  project = var.project_id
  role    = "roles/iam.serviceAccountViewer"
  member  = "serviceAccount:${google_service_account.vault.email}"
}

# --- Static IP for Vault VM ---

resource "google_compute_address" "vault" {
  name         = "demo-vault-ip"
  subnetwork   = google_compute_subnetwork.main.id
  address_type = "INTERNAL"
}

# --- TLS Certificates (self-signed for testing) ---

resource "tls_private_key" "vault" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "tls_self_signed_cert" "vault" {
  private_key_pem = tls_private_key.vault.private_key_pem

  subject {
    common_name  = "vault.internal"
    organization = "Twingate Testing"
  }

  validity_period_hours = 8760 # 1 year
  is_ca_certificate     = false

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]

  dns_names = [
    "vault.internal",
    "localhost",
  ]

  ip_addresses = ["127.0.0.1", google_compute_address.vault.address]
}

# --- Persistent Disk for Vault Data ---

resource "google_compute_disk" "vault_data" {
  name = "demo-vault-server-data"
  size = 10

  depends_on = [google_project_service.compute]
}

# --- Vault VM ---

resource "google_compute_instance" "vault" {
  name                      = "demo-vault-server"
  machine_type              = "e2-small"
  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
    }
  }

  attached_disk {
    source      = google_compute_disk.vault_data.id
    device_name = "demo-vault-data"
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main.id
    network_ip = google_compute_address.vault.address
  }

  tags = ["demo-vault-server"]

  service_account {
    email  = google_service_account.vault.email
    scopes = ["cloud-platform"]
  }

  metadata_startup_script = templatefile("${path.module}/scripts/vault-startup.sh", {
    vault_tls_cert = tls_self_signed_cert.vault.cert_pem
    vault_tls_key  = tls_private_key.vault.private_key_pem
    disk_name      = "demo-vault-data"
  })

  depends_on = [google_project_service.compute]
}
