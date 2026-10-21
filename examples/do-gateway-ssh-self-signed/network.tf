data "digitalocean_vpc" "main" {
  region = var.do_region
}

# Allows inbound traffic only from inside the VPC. Every droplet only needs outbound
# connections to the internet (Twingate, package mirrors).
resource "digitalocean_firewall" "vpc_only" {
  name = "demo-ssh-vpc-only"

  droplet_ids = [
    digitalocean_droplet.gateway.id,
    digitalocean_droplet.connector.id,
    digitalocean_droplet.ssh_server.id,
  ]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "1-65535"
    source_addresses = [data.digitalocean_vpc.main.ip_range]
  }

  inbound_rule {
    protocol         = "udp"
    port_range       = "1-65535"
    source_addresses = [data.digitalocean_vpc.main.ip_range]
  }

  inbound_rule {
    protocol         = "icmp"
    source_addresses = [data.digitalocean_vpc.main.ip_range]
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
