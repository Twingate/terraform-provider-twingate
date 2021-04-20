data "google_client_config" "provider" {}

data "google_container_cluster" "cluster" {
  name     = "beamreachinc-stg"
  location = "us-central1-a"
}


