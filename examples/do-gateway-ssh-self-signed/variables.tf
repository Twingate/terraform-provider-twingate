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
  description = "Optional DNS alias for the SSH resource (added as a DNS SAN in the TLS cert)"
  type        = string
  default     = ""
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
  description = "DigitalOcean droplet size"
  type        = string
  default     = "s-1vcpu-1gb"
}
