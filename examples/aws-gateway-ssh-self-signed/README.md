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

Each instance has a public IP for outbound traffic, but the security group only
allows inbound traffic between the instances themselves. To open a shell from the
AWS console, turn on `debug_ssh`, which allows port 22 from AWS's EC2 Instance
Connect IP range only:

```bash
terraform apply -var debug_ssh=true
```

In the AWS console, open **EC2 > Instances > demo-gateway > Connect > EC2 Instance
Connect**, choose **Connect using public IP**, and log in as `ubuntu`. Your IAM
identity needs the `ec2-instance-connect:SendSSHPublicKey` permission. No SSH key
pair is required.

On the gateway, view the logs:

```bash
sudo journalctl -u gateway -f -o cat | jq -rR 'fromjson? // empty'
```

When you're done, close port 22 again:

```bash
terraform apply -var debug_ssh=false
```

## Clean up

```bash
terraform destroy
```
