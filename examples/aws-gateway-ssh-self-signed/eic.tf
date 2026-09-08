resource "aws_ec2_instance_connect_endpoint" "main" {
  subnet_id          = aws_subnet.main.id
  security_group_ids = [aws_security_group.eic.id]
  preserve_client_ip = false

  tags = { Name = "demo-eic-endpoint" }
}
