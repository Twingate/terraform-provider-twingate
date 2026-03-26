provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_remote_network" "prod" {
  name = "Production Network"
}

resource "twingate_x509_certificate_authority" "tls" {
  name        = "My TLS CA"
  certificate = file("${path.module}/certs/ca.pem")
}

resource "twingate_ssh_certificate_authority" "ssh" {
  name       = "My SSH CA"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIr+Aj3O8csUrFRNWS7wViafil3rMlZ0glQ/OZ0CjCti"
}

resource "twingate_gateway" "main" {
  remote_network_id = twingate_remote_network.prod.id
  address           = "10.0.0.1:8001"
  x509_ca_id        = twingate_x509_certificate_authority.tls.id
  ssh_ca_id         = twingate_ssh_certificate_authority.ssh.id
}

# Kubernetes resource accessed via in-cluster DNS
resource "twingate_kubernetes_resource" "prod_cluster" {
  name              = "Production K8s"
  remote_network_id = twingate_remote_network.prod.id
  gateway_id        = twingate_gateway.main.id
}

resource "twingate_ssh_resource" "ssh_server" {
  name              = "SSH Server"
  gateway_id        = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
  address           = "10.128.0.105"
  username          = "ubuntu"
}

resource "twingate_ssh_resource" "ssh_server_2" {
  name              = "SSH Server 2"
  gateway_id        = twingate_gateway.main.id
  remote_network_id = twingate_remote_network.prod.id
  address           = "10.128.0.106"
  username          = "ubuntu-2"
}

resource "twingate_gateway_config" "config" {
  # Gateway listen port. Default: 8443.
  port = 8443

  # Prometheus metrics port. Default: 9090.
  metrics_port = 9090

  # TLS configuration for the gateway listener.
  # All fields have built-in defaults and can be omitted.
  tls = {
    certificate_file = "/etc/gateway/tls.crt"
    private_key_file = "/etc/gateway/tls.key"
  }

  ssh = {
    # SSH gateway process settings. All fields are optional with built-in defaults.
    gateway = {
      username      = "gateway"  # OS user the gateway process runs as. Default: "gateway".
      key_type      = "ed25519"  # SSH host key algorithm. Default: "ed25519".
      host_cert_ttl = "24h"      # Validity period for issued host certificates. Default: "24h".
      user_cert_ttl = "5m"       # Validity period for issued user certificates. Default: "5m".
    }

    ca = {
      # SSH CA backed by HashiCorp Vault (mutually exclusive with private_key_file).
      vault = {
        address = "https://vault.example.com"

        # Vault SSH secrets engine mount path. Default: "ssh".
        mount = "ssh"

        # Vault role used to sign certificates. Default: "gateway".
        role = "gateway"

        # Path to a custom CA bundle for verifying Vault's TLS certificate.
        # Default: "/etc/ssl/vault-ca.crt".
        ca_bundle_file = "/etc/ssl/vault-ca.crt"
      }

      # Vault authentication — choose one of: token or gcp.
      # auth = {
      #   # Option 1: static Vault token.
      #   token = "s.myVaultToken"
      # }

      auth = {
        # Option 2: GCP IAM / GCE authentication.
        gcp = {
          role  = "my-vault-gcp-role"
          type  = "iam"  # "iam" or "gce"
          mount = "gcp"  # Vault GCP auth mount path. Default: "gcp".

          # Required when type = "iam".
          service_account_email = "gateway-sa@my-project.iam.gserviceaccount.com"
        }
      }

      # Alternative: use a local private key file instead of Vault.
      # Mutually exclusive with vault.address.
      # private_key_file = "/etc/gateway/ssh_ca_key"
    }

    resources = [twingate_ssh_resource.ssh_server, twingate_ssh_resource.ssh_server_2]
  }

  kubernetes = {
    resources = [twingate_kubernetes_resource.prod_cluster]
  }
}

resource "local_sensitive_file" "config" {
  content  = twingate_gateway_config.config.content
  filename = "${path.module}/generated/config.yaml"
}
