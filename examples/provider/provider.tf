provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "autoco"
}

variable "network" {
  default = "autoco"
}