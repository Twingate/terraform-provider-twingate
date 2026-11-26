output "instance_ids" {
  description = "EC2 instance IDs, for SSH access with EC2 Instance Connect"
  value = {
    gateway   = aws_instance.gateway.id
    connector = aws_instance.connector.id
    httpbin   = aws_instance.httpbin.id
  }
}
