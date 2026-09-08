#!/bin/bash
set -e

# Create the gateway user account
useradd -m -s /bin/bash gateway

# Get the SSH CA public key from instance metadata
CA_KEY=$(curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/ssh-ca-public-key)

# Configure sshd to trust certificates signed by our CA
echo "$CA_KEY" > /etc/ssh/twingate-ca.pub
echo "TrustedUserCAKeys /etc/ssh/twingate-ca.pub" >> /etc/ssh/sshd_config

# Only accept the "gateway" principal, and only for the "gateway" account
mkdir -p /etc/ssh/auth_principals
echo "gateway" > /etc/ssh/auth_principals/gateway
echo "AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u" >> /etc/ssh/sshd_config

systemctl restart sshd
