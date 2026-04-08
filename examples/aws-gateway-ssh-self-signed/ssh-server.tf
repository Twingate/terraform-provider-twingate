resource "aws_instance" "ssh_server" {
  ami                    = data.aws_ami.debian.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.main.id
  vpc_security_group_ids = [aws_security_group.internal.id]

  user_data = templatefile("${path.module}/scripts/ssh-server-startup.sh", {
    ssh_ca_public_key = tls_private_key.ssh_ca.public_key_openssh
  })

  root_block_device {
    encrypted = true
  }

  tags = { Name = "demo-ssh-server" }
}
