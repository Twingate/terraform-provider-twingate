# Connector AWS
This example demonstrates how to deploy Twingate connectors to Azure Container Instance.

## Pre-requisite
* Python and PIP
* Pulumi
* Azure CLI
* SSH Public Key

## How to Use
* Clone the repository
* `cd /path/to/repo/examples/connector-azure-container`
* Configure Pulumi-Twingate Provider, see configuration section [here](../../README.md)
* Setup Azure CLI, see [here](https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli)
* `cp pulumi.dev.yaml.example pulumi.dev.yaml` and modify `pulumi.dev.yaml` to desired values including number of connectors to deploy.
* `pulumi up`

**Note**: `pulumi up` should automatically download the required Python dependency and Pulumi Plugins.

**Note**: make sure `dev` part in the file name of `pulumi.dev.yaml` is changed to the Pulumi stack name.

## How to Update Connectors
Modify line `image="twingate/connector:{version number}"` in [__main__.py](./__main__.py) and execute `pulumi up` would trigger connector task definitions to be replaced. This would replace the connector image with the defined version.

**Note**: Connector update can cause the existing connection to be interrupted. 