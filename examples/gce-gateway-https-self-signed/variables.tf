variable "tg_api_token" {
  description = "Twingate API token"
  type        = string
  sensitive   = true
}

variable "tg_network" {
  description = "Twingate network name"
  type        = string
}

variable "tg_url" {
  description = "Twingate URL base domain"
  type        = string
  default     = "twingate.com"
}

variable "resource_alias" {
  description = "DNS alias for the HTTPS resource (added as a DNS SAN in the TLS cert)"
  type        = string
  default     = "gce-http-bin.demo.int"
}

variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "GCP zone"
  type        = string
  default     = "us-central1-a"
}
