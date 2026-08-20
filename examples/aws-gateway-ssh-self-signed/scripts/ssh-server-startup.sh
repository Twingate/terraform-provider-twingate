#!/bin/bash
set -e

# Install the SSM Agent so operators can open a shell over the private AWS
# network path (Session Manager) without a public IP or an open SSH port.
ARCH="$(dpkg --print-architecture)"
curl -sfL "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/debian_$${ARCH}/amazon-ssm-agent.deb" -o /tmp/amazon-ssm-agent.deb
sudo dpkg -i /tmp/amazon-ssm-agent.deb
sudo systemctl enable --now amazon-ssm-agent

useradd -m -s /bin/bash gateway

cat > /etc/ssh/twingate-ca.pub <<'PUBKEY'
${ssh_ca_public_key}
PUBKEY

echo "TrustedUserCAKeys /etc/ssh/twingate-ca.pub" >> /etc/ssh/sshd_config

sudo systemctl restart sshd
