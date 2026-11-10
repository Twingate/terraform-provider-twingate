# AWS Gateway + HTTPS Resource (Self-Signed Certificates)

This example deploys a Twingate Gateway, Connector, and [httpbin](https://github.com/mccutchen/go-httpbin)
test Web App on AWS, using a self-signed X.509 certificate authority.

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

The Gateway, Connector, and httpbin instances run in a private subnet (`network.tf`)
with no public IPs and no ports open to the internet. Their outbound traffic goes through
a NAT gateway. A security group allows inbound TCP traffic between the instances, plus SSH
(port 22) from an EC2 Instance Connect Endpoint for troubleshooting.

Everything talks over private IPs: the Connector reaches the Gateway, and the Gateway
reaches httpbin, on their private addresses.

## Access control

The Twingate Resource grants access to the `Everyone` group so the example works out of
the box. After you've confirmed it works, replace `data.twingate_groups.everyone` in
`twingate.tf` with the groups that should have access.

## Trust the CA certificate

The Gateway serves a certificate signed by the self-signed CA in this example.
`terraform apply` writes the CA certificate to `aws_example_https_ca.crt` in this directory. Add it to
your machine's trust store so your browser accepts the Gateway's certificate. Without it, you will continue to see unsafe connection warnings even when accessing the Resource over HTTPS.

macOS:

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain aws_example_https_ca.crt
```

Debian/Ubuntu:

```bash
sudo cp aws_example_https_ca.crt /usr/local/share/ca-certificates/twingate-gateway-ca.crt
sudo update-ca-certificates
```

## Accessing httpbin

With the Twingate Client connected, open:

```text
https://aws-http-bin.demo.int/headers
```

Or test from the command line:

```bash
curl https://aws-http-bin.demo.int/headers
```

httpbin echoes back the request headers it receives, so you can confirm the
headers configured under `webApp.requestHeaders` in `config.yaml.tftpl` and in `twingate.tf` under `twingate_web_app_resource.request_header_rewrites` are injected by the Gateway.

Changing the `resource_alias` variable in the `terraform.tfvars` file allows you to define what hostname you would like to access the Resource from. This example uses `aws-http-bin.demo.int` but any hostname following the `host.domain.tld` pattern will work. Another example would be the following:

```hcl
resource_alias = "httpbin.example.int"
```

This adds the alias as a DNS SAN on the Gateway's TLS certificate and sets it as
the Resource alias in the Twingate Client. The alias is the certificate's only SAN,
so changing it reissues the certificate and replaces the Gateway instance. Wildcards can
be used to support multiple aliases and avoid the overhead related to reissuing the certificate. See the `tls_cert_request` block in `twingate.tf` for an example.

## CA rotation

The CA certificate is valid for one year (`validity_period_hours` in `twingate.tf`).
To rotate it before it expires, or if its private key is compromised:

1. Replace the CA private key. Terraform generates a new CA certificate, reissues the
   server certificate, updates the Twingate X.509 CA, replaces the Gateway instance, and
   overwrites `aws_example_https_ca.crt`:

   ```bash
   terraform apply -replace=tls_private_key.x509_ca
   ```

2. Remove the old CA certificate from your trust store, then add the new one.

   macOS:

   ```bash
   sudo security delete-certificate -c "AWS HTTPS Example Twingate Gateway CA" /Library/Keychains/System.keychain
   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain aws_example_https_ca.crt
   ```

   Debian/Ubuntu:

   ```bash
   sudo cp aws_example_https_ca.crt /usr/local/share/ca-certificates/twingate-gateway-ca.crt
   sudo update-ca-certificates --fresh
   ```

3. Redistribute the new `aws_example_https_ca.crt` to every other client that trusts
   the old CA. Until a client trusts the new CA, it will see certificate warnings.

## Troubleshooting

SSH into the Gateway instance as `ubuntu` through the EC2 Instance Connect Endpoint, either
from the EC2 console (**Instances > demo-https-gateway > Connect > EC2 Instance Connect Endpoint**) or with the AWS CLI (v2.12+):

```bash
aws ec2-instance-connect ssh \
  --instance-id "$(terraform output -json instance_ids | jq -r .gateway)" \
  --os-user ubuntu \
  --connection-type eice
```

`terraform output instance_ids` lists the IDs of all three instances.

AWS authorizes both methods with your AWS credentials and pushes a short-lived SSH key to
the instance, so you don't need a key pair.

On the Gateway, view the logs:

```bash
sudo journalctl -u gateway -f -o cat | jq -rR 'fromjson? // empty'
```

On httpbin, check that the container is running:

```bash
sudo docker ps --filter name=httpbin
```

The Connector and httpbin instances are reachable the same way.

## Clean up

```bash
terraform destroy
```

Then remove the CA certificate from your trust store.

macOS:

```bash
sudo security delete-certificate -c "AWS HTTPS Example Twingate Gateway CA" /Library/Keychains/System.keychain
```

Debian/Ubuntu:

```bash
sudo rm /usr/local/share/ca-certificates/twingate-gateway-ca.crt
sudo update-ca-certificates --fresh
```
