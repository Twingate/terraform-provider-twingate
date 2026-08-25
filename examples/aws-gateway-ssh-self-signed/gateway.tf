locals {
  gateway_port = 8443

  gateway_config = templatefile("${path.module}/config.yaml.tftpl", {
    twingate_network = var.tg_network
    twingate_host    = var.tg_url
    port             = local.gateway_port
  })
}

# A dedicated ENI gives the gateway a stable private IP that is known before the
# instance is created, so twingate_gateway can use it as its address.
resource "aws_network_interface" "gateway" {
  subnet_id       = aws_subnet.main.id
  security_groups = [aws_security_group.internal.id]

  tags = { Name = "demo-gateway-eni" }
}

# replace_triggered_by only accepts resource references, so the rendered config is
# wrapped here to replace the gateway whenever it changes.
resource "terraform_data" "gateway_config" {
  input = local.gateway_config
}

resource "aws_instance" "gateway" {
  ami           = data.aws_ami.debian.id
  instance_type = var.instance_type
  key_name      = aws_key_pair.debug_key.key_name

  network_interface {
    network_interface_id = aws_network_interface.gateway.id
    device_index         = 0
  }

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
