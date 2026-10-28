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

## Access control

Both Twingate Resources grant access to the `Everyone` group so the example works out of the box. After you've confirmed it works, replace `data.twingate_groups.everyone` in `twingate.tf` with the groups that should have access. The `demo-gateway-ssh` Resource is for debugging, so limit it to administrators:

```hcl
data "twingate_groups" "admins" {
  name = "<your admin group>"
}

resource "twingate_resource" "gateway_ssh" {
  # ...
  access_group {
    group_id = data.twingate_groups.admins.groups[0].id
  }
}
```

## Trust the CA certificate

The Gateway serves a certificate signed by the self-signed CA in this example.
`terraform apply` writes the CA certificate to `do_example_https_ca.crt` in this directory. Add it to
your machine's trust store so your browser accepts the Gateway's certificate. Without it, you will continue to see unsafe connection warnings even when accessing the Resource over HTTPS.

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

Or test from the command line:

```bash
curl https://do-nyc-http-bin.demo.int/headers
```

httpbin echoes back the request headers it receives, so you can confirm the
headers configured under `webApp.requestHeaders` in `config.yaml.tftpl` and in `twingate.tf` under `twingate_web_app_resource.request_header_rewrites` are injected by the Gateway.

Changing the `resource_alias` variable in the `terraform.tfvars` file allows you to define what hostname you would like to access the Resource from. This example uses `do-nyc-http-bin.demo.int` but any hostname following the `host.domain.tld` pattern will work. Another example would be the following:

```hcl
resource_alias = "httpbin.example.int"
```

This adds the alias as a DNS SAN on the Gateway's TLS certificate and sets it as
the Resource alias in the Twingate Client. The alias is the certificate's only SAN,
so changing it reissues the certificate and replaces the Gateway Droplet. Wildcards can
be used to support multiple aliases and avoid the overhead related to reissuing the certificate. See the `tls_cert_request` block in `twingate.tf` for an example.

## CA rotation

The CA certificate is valid for one year (`validity_period_hours` in `twingate.tf`).
To rotate it before it expires, or if its private key is compromised:

1. Replace the CA private key. Terraform generates a new CA certificate, reissues the
   server certificate, registers the new CA with Twingate, points the Gateway at it,
   deletes the old CA, replaces the Gateway Droplet, and overwrites
   `do_example_https_ca.crt`:

   ```bash
   terraform apply -replace=tls_private_key.x509_ca
   ```

2. Remove the old CA certificate from your trust store, then add the new one.

   macOS:

   ```bash
   sudo security delete-certificate -c "DO HTTPS Example Twingate Gateway CA" /Library/Keychains/System.keychain
   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain do_example_https_ca.crt
   ```

   Debian/Ubuntu:

   ```bash
   sudo cp do_example_https_ca.crt /usr/local/share/ca-certificates/twingate-gateway-ca.crt
   sudo update-ca-certificates --fresh
   ```

3. Redistribute the new `do_example_https_ca.crt` to every other client that trusts
   the old CA. Until a client trusts the new CA, it will see certificate warnings.

## Troubleshooting

With the Twingate Client connected, SSH into the Gateway Droplet through the
`demo-gateway-ssh` Resource. Use the private IP shown on the Resource and the root
password. If it is the first time accessing the Droplet, you will need the password associated with the Droplet that DigitalOcean emails to the owner of the API key. You will then be prompted to change the password and log in again:

```bash
ssh root@<gateway-private-ip>
```

Then view the logs:

```bash
journalctl -u gateway -f -o cat
```

If the Connector failed to install, you won't be able to access the Droplet via Twingate. You have two options:

- Temporarily add an inbound rule to the `demo-https-vpc-only` firewall that allows TCP port 22 from your public IP only, then SSH to the Droplet's public IP. Remove the rule when you're done.
- Access the Droplet through the recovery console from the Droplet Overview (**Droplet Overview > Settings > Recovery console > Launch Console**).

Either way, log in with the `root` user as described above.

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
