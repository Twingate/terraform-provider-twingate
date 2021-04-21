variable "api_token" {
  type = string
  sensitive = true
}
variable "network" {
  type = string
}

variable "url" {
  type = string
}

variable "gke_cluster_to_deploy" {
  type = string
}