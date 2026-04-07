locals {
  gateway_port = 8443
}

resource "digitalocean_reserved_ip" "gateway" {
  region = var.do_region
}

resource "digitalocean_reserved_ip_assignment" "gateway" {
  ip_address = digitalocean_reserved_ip.gateway.ip_address
  droplet_id = digitalocean_droplet.gateway.id
}

resource "twingate_gateway_config" "config" {
  port = local.gateway_port

  tls = {
    certificate_file = "/etc/gateway/tls.crt"
    private_key_file = "/etc/gateway/tls.key"
  }

  ssh = {
    gateway = { username = "gateway" }
    ca      = { private_key_file = "/opt/gateway/ssh-ca.key" }

    resources = [
      twingate_ssh_resource.ssh_server,
    ]
  }
}

resource "digitalocean_droplet" "gateway" {
  name   = "demo-gateway"
  region = var.do_region
  size   = var.do_droplet_size
  image  = "debian-12-x64"

  vpc_uuid = digitalocean_vpc.main.id

  user_data = templatefile("${path.module}/scripts/gateway-startup.sh", {
    tls_cert       = tls_locally_signed_cert.server.cert_pem
    tls_key        = tls_private_key.server.private_key_pem
    ssh_ca_key     = tls_private_key.ssh_ca.private_key_openssh
    gateway_config = twingate_gateway_config.config.content
  })

  lifecycle {
    replace_triggered_by = [
      twingate_gateway_config.config,
    ]
  }
}
