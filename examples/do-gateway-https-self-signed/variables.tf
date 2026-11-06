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
  default     = "do-nyc-http-bin.demo.int"
}

variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "do_region" {
  description = "DigitalOcean region for resources"
  type        = string
  default     = "nyc3"
}

variable "do_droplet_size" {
  description = "DigitalOcean Droplet size"
  type        = string
  default     = "s-1vcpu-1gb"
}
