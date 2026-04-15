#!/bin/bash
set -e

GATEWAY_DIR="/etc/gateway"
# Check https://github.com/Twingate/gateway/releases for the latest version
BINARY_URL="https://github.com/Twingate/gateway/releases/download/v0.13.0/gateway_Linux_x86_64.tar.gz"


mkdir -p "$GATEWAY_DIR"

# Write TLS cert, key, and SSH CA key from instance metadata
curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/tls-cert \
  > "$GATEWAY_DIR/tls.crt"

curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/tls-key \
  > "$GATEWAY_DIR/tls.key"

curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/ssh-ca-key \
  > "$GATEWAY_DIR/ssh-ca.key"

chmod 600 "$GATEWAY_DIR/tls.key" "$GATEWAY_DIR/ssh-ca.key"

# Write config.yaml from metadata
curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/gateway-config \
  > "$GATEWAY_DIR/config.yaml"

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
