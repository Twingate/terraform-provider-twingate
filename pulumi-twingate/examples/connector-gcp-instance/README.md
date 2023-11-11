# Connector GKE
This example demonstrates how to deploy Twingate connectors to GCP instances.

## Pre-requisite
* Python and PIP
* Pulumi
* GCP CLI

## How to Use
* Clone the repository
* `cd /path/to/repo/examples/connector-gcp-instance`
* Configure Pulumi-Twingate Provider, see configuration section [here](../../README.md)
* Setup GCP CLI, see [here](https://cloud.google.com/sdk/docs/install-sdk#initializing_the)
* `cp pulumi.dev.yaml.example pulumi.dev.yaml` and modify `pulumi.dev.yaml` to desired values including number of connectors to deploy.
* `pulumi up`

**Note**: `pulumi up` should automatically download the required Python dependency and Pulumi Plugins.

**Note**: make sure `dev` part in the file name of `pulumi.dev.yaml` is changed to the Pulumi stack name.

## How to Update Connectors
Auto connector update is enabled and is default to be performed at 02:00 AM server time. To modify the behaviour, see variable `start_script` in [__main__.py](./__main__.py)

**Note**: Connector update can cause the existing connection to be interrupted. 