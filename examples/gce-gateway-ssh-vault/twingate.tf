resource "twingate_remote_network" "main" {
  name = "demo-vault-ssh-network"
}

resource "twingate_connector" "main" {
  remote_network_id = twingate_remote_network.main.id
  name              = "demo-vault-connector"
}

resource "twingate_connector_tokens" "main" {
  connector_id = twingate_connector.main.id
}

resource "twingate_ssh_certificate_authority" "vault" {
  name       = "demo-vault-ssh-ca"
  public_key = vault_ssh_secret_backend_ca.ssh.public_key
}

resource "twingate_x509_certificate_authority" "vault" {
  name        = "demo-vault-x509-ca"
  certificate = vault_pki_secret_backend_root_cert.root.certificate
}

resource "twingate_gateway" "vault" {
  remote_network_id = twingate_remote_network.main.id
  address           = "${google_compute_address.gateway.address}:${local.gateway_port}"
  x509_ca_id        = twingate_x509_certificate_authority.vault.id
  ssh_ca_id         = twingate_ssh_certificate_authority.vault.id
}

data "twingate_groups" "everyone" {
  name = "Everyone"
}

resource "twingate_ssh_resource" "ssh_server" {
  name              = "demo-vault-ssh-server"
  address           = google_compute_instance.ssh_server.network_interface[0].network_ip
  alias             = var.resource_alias != "" ? var.resource_alias : null
  remote_network_id = twingate_remote_network.main.id
  gateway_id        = twingate_gateway.vault.id

  access_group {
    group_id = data.twingate_groups.everyone.groups[0].id
  }
}
