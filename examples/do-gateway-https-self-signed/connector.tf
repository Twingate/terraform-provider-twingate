resource "digitalocean_droplet" "connector" {
  name   = "demo-connector"
  region = var.do_region
  size   = var.do_droplet_size
  image  = "debian-13-x64"

  vpc_uuid = data.digitalocean_vpc.main.id
  tags     = [digitalocean_tag.vpc_only.id]

  user_data = templatefile("${path.module}/scripts/connector-startup.sh", {
    access_token  = twingate_connector_tokens.main.access_token
    refresh_token = twingate_connector_tokens.main.refresh_token
    twingate_url  = "https://${var.tg_network}.${var.tg_url}"
  })

  depends_on = [digitalocean_firewall.vpc_only]
}
