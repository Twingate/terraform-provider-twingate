output "gateway_instance_id" {
  description = "EC2 instance ID of the gateway"
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
