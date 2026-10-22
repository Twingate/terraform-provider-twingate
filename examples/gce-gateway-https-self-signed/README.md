# GCE Gateway + HTTPS Resource (Self-Signed Certificates)

This example deploys a Twingate Gateway, Connector, and [httpbin](https://github.com/mccutchen/go-httpbin)
test Web App on GCE, using a self-signed X.509 certificate authority.

> **Warning:** This example generates private keys and certificates that are stored
> unencrypted in the Terraform state. Use a
> [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration)
> to protect sensitive state data.

## Prerequisites

- Terraform >= 1.4
- A Twingate account with an [API token](https://docs.twingate.com/docs/api-overview)
- A GCP project with the Compute Engine API enabled
- `gcloud` CLI authenticated

## Usage

```bash
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars

terraform init
terraform apply
```

See `variables.tf` for the full list of inputs.

## Network access

None of the VMs have an external IP. Cloud NAT gives them outbound internet access,
and a firewall rule (`network.tf`) allows inbound TCP only from inside the subnet. The
Connector reaches the Gateway on its internal address, and the Gateway reaches httpbin
on its internal address.

## Access control

The Twingate Resource grants access to the `Everyone` group so the example works out of
the box. After you've confirmed it works, replace `data.twingate_groups.everyone` in
`twingate.tf` with the groups that should have access.

## Trust the CA certificate

The Gateway serves a certificate signed by the self-signed CA in this example.
`terraform apply` writes the CA certificate to `gce_example_https_ca.crt` in this directory. Add it to
your machine's trust store so your browser accepts the Gateway's certificate.

macOS:

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain gce_example_https_ca.crt
```

Debian/Ubuntu:

```bash
sudo cp gce_example_https_ca.crt /usr/local/share/ca-certificates/twingate-gateway-ca.crt
sudo update-ca-certificates
```

## Accessing httpbin

With the Twingate Client connected, open:

```text
https://gce-http-bin.demo.int/headers
```

httpbin echoes back the request headers it receives, so you can confirm the
headers configured under `webApp.requestHeaders` in `config.yaml.tftpl` and in `twingate.tf` under `twingate_web_app_resource.request_header_rewrites` are injected by the Gateway.

If you would like to access the Resource under a different hostname, set `resource_alias` in the `terraform.tfvars` file:

```hcl
resource_alias = "httpbin.example.int"
```

## Troubleshooting

Every VM (`demo-https-gateway`, `demo-https-connector`, `demo-httpbin`) is tagged `iap-ssh` and reachable via IAP, even though none has an external IP:

```bash
gcloud compute ssh demo-https-gateway --tunnel-through-iap
```

On the gateway, view the logs:

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
```

Then remove the CA certificate from your trust store.

macOS:

```bash
sudo security delete-certificate -c "GCE HTTPS Example Twingate Gateway CA" /Library/Keychains/System.keychain
```

Debian/Ubuntu:

```bash
sudo rm /usr/local/share/ca-certificates/twingate-gateway-ca.crt
sudo update-ca-certificates --fresh
```
