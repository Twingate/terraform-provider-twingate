data "digitalocean_vpc" "main" {
  region = var.do_region
}

resource "digitalocean_tag" "vpc_only" {
  name = "demo-https-vpc-only"
}

# Allows inbound traffic only from inside the VPC. Every Droplet only needs outbound
# connections to the internet (Twingate, package mirrors, container images).
resource "digitalocean_firewall" "vpc_only" {
  name = "demo-https-vpc-only"
  tags = [digitalocean_tag.vpc_only.id]

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
