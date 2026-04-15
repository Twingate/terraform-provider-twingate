variable "tg_api_token" {
  description = "Twingate API token"
  type        = string
  sensitive   = true
}

variable "tg_network" {
  description = "Twingate network name"
  type        = string
}

variable "resource_alias" {
  description = "Optional alias for the SSH resource, added as a SAN in the TLS cert"
  type        = string
  default     = ""
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

variable "vault_addr" {
  description = "Vault address (use https://localhost:8200 with IAP tunnel)"
  type        = string
  default     = "https://localhost:8200"
}

variable "vault_token" {
  description = "Vault root token"
  type        = string
  sensitive   = true
}
