#!/bin/bash
set -e

# Create the gateway user account
useradd -m -s /bin/bash gateway

# Write the SSH CA public key
cat > /etc/ssh/twingate-ca.pub <<'PUBKEY'
${ssh_ca_public_key}
PUBKEY

# Configure sshd to trust certificates signed by our CA
echo "TrustedUserCAKeys /etc/ssh/twingate-ca.pub" >> /etc/ssh/sshd_config

# Only accept the "gateway" principal, and only for the "gateway" account
mkdir -p /etc/ssh/auth_principals
echo "gateway" > /etc/ssh/auth_principals/gateway
echo "AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u" >> /etc/ssh/sshd_config

systemctl restart sshd
