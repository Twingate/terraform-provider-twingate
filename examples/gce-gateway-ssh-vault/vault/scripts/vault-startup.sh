#!/bin/bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

# --- Format and mount the persistent data disk ---
DATA_DISK="/dev/disk/by-id/google-${disk_name}"
MOUNT_POINT="/opt/vault/data"

if ! blkid "$DATA_DISK"; then
  mkfs.ext4 -m 0 -F -E lazy_itable_init=0,lazy_journal_init=0 "$DATA_DISK"
fi

mkdir -p "$MOUNT_POINT"
mount -o discard,defaults "$DATA_DISK" "$MOUNT_POINT"

# Persist mount across reboots
if ! grep -q "${disk_name}" /etc/fstab; then
  echo "$DATA_DISK $MOUNT_POINT ext4 discard,defaults,nofail 0 2" >> /etc/fstab
fi

# Detect stale data from a previous VM instance and wipe it.
# The persistent disk survives terraform destroy/apply cycles, but a new VM
# needs a fresh Vault initialization.
INSTANCE_ID=$(curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/id || echo "unknown")
INSTANCE_MARKER="$MOUNT_POINT/.vault-instance-id"

if [ -f "$INSTANCE_MARKER" ]; then
  OLD_ID=$(cat "$INSTANCE_MARKER")
  if [ "$OLD_ID" != "$INSTANCE_ID" ]; then
    echo "Detected stale Vault data from previous VM (old=$OLD_ID, new=$INSTANCE_ID). Wiping data..."
    find "$${MOUNT_POINT:?}" -mindepth 1 -delete
    rm -f /opt/vault/init-output.json
  fi
fi

echo "$INSTANCE_ID" > "$INSTANCE_MARKER"

# --- Install Vault ---
apt-get update -y
apt-get install -y curl gnupg lsb-release software-properties-common

curl -fsSL https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" \
  > /etc/apt/sources.list.d/hashicorp.list

apt-get update -y
apt-get install -y vault

# --- Write TLS certificates ---
mkdir -p /opt/vault/tls
cat > /opt/vault/tls/vault-cert.pem <<'CERT'
${vault_tls_cert}
CERT

cat > /opt/vault/tls/vault-key.pem <<'KEY'
${vault_tls_key}
KEY

chmod 600 /opt/vault/tls/vault-key.pem
chmod 644 /opt/vault/tls/vault-cert.pem
chown -R vault:vault /opt/vault

# --- Write Vault configuration ---
cat > /etc/vault.d/vault.hcl <<'EOF'
ui = true

storage "file" {
  path = "/opt/vault/data"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_cert_file = "/opt/vault/tls/vault-cert.pem"
  tls_key_file  = "/opt/vault/tls/vault-key.pem"
}

api_addr = "https://127.0.0.1:8200"
EOF

# --- Enable and start Vault ---
systemctl enable vault
systemctl start vault

# --- Wait for Vault to be ready and initialize ---
export VAULT_ADDR="https://127.0.0.1:8200"
export VAULT_SKIP_VERIFY=true
export PATH="/usr/bin:$PATH"

INIT_FILE="/opt/vault/init-output.json"

echo "Waiting for Vault to start..."
VAULT_READY=false
for i in $(seq 1 30); do
  # vault status exits 1 (error), 2 (sealed/uninit) — both mean it's responding
  STATUS=$(vault status -format=json 2>/dev/null) && true
  if echo "$STATUS" | grep -q '"initialized"'; then
    echo "Vault is responding."
    VAULT_READY=true
    break
  fi
  echo "  attempt $i: not ready yet..."
  sleep 2
done

if [ "$VAULT_READY" = false ]; then
  echo "ERROR: Vault did not become ready in time."
  exit 1
fi

# Check if Vault needs initialization
INITIALIZED=$(echo "$STATUS" | python3 -c "import sys,json; print(json.load(sys.stdin).get('initialized', False))" 2>/dev/null || echo "unknown")

if [ "$INITIALIZED" = "False" ]; then
  echo "Initializing Vault..."
  vault operator init \
    -key-shares=1 \
    -key-threshold=1 \
    -format=json > "$INIT_FILE"

  chmod 600 "$INIT_FILE"
  chown vault:vault "$INIT_FILE"

  echo "Vault initialized. Credentials saved to $INIT_FILE"

  # Unseal Vault using the stored key
  UNSEAL_KEY=$(python3 -c "import sys,json; print(json.load(sys.stdin)['unseal_keys_b64'][0])" < "$INIT_FILE")
  vault operator unseal "$UNSEAL_KEY"
  echo "Vault is unsealed and active."
else
  echo "Vault already initialized (initialized=$INITIALIZED)."

  # Check if Vault is sealed (e.g. after a reboot)
  SEALED=$(echo "$STATUS" | python3 -c "import sys,json; print(json.load(sys.stdin).get('sealed', True))" 2>/dev/null || echo "True")
  if [ "$SEALED" = "True" ]; then
    if [ -f "$INIT_FILE" ]; then
      echo "Vault is sealed, unsealing from stored key..."
      UNSEAL_KEY=$(python3 -c "import sys,json; print(json.load(sys.stdin)['unseal_keys_b64'][0])" < "$INIT_FILE")
      vault operator unseal "$UNSEAL_KEY"
    else
      echo "ERROR: Vault is sealed but no init file found at $INIT_FILE"
      exit 1
    fi
  fi
fi

echo "Vault startup complete."
