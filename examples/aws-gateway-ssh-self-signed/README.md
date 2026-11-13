# AWS Gateway + SSH Resource (Self-Signed Certificates)

This example deploys a Twingate Gateway, Connector, and SSH server on AWS using
self-signed X.509 and SSH certificate authorities.

> **Warning:** This example generates private keys and certificates that are stored
> unencrypted in the Terraform state. Use a
> [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration)
> to protect sensitive state data.

## Prerequisites

- Terraform >= 1.11
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

## Network access

The Gateway, Connector, and SSH server instances run in a private subnet (`network.tf`)
with no public IPs and no ports open to the internet. Their outbound traffic goes through
a NAT gateway. A security group allows inbound TCP traffic between the instances, plus SSH
(port 22) from an EC2 Instance Connect Endpoint for troubleshooting.

Everything talks over private IPs: the Connector reaches the Gateway, and the Gateway
reaches the SSH server, on their private addresses.

## Access control

The SSH Resource grants access to the `Everyone` group so the example works out of the box. After you've confirmed it works, replace `data.twingate_groups.everyone` in `twingate.tf` with the groups that should have access.

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

SSH into any instance as `ubuntu` through the EC2 Instance Connect Endpoint, either
from the EC2 console by selecting the instance from the instance list and navigating to the web console (**Connect > EC2 Instance Connect Endpoint**) or with the AWS CLI (v2.12+).
Set `INSTANCE` to `gateway`, `connector` or `ssh_server`:

```bash
INSTANCE=gateway
aws ec2-instance-connect ssh \
  --instance-id "$(terraform output -json instance_ids | jq -r ".$INSTANCE")" \
  --os-user ubuntu \
  --connection-type eice
```

AWS authorizes both methods with your AWS credentials and pushes a short-lived SSH key to
the instance, so you don't need a key pair.

On the Gateway, view the logs:

```bash
sudo journalctl -u gateway -f -o cat | jq -rR 'fromjson? // empty'
```

## Clean up

```bash
terraform destroy
```
