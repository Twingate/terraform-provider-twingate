provider "google" {
  project = "twingate-qa"
  zone = "us-central1"
}

provider "helm" {
  kubernetes {
    host  = "https://${data.google_container_cluster.cluster.endpoint}"
    token = data.google_client_config.provider.access_token
    cluster_ca_certificate = base64decode(
    data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate
    )
  }
}

provider "twingate" {
  api_token = var.api_token
  network = var.network
  url = var.url
}