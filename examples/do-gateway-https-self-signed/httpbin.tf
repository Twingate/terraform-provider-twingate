locals {
  httpbin_port = 8080
}

# The Marketplace "Docker on Ubuntu 22.04" image ships with Docker preinstalled.
resource "digitalocean_droplet" "httpbin" {
  name   = "demo-httpbin"
  region = var.do_region
  size   = var.do_droplet_size
  image  = "docker-20-04"

  vpc_uuid = data.digitalocean_vpc.main.id

  user_data = <<-EOT
    #!/bin/bash
    docker run -d --restart always -p ${local.httpbin_port}:8080 mccutchen/go-httpbin
  EOT
}
