output "vm_admin_private_key" {
  description = "Private key for the VMs' azureadmin user, for SSH through Azure Bastion"
  value       = tls_private_key.vm_admin.private_key_openssh
  sensitive   = true
}
