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
