data "google_client_config" "provider" {}

data "google_container_cluster" "cluster" {
  name     = var.gke_cluster_to_deploy
  location = "us-central1-a"
}


