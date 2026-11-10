locals {
  httpbin_port = 8080
}

resource "aws_instance" "httpbin" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.main.id
  vpc_security_group_ids = [aws_security_group.internal.id]

  user_data = templatefile("${path.module}/scripts/httpbin-startup.sh", {
    httpbin_port = local.httpbin_port
  })

  # cloud-init only runs user_data on first boot, so a port change replaces the
  # instance instead of updating it in place.
  user_data_replace_on_change = true

  root_block_device {
    encrypted = true
  }

  tags = { Name = "demo-httpbin" }

  # The startup script installs Docker and pulls the image, so the full outbound
  # path (private subnet -> NAT gateway -> internet gateway) must exist before the
  # instance launches.
  depends_on = [
    aws_route_table_association.private,
    aws_route_table_association.public,
  ]
}
