resource "digitalocean_droplet" "connector" {
  name   = "demo-connector"
  region = var.do_region
  size   = var.do_droplet_size
  image  = "debian-12-x64"

  vpc_uuid = data.digitalocean_vpc.main.id

  user_data = templatefile("${path.module}/scripts/connector-startup.sh", {
    access_token  = twingate_connector_tokens.main.access_token
    refresh_token = twingate_connector_tokens.main.refresh_token
    twingate_url  = "https://${var.tg_network}.twingate.com"
  })
}
