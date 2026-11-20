# Azure Gateway + HTTPS Resource (Self-Signed Certificates)

This example deploys a Twingate Gateway, Connector, and [httpbin](https://github.com/mccutchen/go-httpbin)
test Web App on Azure, using a self-signed X.509 certificate authority.

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

The Gateway, Connector, and httpbin VMs run in a single subnet (`network.tf`) with no
public IPs and no ports open to the internet. Their outbound traffic goes through a NAT
gateway. The network security group relies on Azure's default rules to allow traffic
inside the VNet and deny inbound traffic from the internet, plus one rule that allows SSH
(port 22) from Azure Bastion for troubleshooting.

Everything talks over private IPs: the Connector reaches the Gateway, and the Gateway
reaches httpbin, on their private addresses.

## Access control

The Twingate Resource grants access to the `Everyone` group so the example works out of
the box. After you've confirmed it works, replace `data.twingate_groups.everyone` in
`twingate.tf` with the groups that should have access.

## Trust the CA certificate

The Gateway serves a certificate signed by the self-signed CA in this example.
`terraform apply` writes the CA certificate to `azure_example_https_ca.crt` in this directory. Add it to
your machine's trust store so your browser accepts the Gateway's certificate. Without it, you will continue to see unsafe connection warnings even when accessing the Resource over HTTPS.

macOS:

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain azure_example_https_ca.crt
```

Debian/Ubuntu:

```bash
sudo cp azure_example_https_ca.crt /usr/local/share/ca-certificates/twingate-gateway-ca.crt
sudo update-ca-certificates
```

## Accessing httpbin

With the Twingate Client connected, open:

```text
https://azure-http-bin.demo.int/headers
```

Or test from the command line:

```bash
curl https://azure-http-bin.demo.int/headers
```

httpbin echoes back the request headers it receives, so you can confirm the
headers configured under `webApp.requestHeaders` in `config.yaml.tftpl` and in `twingate.tf` under `twingate_web_app_resource.request_header_rewrites` are injected by the Gateway.

Changing the `resource_alias` variable in the `terraform.tfvars` file allows you to define what hostname you would like to access the Resource from. This example uses `azure-http-bin.demo.int` but any hostname following the `host.domain.tld` pattern will work. Another example would be the following:

```hcl
resource_alias = "httpbin.example.int"
```

This adds the alias as a DNS SAN on the Gateway's TLS certificate and sets it as
the Resource alias in the Twingate Client. The alias is the certificate's only SAN,
so changing it reissues the certificate and replaces the Gateway VM. Wildcards can
be used to support multiple aliases and avoid the overhead related to reissuing the certificate. See the `tls_cert_request` block in `twingate.tf` for an example.

## CA rotation

The CA certificate is valid for one year (`validity_period_hours` in `twingate.tf`).
To rotate it before it expires, or if its private key is compromised:

1. Replace the CA private key. Terraform generates a new CA certificate, reissues the
   server certificate, updates the Twingate X.509 CA, replaces the Gateway VM, and
   overwrites `azure_example_https_ca.crt`:

   ```bash
   terraform apply -replace=tls_private_key.x509_ca
   ```

2. Remove the old CA certificate from your trust store, then add the new one.

   macOS:

   ```bash
   sudo security delete-certificate -c "Azure HTTPS Example Twingate Gateway CA" /Library/Keychains/System.keychain
   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain azure_example_https_ca.crt
   ```

   Debian/Ubuntu:

   ```bash
   sudo cp azure_example_https_ca.crt /usr/local/share/ca-certificates/twingate-gateway-ca.crt
   sudo update-ca-certificates --fresh
   ```

3. Redistribute the new `azure_example_https_ca.crt` to every other client that trusts
   the old CA. Until a client trusts the new CA, it will see certificate warnings.

## Troubleshooting

SSH into any VM (`demo-gateway`, `demo-connector`, `demo-httpbin`) as `azureadmin`
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

On httpbin, check that the container is running:

```bash
sudo docker ps --filter name=httpbin
```

## Clean up

```bash
terraform destroy
rm -f azureadmin.pem
```

Then remove the CA certificate from your trust store.

macOS:

```bash
sudo security delete-certificate -c "Azure HTTPS Example Twingate Gateway CA" /Library/Keychains/System.keychain
```

Debian/Ubuntu:

```bash
sudo rm /usr/local/share/ca-certificates/twingate-gateway-ca.crt
sudo update-ca-certificates --fresh
```
