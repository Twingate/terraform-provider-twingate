#!/bin/bash
set -e

# Install the SSM Agent so operators can open a shell over the private AWS
# network path (Session Manager) without a public IP or an open SSH port.
ARCH="$(dpkg --print-architecture)"
curl -sfL "https://s3.amazonaws.com/ec2-downloads-windows/SSMAgent/latest/debian_$${ARCH}/amazon-ssm-agent.deb" -o /tmp/amazon-ssm-agent.deb
sudo dpkg -i /tmp/amazon-ssm-agent.deb
sudo systemctl enable --now amazon-ssm-agent

# Check https://github.com/Twingate/gateway/releases for the latest version
BINARY_URL="https://github.com/Twingate/gateway/releases/download/v0.13.0/gateway_Linux_x86_64.tar.gz"
GATEWAY_DIR="/etc/gateway"

mkdir -p "$GATEWAY_DIR"

cat > "$GATEWAY_DIR/tls.crt" <<'CERT'
${tls_cert}
CERT

cat > "$GATEWAY_DIR/tls.key" <<'KEY'
${tls_key}
KEY

chmod 600 "$GATEWAY_DIR/tls.key"

cat > "$GATEWAY_DIR/ssh-ca.key" <<'SSHKEY'
${ssh_ca_key}
SSHKEY

chmod 600 "$GATEWAY_DIR/ssh-ca.key"

cat > "$GATEWAY_DIR/config.yaml" <<'CONFIG'
${gateway_config}
CONFIG

curl -sfL "$BINARY_URL" | tar xz -C "$GATEWAY_DIR"

cat > /etc/systemd/system/gateway.service <<EOF
[Unit]
Description=Twingate Access Gateway
After=network.target

[Service]
ExecStart=$GATEWAY_DIR/gateway start --config $GATEWAY_DIR/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now gateway
