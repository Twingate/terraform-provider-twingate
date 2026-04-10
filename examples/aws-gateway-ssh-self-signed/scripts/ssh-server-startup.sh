#!/bin/bash
set -e

useradd -m -s /bin/bash gateway

cat > /etc/ssh/twingate-ca.pub <<'PUBKEY'
${ssh_ca_public_key}
PUBKEY

echo "TrustedUserCAKeys /etc/ssh/twingate-ca.pub" >> /etc/ssh/sshd_config

systemctl restart sshd
