# Connector GKE
This example demonstrates how to deploy Twingate connectors to GKE Kubernetes Cluster.

## Pre-requisite
* Python and PIP
* Pulumi
* GCP CLI

## How to Use
* Clone the repository
* `cd /path/to/repo/examples/connector-gcp-gke`
* Configure Pulumi-Twingate Provider, see configuration section [here](../../README.md)
* Setup GCP CLI, see [here](https://cloud.google.com/sdk/docs/install-sdk#initializing_the)
* `cp pulumi.dev.yaml.example pulumi.dev.yaml` and modify `pulumi.dev.yaml` to desired values including number of connectors to deploy.
* `pulumi up`

**Note**: `pulumi up` should automatically download the required Python dependency and Pulumi Plugins.

**Note**: make sure `dev` part in the file name of `pulumi.dev.yaml` is changed to the Pulumi stack name.

## How to Update Connectors
Adding parameter `recreate_pods=True,` to `ReleaseArgs` in [__main__.py](./__main__.py) and execute `pulumi up` would trigger the connector pod to be replaced. This can be used to update pods to use a newer Docker image.

**Note**: Connector update can cause the existing connection to be interrupted. 