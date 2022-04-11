provider "twingate" {}

# find network by id
data "twingate_remote_network" "test" {
  id = "UmVtb3RlTmV0d29yazozOTU5Nw=="
}

# retrieve network name
output "my_network" {
  value = data.twingate_remote_network.test.name
}