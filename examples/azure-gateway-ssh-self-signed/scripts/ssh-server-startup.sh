#!/bin/bash
set -e

# Create the gateway user account
useradd -m -s /bin/bash gateway

# Write the SSH CA public key
cat > /etc/ssh/twingate-ca.pub <<'PUBKEY'
${ssh_ca_public_key}
PUBKEY

# Only accept the "gateway" principal, and only for the "gateway" account
mkdir -p /etc/ssh/auth_principals
echo "gateway" > /etc/ssh/auth_principals/gateway

# Configure sshd to trust certificates signed by our CA
cat > /etc/ssh/sshd_config.d/00-twingate.conf <<'SSHD'
TrustedUserCAKeys /etc/ssh/twingate-ca.pub
AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u
SSHD

systemctl restart ssh
