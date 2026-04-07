resource "digitalocean_vpc" "main" {
  name        = "demo-network"
  region      = var.do_region
  description = "VPC for Twingate SSH demo"
}

resource "digitalocean_firewall" "main" {
  name = "demo-firewall"

  droplet_ids = [
    digitalocean_droplet.ssh_server.id,
    digitalocean_droplet.connector.id,
    digitalocean_droplet.gateway.id,
  ]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "8443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}
