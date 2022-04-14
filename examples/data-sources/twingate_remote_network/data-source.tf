provider "twingate" {}

# find network by name
data "twingate_remote_network" "net" {
  name = "tf-acc-1915686675692758735"
}

# retrieve network id
output "my_network_id" {
  value = data.twingate_remote_network.net.id
}