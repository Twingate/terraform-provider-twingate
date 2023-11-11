# Connector AWS
This example demonstrates how to deploy Twingate connectors to Azure VMs.

## Pre-requisite
* Python and PIP
* Pulumi
* Azure CLI
* SSH Public Key

## How to Use
* Clone the repository
* `cd /path/to/repo/examples/connector-azure-vm`
* Configure Pulumi-Twingate Provider, see configuration section [here](../../README.md)
* Setup Azure CLI, see [here](https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli)
* `cp pulumi.dev.yaml.example pulumi.dev.yaml` and modify `pulumi.dev.yaml` to desired values including number of connectors to deploy.
* `pulumi up`

**Note**: `pulumi up` should automatically download the required Python dependency and Pulumi Plugins.

**Note**: make sure `dev` part in the file name of `pulumi.dev.yaml` is changed to the Pulumi stack name.

## How to Update Connectors
Auto connector update is enabled and is default to be performed at 02:00 AM server time. To modify the behaviour, see function `get_script` in [__main__.py](./__main__.py) 

**Note**: Connector update can cause the existing connection to be interrupted. 