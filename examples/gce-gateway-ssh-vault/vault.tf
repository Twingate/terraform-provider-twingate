# --- Vault authentication ---

resource "vault_auth_backend" "gcp" {
  type = "gcp"
}

resource "vault_gcp_auth_backend_role" "vm" {
  backend                = vault_auth_backend.gcp.path
  role                   = "vm-role"
  type                   = "gce"
  token_policies         = [vault_policy.ssh_server.name]
  token_ttl              = 600 # 10 minutes
  token_max_ttl          = 600 # 10 minutes
  bound_projects         = [var.project_id]
  bound_zones            = [var.zone]
  bound_service_accounts = [google_service_account.vm.email]
}

resource "vault_gcp_auth_backend_role" "gateway" {
  backend                = vault_auth_backend.gcp.path
  role                   = "gateway-role"
  type                   = "gce"
  token_policies         = [vault_policy.gateway.name]
  token_ttl              = 86400 # 24 hours
  token_max_ttl          = 86400 # 24 hours
  bound_projects         = [var.project_id]
  bound_zones            = [var.zone]
  bound_service_accounts = [google_service_account.gateway.email]
}

# --- Vault PKI backend (X.509 / TLS certs) ---

resource "vault_mount" "pki" {
  path                      = "pki"
  type                      = "pki"
  description               = "PKI backend for X.509 certificates"
  default_lease_ttl_seconds = 3600 # 1 hour
  max_lease_ttl_seconds     = 31536000 # 1 year
}

resource "vault_pki_secret_backend_root_cert" "root" {
  backend     = vault_mount.pki.path
  type        = "internal"
  common_name = "Demo Root CA"
  ttl         = "8760h" # 1 year
}

resource "vault_pki_secret_backend_role" "gateway" {
  backend        = vault_mount.pki.path
  name           = "gateway"
  allow_any_name = true
  key_usage      = ["DigitalSignature", "KeyEncipherment"]
  ext_key_usage  = ["ServerAuth"]
  generate_lease = true
}

# --- Vault SSH backend (SSH certs) ---

resource "vault_mount" "ssh" {
  path = "ssh"
  type = "ssh"
}

resource "vault_ssh_secret_backend_ca" "ssh" {
  backend              = vault_mount.ssh.path
  generate_signing_key = true
  key_type             = "ssh-ed25519" # smaller and faster than RSA
}

# Gateway needs to request both user and host certificates for SSH access.
resource "vault_ssh_secret_backend_role" "gateway" {
  name                    = "gateway"
  backend                 = vault_mount.ssh.path

  key_type                = "ca"
  ttl                     = "720h"  # 30 days
  max_ttl                 = "8760h" # 365 days

  allow_empty_principals  = true
  allow_host_certificates = true
  allow_user_certificates = true
  allowed_domains         = "*"
  allowed_users           = "gateway"
  allowed_extensions      = "permit-X11-forwarding,permit-agent-forwarding,permit-port-forwarding,permit-pty,permit-user-rc"
}

resource "vault_policy" "gateway" {
  name = "gateway-signing"

  policy = <<-EOT
    path "${vault_mount.ssh.path}/sign/${vault_ssh_secret_backend_role.gateway.name}" {
      capabilities = ["create", "update"]
    }
    path "${vault_mount.ssh.path}/config/ca" {
      capabilities = ["read"]
    }
  EOT
}

# SSH server only needs to request a host certificate.
resource "vault_ssh_secret_backend_role" "ssh_server" {
  name                    = "ssh-server"
  backend                 = vault_mount.ssh.path

  key_type                = "ca"
  ttl                     = "720h"  # 30 days
  max_ttl                 = "8760h" # 365 days

  allow_empty_principals  = true
  allow_host_certificates = true
  allowed_domains         = "*"
}

resource "vault_policy" "ssh_server" {
  name = "ssh-server-signing"

  policy = <<-EOT
    path "${vault_mount.ssh.path}/sign/${vault_ssh_secret_backend_role.ssh_server.name}" {
      capabilities = ["create", "update"]
    }
    path "${vault_mount.ssh.path}/config/ca" {
      capabilities = ["read"]
    }
  EOT
}
