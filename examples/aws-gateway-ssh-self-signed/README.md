# AWS Gateway + SSH Resource (Self-Signed Certificates)

This example deploys a Twingate Gateway, Connector, and SSH server on AWS using
self-signed X.509 and SSH certificate authorities.

> **Warning:** This example generates private keys and certificates that are stored
> unencrypted in the Terraform state. Use a
> [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration)
> to protect sensitive state data.

## Prerequisites

- Terraform >= 1.4
- A Twingate account with an [API token](https://docs.twingate.com/docs/api-overview)
- An AWS account with credentials configured (`aws configure` or environment variables)

## Usage

```bash
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars

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

The instances have no public IP and no open SSH port. Open a shell on the
gateway instance over the private network path with an EIC endpoint, then
view the logs:

```bash
aws ec2-instance-connect ssh --instance-id "$(terraform output -raw gateway_instance_id)" --os-user admin
sudo journalctl -u gateway -f -o cat | jq -rR 'fromjson? // empty'
```

If you have not added an output for the instance ID, look it up by tag:

```bash
aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=demo-gateway" \
  --query "Reservations[].Instances[].InstanceId" --output text
```

## Clean up

```bash
terraform destroy
```
