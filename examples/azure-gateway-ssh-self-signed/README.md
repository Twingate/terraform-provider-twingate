# Azure Gateway + SSH Resource (Self-Signed Certificates)

This example deploys a Twingate Gateway, Connector, and SSH server on Azure using
self-signed X.509 and SSH certificate authorities.

> **Warning:** This example generates private keys and certificates that are stored
> unencrypted in the Terraform state. Use a
> [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration)
> to protect sensitive state data.

## Prerequisites

- Terraform >= 1.11
- A Twingate account with an [API token](https://docs.twingate.com/docs/api-overview)
- An Azure subscription
- `az` CLI authenticated (`az login`)

## Usage

```bash
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars

terraform init
terraform apply
```

See `variables.tf` for the full list of inputs. `location` must offer the Azure Bastion
Developer SKU (the default, `eastus2`, does).

## Network access

The Gateway, Connector, and SSH server VMs run in a single subnet (`network.tf`) with no
public IPs and no ports open to the internet. Their outbound traffic goes through a NAT
gateway. The network security group relies on Azure's default rules to allow traffic
inside the VNet and deny inbound traffic from the internet, plus one rule that allows SSH
(port 22) from Azure Bastion for troubleshooting.

Everything talks over private IPs: the Connector reaches the Gateway, and the Gateway
reaches the SSH server, on their private addresses.

## Access control

The SSH Resource grants access to the `Everyone` group so the example works out of the
box. After you've confirmed it works, replace `data.twingate_groups.everyone` in
`twingate.tf` with the groups that should have access.

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

SSH into any VM (`demo-gateway`, `demo-connector`, `demo-ssh-server`) as `azureadmin`
through the Azure Bastion Developer SKU. First save the admin private key:

```bash
terraform output -raw vm_admin_private_key > azureadmin.pem
chmod 600 azureadmin.pem
```

Then in the Azure portal, open the VM and go to **Connect > Bastion**. Choose
**SSH Private Key from Local File**, enter `azureadmin` as the username, and select
`azureadmin.pem`. The Developer SKU only supports connections from the portal, not
`az network bastion ssh`.

On the Gateway, view the logs:

```bash
sudo journalctl -u gateway -f -o cat | jq -rR 'fromjson? // empty'
```

## Clean up

```bash
terraform destroy
rm -f azureadmin.pem
```
