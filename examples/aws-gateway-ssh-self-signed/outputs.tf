output "gateway_instance_id" {
  description = "EC2 instance ID of the gateway (use with: aws ec2-instance-connect ssh --instance-id <id> --os-user admin)"
  value       = aws_instance.gateway.id
}

output "connector_instance_id" {
  description = "EC2 instance ID of the connector"
  value       = aws_instance.connector.id
}

output "ssh_server_instance_id" {
  description = "EC2 instance ID of the SSH server"
  value       = aws_instance.ssh_server.id
}
