resource "aws_instance" "connector" {
  ami                    = data.aws_ami.debian.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.private.id
  vpc_security_group_ids = [aws_security_group.internal.id]

  user_data = templatefile("${path.module}/scripts/connector-startup.sh", {
    access_token  = twingate_connector_tokens.main.access_token
    refresh_token = twingate_connector_tokens.main.refresh_token
    twingate_url  = "https://${var.tg_network}.twingate.com"
  })

  root_block_device {
    encrypted = true
  }

  tags = { Name = "demo-connector" }
}
