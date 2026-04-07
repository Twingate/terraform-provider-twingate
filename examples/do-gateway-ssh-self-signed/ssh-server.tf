resource "digitalocean_droplet" "ssh_server" {
  name   = "demo-ssh-server"
  region = var.do_region
  size   = var.do_droplet_size
  image  = "debian-12-x64"

  vpc_uuid = digitalocean_vpc.main.id

  user_data = templatefile("${path.module}/scripts/ssh-server-startup.sh", {
    ssh_ca_public_key = tls_private_key.ssh_ca.public_key_openssh
  })
}
