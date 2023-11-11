# Twingate Resource Provider
This is a non-official Pulumi resource provider for Twingate that was built using the pulumi terraform bridge utilizing 
the official [Twingate Terraform Provider](https://registry.terraform.io/providers/Twingate/twingate/latest).

## Configuration

The following configuration points are available for the `twingate` provider:

- `twingate:apiToken` - The access key for API operations. You can retrieve this from the Twingate Admin Console
  ([documentation](https://docs.twingate.com/docs/api-overview)). Alternatively, this can be specified using the
  TWINGATE_API_TOKEN environment variable.
- `twingate:network` - Your Twingate network ID for API operations. You can find it in the Admin Console URL, for example:
  `autoco.twingate.com`, where `autoco` is your network ID. Alternatively, this can be specified using the TWINGATE_NETWORK
  environment variable.
- `twingate:url` - The default is 'twingate.com'. This is optional and shouldn't be changed under normal circumstances.

## Examples
* [TypeScript](./examples/ts): Demonstrating how Twingate remote network, service account, service key and resources can be created and configured in Typescript.
* [Python](./examples/python): Demonstrating how Twingate remote network, service account, service key and resources can be created and configured in Python.
* [AWS EC2 Connector](./examples/connector-aws-ec2): Deploying Twingate connectors to AWS EC2 instances.
* [AWS ECS Connector](./examples/connector-aws-ecs): Deploying Twingate connectors to AWS ECS cluster.
* [GCP VM Connector](./examples/connector-gcp-instance): Deploying Twingate connectors to GCP instances.
* [GKE Connectors](./examples/connector-gcp-gke): Deploying Twingate connectors to GKE Kubernetes cluster.
* [Azure VM Connectors ](./examples/connector-azure-vm): Deploying Twingate connectors to Azure VM instances.
* [Azure Container Connectors](./examples/connector-azure-container): Deploying Twingate connectors to Azure Container Instance.