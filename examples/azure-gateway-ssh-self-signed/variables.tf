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
  description = "Optional alias for the SSH resource, added as a SAN in the TLS cert"
  type        = string
  default     = ""
}

variable "subscription_id" {
  description = "Azure subscription ID"
  type        = string
}

variable "location" {
  description = "Azure region. Must offer the Azure Bastion Developer SKU (eastus2 does, eastus doesn't)"
  type        = string
  default     = "eastus2"
}

variable "vm_size" {
  description = "Azure VM size. Must be x86_64: the startup scripts download amd64 binaries"
  type        = string
  default     = "Standard_B1s"
}
