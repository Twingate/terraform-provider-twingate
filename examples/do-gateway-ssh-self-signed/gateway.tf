locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port
  })
}

resource "digitalocean_reserved_ip" "gateway" {
  region = var.do_region
}

resource "digitalocean_reserved_ip_assignment" "gateway" {
  ip_address = digitalocean_reserved_ip.gateway.ip_address
  droplet_id = digitalocean_droplet.gateway.id
}

# replace_triggered_by only accepts resource references, so the rendered config is
# wrapped here to replace the gateway whenever it changes.
resource "terraform_data" "gateway_config" {
  input = local.gateway_config
}

resource "digitalocean_droplet" "gateway" {
  name   = "demo-gateway"
  region = var.do_region
  size   = var.do_droplet_size
  image  = "debian-12-x64"

  vpc_uuid = data.digitalocean_vpc.main.id

  user_data = templatefile("${path.module}/scripts/gateway-startup.sh", {
    tls_cert       = tls_locally_signed_cert.server.cert_pem
    tls_key        = tls_private_key.server.private_key_pem
    ssh_ca_key     = tls_private_key.ssh_ca.private_key_openssh
    gateway_config = local.gateway_config
  })

  lifecycle {
    replace_triggered_by = [
      terraform_data.gateway_config,
    ]
  }
}
