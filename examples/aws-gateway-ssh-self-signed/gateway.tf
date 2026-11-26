locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port
  })
}

resource "aws_instance" "gateway" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.main.id
  vpc_security_group_ids = [aws_security_group.internal.id]

  user_data = templatefile("${path.module}/scripts/gateway-startup.sh", {
    tls_cert       = tls_locally_signed_cert.server.cert_pem
    tls_key        = tls_private_key.server.private_key_pem
    ssh_ca_key     = tls_private_key.ssh_ca.private_key_openssh
    gateway_config = local.gateway_config
  })

  # cloud-init only runs user_data on first boot, so a new certificate, key or
  # config replaces the instance instead of updating it in place.
  user_data_replace_on_change = true

  root_block_device {
    encrypted = true
  }

  tags = { Name = "demo-gateway" }

  # The startup script runs once at boot and downloads the Gateway binary, so the
  # full outbound path (private subnet -> NAT gateway -> internet gateway) must
  # exist before the instance launches.
  depends_on = [
    aws_route_table_association.private,
    aws_route_table_association.public,
  ]
}
