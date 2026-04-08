# GCE Gateway + SSH Resource (Vault-Backed Certificates)

This example deploys a Twingate Gateway, Connector, and SSH server on GCE using
HashiCorp Vault for SSH certificate signing and X.509 TLS certificates.

> **Warning:** The Vault root token is stored unencrypted in the Terraform state.
> Use a
> [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration)
> to protect sensitive state data.

## Prerequisites

- Terraform >= 1.4
- A Twingate account with an [API token](https://docs.twingate.com/docs/api-overview)
- A GCP project with the Compute Engine API enabled
- `gcloud` CLI authenticated

## Usage

This example is a two-step deploy: first the Vault infrastructure, then the root
module.

### 1. Deploy the Vault infrastructure

This demo uses a minimal Vault configuration optimized for simplicity, not
production hardening:

- **File storage backend** on a GCE persistent disk — single-node, no replication
- **Single unseal key** (`key-shares=1, key-threshold=1`) for convenience
- **Auto-init and auto-unseal** from a plaintext key file stored on the VM itself

> **Production considerations:** A production Vault deployment would need:
> - **KMS-based auto-unseal** (e.g., GCP Cloud KMS) instead of a plaintext key file on disk
> - **Multiple key shares** with a higher threshold (e.g., 5 shares, 3 threshold)
> - **Raft integrated storage** or a Consul backend for HA and replication
> - **Unseal keys stored off-VM** in a secure, separate system

```bash
cd vault
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars

terraform init
terraform apply
```

After Vault starts (~60 s), SSH in to get the root token:

```bash
gcloud compute ssh demo-vault-server --zone us-central1-a --tunnel-through-iap -- \
  "sudo cat /opt/vault/init-output.json"
```

### 2. Deploy the root module

In a separate terminal, start an IAP tunnel to Vault:

```bash
gcloud compute start-iap-tunnel demo-vault-server 8200 \
  --local-host-port=localhost:8200 --zone=us-central1-a
```

Then deploy:

```bash
cd ..   # back to gce-gateway-ssh-vault/
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars (add vault_token from step 1)

terraform init
terraform apply
```

See `variables.tf` for the full list of inputs.

## Resource alias

By default, users connect to the SSH server by its internal IP:

```bash
ssh <internal-ip>
```

To use a hostname instead, set `resource_alias`:

```hcl
resource_alias = "ssh-server.int"
```

This adds the alias as a DNS SAN on the Gateway's TLS certificate and sets it as
the resource alias in the Twingate Client. Users can then connect with:

```bash
ssh ssh-server.int
```

## Troubleshooting

SSH into the gateway VM via IAP (the VM has no external IP), then view the logs:

```bash
sudo journalctl -u gateway -f -o cat | jq -rR 'fromjson? // empty'
```

## Clean up

```bash
terraform destroy
cd vault && terraform destroy
```
