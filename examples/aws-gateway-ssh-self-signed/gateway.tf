locals {
  gateway_port = 8443

  # Pinned so twingate_gateway can use the address before the instance exists.
  gateway_private_ip = cidrhost(aws_subnet.main.cidr_block, 10)

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network   = var.tg_network
    twingate_host      = var.tg_url
    port               = local.gateway_port
    ssh_server_name    = twingate_ssh_resource.ssh_server.name
    ssh_server_address = aws_instance.ssh_server.private_ip
  })
}

# replace_triggered_by only accepts resource references, so the rendered config is
# wrapped here to replace the gateway whenever it changes.
resource "terraform_data" "gateway_config" {
  input = local.gateway_config
}

resource "aws_instance" "gateway" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.main.id
  private_ip             = local.gateway_private_ip
  vpc_security_group_ids = [aws_security_group.internal.id]

  user_data = templatefile("${path.module}/scripts/gateway-startup.sh", {
    tls_cert       = tls_locally_signed_cert.server.cert_pem
    tls_key        = tls_private_key.server.private_key_pem
    ssh_ca_key     = tls_private_key.ssh_ca.private_key_openssh
    gateway_config = local.gateway_config
  })

  root_block_device {
    encrypted = true
  }

  lifecycle {
    replace_triggered_by = [
      terraform_data.gateway_config,
    ]
  }

  tags = { Name = "demo-gateway" }
}
