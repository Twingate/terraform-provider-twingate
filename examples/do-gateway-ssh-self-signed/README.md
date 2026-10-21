# DigitalOcean Gateway + SSH Resource (Self-Signed Certificates)

This example deploys a Twingate Gateway, Connector, and SSH server on DigitalOcean
using self-signed X.509 and SSH certificate authorities.

> **Warning:** This example generates private keys and certificates that are stored
> unencrypted in the Terraform state. Use a
> [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration)
> to protect sensitive state data.

## Prerequisites

- Terraform >= 1.4
- A Twingate account with an [API token](https://docs.twingate.com/docs/api-overview)
- A DigitalOcean account with an [API token](https://docs.digitalocean.com/reference/api/create-personal-access-token/)

## Usage

```bash
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars

terraform init
terraform apply
```

See `variables.tf` for the full list of inputs.

## Network access

A DigitalOcean cloud firewall (`network.tf`) allows inbound traffic to the Gateway,
Connector, and SSH server droplets only from inside the VPC. Outbound connections are
unrestricted. Components talk to each other over private IPs: the Connector reaches the
Gateway on its private address, and the Gateway reaches the SSH server on its private address.

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

With the Twingate Client connected, SSH into the Gateway droplet through the
`demo-gateway-ssh` Resource. Use the private IP shown on the Resource and the root
password. If it is the first time accessing the Droplet, you will need the password associated with the Droplet that DigitalOcean emails to the owner of the API key. You will then be prompted to change the password and log in again:

```bash
ssh root@<gateway-private-ip>
```

Then view the logs:

```bash
journalctl -u gateway -f -o cat
```

If the Connector failed to install, you won't be able to access the Droplet via Twingate. You can remove the firewall rule for SSH or access the Droplet through the recovery console. The recovery console can be accessed from the Droplet Overview (**Droplet Overview > Settings > Recovery console > Launch Console**). You will follow the same SSH process as above by logging in with the `root` user.

To access the Connector Droplet over Twingate, add a Resource for it in the Admin Console.

The break glass recovery console troubleshooting steps will work for any of the Droplets.

## Clean up

```bash
terraform destroy
```
