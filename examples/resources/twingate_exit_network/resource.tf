provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_exit_network" "aws_network" {
  name = "aws_exit_network"
}
