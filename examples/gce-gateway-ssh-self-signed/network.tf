resource "google_compute_network" "main" {
  name                    = "demo-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "main" {
  name          = "demo-subnetwork"
  ip_cidr_range = "172.16.0.0/24"
  network       = google_compute_network.main.id
}

resource "google_compute_firewall" "internal" {
  name    = "demo-firewall"
  network = google_compute_network.main.id

  allow {
    protocol = "tcp"
  }

  source_ranges = [google_compute_subnetwork.main.ip_cidr_range]
}

# Allow SSH from the IAP tunnel so the VMs (no external IP) can be reached for debugging
resource "google_compute_firewall" "allow_iap" {
  name    = "demo-allow-iap"
  network = google_compute_network.main.id

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["35.235.240.0/20"]
  target_tags   = ["iap-ssh"]
}

# Cloud NAT allows instances without external IPs to reach the internet
resource "google_compute_router" "main" {
  name    = "demo-router"
  network = google_compute_network.main.id
}

resource "google_compute_router_nat" "main" {
  name                               = "demo-nat"
  router                             = google_compute_router.main.name
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}
