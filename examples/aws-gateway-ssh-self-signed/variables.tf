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

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.large"
}
