locals {
  gateway_port = 8443
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

resource "aws_instance" "gateway" {
  ami                    = data.aws_ami.debian.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.private.id
  vpc_security_group_ids = [aws_security_group.internal.id]

  user_data = templatefile("${path.module}/scripts/gateway-startup.sh", {
    tls_cert       = tls_locally_signed_cert.server.cert_pem
    tls_key        = tls_private_key.server.private_key_pem
    ssh_ca_key     = tls_private_key.ssh_ca.private_key_openssh
    gateway_config = twingate_gateway_config.config.content
  })

  root_block_device {
    encrypted = true
  }

  lifecycle {
    replace_triggered_by = [
      twingate_gateway_config.config,
    ]
  }

  tags = { Name = "demo-gateway" }
}
