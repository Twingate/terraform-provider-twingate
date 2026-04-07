#!/bin/bash
set -e

GATEWAY_DIR="/opt/gateway"
# Check https://github.com/Twingate/gateway/releases for the latest version
BINARY_URL="https://github.com/Twingate/gateway/releases/download/v0.13.0/gateway_Linux_x86_64.tar.gz"

mkdir -p "$GATEWAY_DIR"
mkdir -p /etc/gateway

cat > /etc/gateway/tls.crt <<'CERT'
${tls_cert}
CERT

cat > /etc/gateway/tls.key <<'KEY'
${tls_key}
KEY

chmod 600 /etc/gateway/tls.key

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
