#!/bin/bash
set -e

GATEWAY_DIR="/etc/gateway"
# Check https://github.com/Twingate/gateway/releases for the latest version
BINARY_URL="https://github.com/Twingate/gateway/releases/download/v0.13.0/gateway_Linux_x86_64.tar.gz"


mkdir -p "$GATEWAY_DIR"

# Write TLS cert and key from templatefile variables
cat > "$GATEWAY_DIR/tls.crt" <<'CERT'
${tls-cert}
CERT

cat > "$GATEWAY_DIR/tls.key" <<'KEY'
${tls-key}
KEY

chmod 600 "$GATEWAY_DIR/tls.key"

# Write Vault CA cert
cat > /etc/ssl/vault-ca.crt <<'VAULTCA'
${vault_ca_cert}
VAULTCA

# Write config.yaml
cat > "$GATEWAY_DIR/config.yaml" <<'CONFIG'
${gateway-config}
CONFIG

# Download and extract the gateway binary
curl -sfL "$BINARY_URL" | tar xz -C "$GATEWAY_DIR"

# Create systemd service
cat > /etc/systemd/system/gateway.service <<EOF
[Unit]
Description=Twingate Gateway
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
