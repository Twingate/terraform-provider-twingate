resource "aws_instance" "connector" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.main.id
  vpc_security_group_ids = [aws_security_group.internal.id]

  user_data = templatefile("${path.module}/scripts/connector-startup.sh", {
    access_token  = twingate_connector_tokens.main.access_token
    refresh_token = twingate_connector_tokens.main.refresh_token
    twingate_url  = "https://${var.tg_network}.${var.tg_url}"
  })

  root_block_device {
    encrypted = true
  }

  tags = { Name = "demo-https-connector" }

  # The startup script runs once at boot and downloads the Connector, so the full
  # outbound path (private subnet -> NAT gateway -> internet gateway) must exist
  # before the instance launches.
  depends_on = [
    aws_route_table_association.private,
    aws_route_table_association.public,
  ]
}
