resource "google_compute_network" "main" {
  name                    = "demo-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "main" {
  name          = "demo-subnetwork"
  ip_cidr_range = "172.16.0.0/24"
  network       = google_compute_network.main.id
  region        = var.region
}

# Allow all internal TCP traffic within the subnet
resource "google_compute_firewall" "allow_internal" {
  name    = "demo-allow-internal"
  network = google_compute_network.main.id

  allow {
    protocol = "tcp"
  }

  source_ranges = [google_compute_subnetwork.main.ip_cidr_range]
}

# Allow specific ports via IAP tunnel
resource "google_compute_firewall" "allow_iap" {
  name    = "demo-allow-iap"
  network = google_compute_network.main.id

  allow {
    protocol = "tcp"
    ports    = ["22", "8200"] # SSH, Vault API
  }

  source_ranges = ["35.235.240.0/20"]
  target_tags   = ["demo-vault-server", "iap-ssh"]
}

# --- Cloud NAT for outbound internet access (apt-get, HashiCorp repo) ---

resource "google_compute_router" "main" {
  name    = "demo-router"
  region  = var.region
  network = google_compute_network.main.id
}

resource "google_compute_router_nat" "main" {
  name                               = "demo-nat"
  router                             = google_compute_router.main.name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}
