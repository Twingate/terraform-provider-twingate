# GCE Gateway + HTTPS Resource (Public CA Certificates)

This example deploys a Twingate Gateway, Connector, and [httpbin](https://github.com/mccutchen/go-httpbin)
test Web App on GCE, using a wildcard certificate issued by [Let's Encrypt](https://letsencrypt.org/).

Terraform requests the certificate with the [ACME provider](https://registry.terraform.io/providers/vancluever/acme/latest/docs)
and proves domain ownership with a DNS-01 challenge against Cloud DNS. The certificate chains to
a publicly trusted root, so clients do not need to install a CA certificate.

> **Warning:** This example stores the Gateway's private key and the ACME account key
> unencrypted in the Terraform state. Use a
> [remote backend with encryption](https://developer.hashicorp.com/terraform/language/settings/backends/configuration)
> to protect sensitive state data.

## Prerequisites

- Terraform >= 1.11
- A Twingate account with an [API token](https://docs.twingate.com/docs/api-overview)
- A GCP project with the Compute Engine and Cloud DNS APIs enabled
- A domain you control, delegated to a **public** Cloud DNS zone in that project
- `gcloud` CLI authenticated, with [Application Default Credentials](https://cloud.google.com/docs/authentication/provide-credentials-adc)
  for Terraform:

  ```bash
  gcloud auth login
  gcloud auth application-default login
  ```

  The ACME provider uses the same credentials to create the challenge TXT record, so they need
  permission to change records in the zone (for example, `roles/dns.admin`).

## Usage

```bash
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars

terraform init
terraform apply
```

Set `domain` to a name under your Cloud DNS zone. The certificate covers `*.<domain>`, and the
Twingate Resource alias is `httpbin.<domain>`. The alias does not need a public DNS record: the
Twingate Client resolves it. The only record written to the zone is the temporary
`_acme-challenge` TXT record.

See `variables.tf` for the full list of inputs.

## Access control

The Twingate Resource grants access to the `Everyone` group so the example works out of
the box. After you've confirmed it works, replace `data.twingate_groups.everyone` in
`twingate.tf` with the groups that should have access.

## Certificate authority

`isrg-root-x1.pem` is [ISRG Root X1](https://letsencrypt.org/certificates/), the Let's Encrypt
root the Gateway's certificate chains to. The example registers it with Twingate as the Gateway's
X.509 CA. Twingate needs the root rather than an intermediate, because Let's Encrypt can choose a
different intermediate for each certificate it issues.

## Renewal

Let's Encrypt certificates are valid for 90 days, and nothing renews them in the background. Run
`terraform apply` within 30 days of expiry (`min_days_remaining` in `twingate.tf`) to issue a new
certificate. The apply also replaces the Gateway VM.

## Accessing httpbin

With the Twingate Client connected, open:

```text
https://httpbin.<domain>/headers
```

Or test from the command line:

```bash
curl https://httpbin.<domain>/headers
```

httpbin echoes back the request headers it receives. Use it to confirm that the Gateway injects
the headers from `webApp.requestHeaders` in `config.yaml.tftpl` and from
`twingate_web_app_resource.request_header_rewrites` in `twingate.tf`.

## Troubleshooting

Every VM (`demo-https-public-ca-gateway`, `demo-https-public-ca-connector`, `demo-https-public-ca-httpbin`) is tagged `iap-ssh` and reachable via IAP, even though the VMs do not have an external IP:

```bash
gcloud compute ssh demo-https-public-ca-gateway --tunnel-through-iap
```

On the gateway, view the logs:

```bash
sudo journalctl -u gateway -f -o cat | jq -rR 'fromjson? // empty'
```

On httpbin, check that the container is running:

```bash
sudo docker ps --filter name=httpbin
```

If certificate issuance fails during `terraform apply`, check that `domain` is in a public zone
in `project_id` and that your credentials can create records in it. Let's Encrypt
[rate limits](https://letsencrypt.org/docs/rate-limits/) failed validations to 5 per hostname per
hour, so fix the cause before retrying.

## Clean up

```bash
terraform destroy
```

Destroying `acme_certificate` revokes the certificate with Let's Encrypt.
