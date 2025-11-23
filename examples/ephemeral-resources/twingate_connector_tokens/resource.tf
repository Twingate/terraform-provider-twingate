provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_remote_network" "aws_network" {
  name = "test-aws_remote_network"
}

resource "twingate_connector" "aws_connector" {
  remote_network_id = twingate_remote_network.aws_network.id
}

ephemeral "twingate_connector_tokens" "aws_connector_tokens" {
  connector_id = twingate_connector.aws_connector.id
}

## Google Secret Manager to store the token
provider "google" {
  project = "my-project-id"
  region  = "us-central1"
}

resource "google_secret_manager_secret" "twingate_token" {
  secret_id = "twingate-token"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "twingate_secret" {
  secret                 = google_secret_manager_secret.twingate_token.id
  secret_data_wo_version = 1
  secret_data_wo         = ephemeral.twingate_connector_tokens.aws_connector_tokens.access_token
}

data "google_secret_manager_secret_version" "twingate_secret" {
  secret  = google_secret_manager_secret.twingate_token.id
  version = google_secret_manager_secret_version.twingate_secret.version
}

## AWS Secret Manager to store the token
provider "aws" {
  region = "us-east-1"
}

resource "aws_secretsmanager_secret" "twingate_token" {
  name = "twingate-token"
}

resource "aws_secretsmanager_secret_version" "twingate_secret" {
  secret_id                = aws_secretsmanager_secret.twingate_token.id
  secret_string_wo_version = 1
  secret_string_wo         = ephemeral.twingate_connector_tokens.aws_connector_tokens.access_token
}

data "aws_secretsmanager_secret_version" "twingate_secret" {
  secret_id  = aws_secretsmanager_secret.twingate_token.id
  version_id = aws_secretsmanager_secret_version.twingate_secret.version_id
}

# Set the secret values to local variables for use
locals {
  gcp_access_token = data.google_secret_manager_secret_version.twingate_secret.secret_data
  aws_access_token = data.aws_secretsmanager_secret_version.twingate_secret.secret_string
}
