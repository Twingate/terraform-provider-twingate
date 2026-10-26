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

variable "domain" {
  description = "Domain covered by the wildcard certificate (*.<domain>). Must be in a public Cloud DNS zone in project_id. The Resource alias is httpbin.<domain>."
  type        = string
}

variable "acme_email" {
  description = "Email address for the Let's Encrypt account (expiry notices)"
  type        = string
}

variable "letsencrypt_staging" {
  description = "Issue from the Let's Encrypt staging environment, which has higher rate limits but is not publicly trusted"
  type        = bool
  default     = false
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
