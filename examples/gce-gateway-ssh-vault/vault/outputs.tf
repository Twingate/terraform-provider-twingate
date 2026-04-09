output "vault_internal_ip" {
  description = "Internal IP of the Vault VM"
  value       = google_compute_instance.vault.network_interface[0].network_ip
}

output "vault_tls_cert" {
  description = "Vault server TLS certificate (self-signed CA)"
  value       = tls_self_signed_cert.vault.cert_pem
}

output "subnetwork_id" {
  description = "Subnetwork ID"
  value       = google_compute_subnetwork.main.id
}
