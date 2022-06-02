---
subcategory: "gke"
page_title: "GKE Helm Provider Deployment Guide"
description: |-
This document walks you through a basic deployment using Twingate's Terraform provider on GKE using the Helm Terraform provider
---

# Deployment Guide

This deployment guide walks you through a Twingate connector helm deployment in a GKE cluster.

## Before you begin

* Sign up for an account on the [Twingate website](https://www.twingate.com). You will need the Twingate Enterprise tier to use Terraform with Twingate.
* Create a Twingate [API key](https://docs.twingate.com/docs/api-overview). The key will need to have full permissions to Read, Write, & Provision, in order to deploy Connectors through Terraform.

## Setting up the Provider

First, we need to set up the Twingate Terraform provider by providing your network ID and the API key you provisioned earlier.

```terraform
provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "autoco"
}

variable "network" {
  default = "autoco"
}
```


## Provider requirements

provider versions are excluded from the example below

```terraform
terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
    }
    google = {
      source  = "hashicorp/google"
    }
    twingate = {
      source  = "twingate/twingate"
    }
    docker = {
      source  = "kreuzwerker/docker"
    }
  }
}
```

## Provider setup

```terraform

provider "google" {
  project     = var.project_id
  region      = var.region
  zone        = var.zone
}

data "google_client_config" "provider" {}

data "google_container_cluster" "cluster" {
  name     = "you-cluster"
  location = var.cluster_location
}

provider "helm" {
  kubernetes {
    host  = "https://${data.google_container_cluster.cluster.endpoint}"
    token = data.google_client_config.provider.access_token
    cluster_ca_certificate = base64decode(
      data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate
    )
  }
}

data "docker_registry_image" "connector" {
  name = "twingate/connector:1"
}

```
## Creating the Remote Network and Connectors in Twingate

Next, we'll create the objects in Twingate that correspond to the GCP network that we're deploying Twingate into
A Remote Network to represent the GKE subnet, and a Connector to be deployed in that subnet. We'll use these objects when we're deploying the Connector helm chart.

```terraform
resource "twingate_remote_network" "gcp_network" {
  name = "GCP Network"
}

resource "twingate_connector" "gke_connector" {
  remote_network_id = twingate_remote_network.gcp_network.id
}

resource "twingate_connector_tokens" "gke_connector_tokens" {
  connector_id = twingate_connector.gke_connector.id
}
```

## Deploying the Connector

Now that we have the data types created in Twingate, we need to deploy a Connector into the GKE cluster to handle Twingate traffic.

```terraform
resource "helm_release" "connector" {
  chart            = "connector"
  name             = "twingate-connector"
  repository       = "https://twingate.github.io/helm-charts"
  namespace        = "twingate"
  create_namespace = true
  recreate_pods    = true

  set {
    name  = "connector.network"
    value = var.network
  }

  # connector image updates are not tied to helm chart updates, so in order to keep the connector up to date we are using its image sha256 as helm property
  # every time a new version of the connector is pushed and the terraform build runs, the connector will be updated and restarted
  set {
    name  = "sha256"
    value = data.docker_registry_image.connector.sha256_digest
  }

  set {
    name  = "icmpSupport.enabled"
    value = true
  }

  set {
    name  = "connector.accessToken"
    value = twingate_connector_tokens.gke_connector_tokens.access_token
  }

  set {
    name  = "connector.refreshToken"
    value = twingate_connector_tokens.gke_connector_tokens.refresh_token
  }

}
```

