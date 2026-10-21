# DigitalOcean Gateway + HTTPS Resource (Self-Signed Certificates)

This example deploys a Twingate Gateway, Connector, and [httpbin](https://github.com/mccutchen/go-httpbin)
test Web App on DigitalOcean, using a self-signed X.509 certificate authority.

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
Connector, and httpbin Droplets only from inside the VPC. Outbound connections are
unrestricted. Components talk to each other over private IPs. The Connector reaches the
Gateway on its private address, and the Gateway reaches httpbin on its private address.

## Trust the CA certificate

The Gateway serves a certificate signed by the self-signed CA in this example.
`terraform apply` writes the CA certificate to `do_example_https_ca.crt` in this directory. Add it to
your machine's trust store so your browser accepts the Gateway's certificate.

macOS:

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain do_example_https_ca.crt
```

Debian/Ubuntu:

```bash
sudo cp do_example_https_ca.crt /usr/local/share/ca-certificates/twingate-gateway-ca.crt
sudo update-ca-certificates
```

## Accessing httpbin

With the Twingate Client connected, open:

```text
https://do-nyc-http-bin.demo.int/headers
```

httpbin echoes back the request headers it receives, so you can confirm the
headers configured under `webApp.requestHeaders` in `config.yaml.tftpl` and in `twingate.tf` under `twingate_web_app_resource.request_header_rewrites` are injected by the Gateway.

If you would like to access the Resource under a different hostname, set `resource_alias` in the `terraform.tfvars` file:

```hcl
resource_alias = "httpbin.example.int"
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

To access the Connector or httpbin Droplets over Twingate, add a Resource for them in the Admin Console.

The break glass recovery console troubleshooting steps will work for any of the Droplets.

## Clean up

```bash
terraform destroy
```

Then remove the CA certificate from your trust store.

macOS:

```bash
sudo security delete-certificate -c "DO HTTPS Example Twingate Gateway CA" /Library/Keychains/System.keychain
```

Debian/Ubuntu:

```bash
sudo rm /usr/local/share/ca-certificates/twingate-gateway-ca.crt
sudo update-ca-certificates --fresh
```
