provider "twingate" {
#  api_token = "1234567890abcdef"
#  network   = "mynetwork"
}

resource "twingate_remote_network" "aws_network" {
  name = "tf-acc-8039014946868982466"
}

resource "twingate_connector" "aws_connector" {
  remote_network_id = "UmVtb3RlTmV0d29yazo0MDA2NA=="
  name = "updated connector name"
}