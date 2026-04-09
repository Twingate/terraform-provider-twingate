#!/bin/bash
set -euo pipefail

apt-get update -qq && apt-get install -y -qq jq

# Create the gateway user
useradd -m -s /bin/bash gateway

# Write the SSH CA public key from templatefile variable
cat > /etc/ssh/vault-ssh-ca.pub <<'SSHCA'
${ssh-ca-public-key}
SSHCA

# Write the Vault CA cert for TLS verification
cat > /etc/ssl/vault-ca.crt <<'VAULTCA'
${vault_ca_cert}
VAULTCA

# Authenticate to Vault via GCP auth
JWT=$(curl -sf -H "Metadata-Flavor: Google" \
  "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=vault/${vault_gcp_role}&format=full")

cat > /tmp/login-payload.json <<EOF
{
  "role": "${vault_gcp_role}",
  "jwt": "$JWT"
}
EOF

LOGIN_RESPONSE=$(curl -s --fail-with-body \
  --cacert /etc/ssl/vault-ca.crt \
  -X POST \
  --data @/tmp/login-payload.json \
  "${vault_addr}/v1/auth/${vault_gcp_mount}/login")

VAULT_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.auth.client_token') || {
  echo "ERROR: Failed to authenticate to Vault via GCP auth"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
}

# Get this VM's internal IP from GCP metadata
MY_IP=$(curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip)

# Sign the SSH server's host key with Vault SSH CA
HOST_PUB_KEY=$(cat /etc/ssh/ssh_host_ed25519_key.pub)

cat > /tmp/sign-payload.json <<EOF
{
  "public_key": "$HOST_PUB_KEY",
  "cert_type": "host",
  "valid_principals": "$MY_IP",
  "ttl": "8760h"
}
EOF

RESPONSE=$(curl -s --fail-with-body \
  --cacert /etc/ssl/vault-ca.crt \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -X POST \
  --data @/tmp/sign-payload.json \
  "${vault_addr}/v1/${vault_mount}/sign/${vault_role}")

SIGNED_CERT=$(echo "$RESPONSE" | jq -r '.data.signed_key') || {
  echo "ERROR: Failed to get signed host certificate from Vault"
  echo "Response: $RESPONSE"
  exit 1
}

echo "$SIGNED_CERT" > /etc/ssh/ssh_host_ed25519_key-cert.pub
chmod 640 /etc/ssh/ssh_host_ed25519_key-cert.pub

# Configure sshd to trust CA certs and present host certificate
cat >> /etc/ssh/sshd_config <<EOF
TrustedUserCAKeys /etc/ssh/vault-ssh-ca.pub
HostCertificate /etc/ssh/ssh_host_ed25519_key-cert.pub
EOF

systemctl restart sshd
echo "SSH server configured with Vault-signed host certificate"
