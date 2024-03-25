import pulumi
import pulumi_azure_native as azure_native
import pulumi_twingate as tg
import os

config = pulumi.Config()
data = config.require_object("data")
twingate_config = pulumi.Config("twingate")

try:
    tg_account = twingate_config.get("network")
    if tg_account is None:
        tg_account = os.getenv('TWINGATE_NETWORK')
except:
    tg_account = os.getenv('TWINGATE_NETWORK')

# Create a resource group
resource_group = azure_native.resources.ResourceGroup(data.get("resource_group_name"),
                                                      resource_group_name=data.get("resource_group_name"))

# Create a virtual network
virtual_network = azure_native.network.VirtualNetwork(
    data.get("vnet_name"),
    virtual_network_name=data.get("vnet_name"),
    resource_group_name=resource_group.name,
    address_space=azure_native.network.AddressSpaceArgs(
        address_prefixes=[
            data.get("vnet_cidr"),
        ],
    )
)

# Create a public IP
public_ip_address = azure_native.network.PublicIPAddress(data.get("pub_ip_name"),
                                                         public_ip_address_name=data.get("pub_ip_name"),
                                                         resource_group_name=resource_group.name,
                                                         public_ip_address_version="IPv4",
                                                         public_ip_allocation_method="Static",
                                                         sku=azure_native.network.PublicIPAddressSkuArgs(
                                                             name="Standard",
                                                             tier="Regional",
                                                         ))

# Create a NAT gateway
nat_gateway = azure_native.network.NatGateway(data.get("nat_name"),
                                              nat_gateway_name=data.get("nat_name"),
                                              resource_group_name=resource_group.name,
                                              public_ip_addresses=[azure_native.network.SubResourceArgs(
                                                  id=public_ip_address.id,
                                              )],
                                              sku=azure_native.network.NatGatewaySkuArgs(
                                                  name="Standard"))

# Create a subnet
subnet = azure_native.network.Subnet(data.get("subnet_name"),
                                     resource_group_name=resource_group.name,
                                     virtual_network_name=virtual_network.name,
                                     address_prefixes=[data.get("subnet_cidr")],
                                     delegations=[azure_native.network.DelegationArgs(
                                         name="delegation",
                                         service_name="Microsoft.ContainerInstance/containerGroups",
                                         type="Microsoft.Network/virtualNetworks/subnets/delegations",
                                     )],
                                     nat_gateway=azure_native.network.SubResourceArgs(id=nat_gateway.id))

# Create a network profile
network_profile = azure_native.network.NetworkProfile(data.get("network_profile_name"),
                                                      container_network_interface_configurations=[{
                                                          "ipConfigurations": [{
                                                              "name": "ipconfig1",
                                                              "subnet": azure_native.network.SubnetArgs(
                                                                  id=subnet.id,
                                                              ),
                                                          }],
                                                          "name": data.get("network_interface_name"),
                                                      }],
                                                      network_profile_name=data.get("network_profile_name"),
                                                      resource_group_name=resource_group.name)

remote_network = tg.TwingateRemoteNetwork(data.get("tg_remote_network"), name=data.get("tg_remote_network"))

connectors = data.get("connectors")

# Create container group for each connector
for i in range(1, connectors + 1):
    connector = tg.TwingateConnector(f"twingate_connector_{i}", name="", remote_network_id=remote_network.id)
    connector_token = tg.TwingateConnectorTokens(f"connector_token_{i}", connector_id=connector.id)
    container_group = azure_native.containerinstance.ContainerGroup(f"Twingate-Connector-{i}",
                                                                    container_group_name=connector.name,
                                                                    containers=[{
                                                                        "command": [],
                                                                        "environmentVariables": [
                                                                            {"name": "TENANT_URL", "value": f"https://{tg_account}.twingate.com"},
                                                                            {"name": "ACCESS_TOKEN", "value": connector_token.access_token},
                                                                            {"name": "REFRESH_TOKEN", "value": connector_token.refresh_token},
                                                                            {"name": "TWINGATE_LABEL_DEPLOYEDBY", "value": "tg-pulumi-azure-container"}
                                                                        ],
                                                                        "image": "twingate/connector:1",
                                                                        "name": connector.name,
                                                                        "ports": [
                                                                            azure_native.containerinstance.ContainerPortArgs(
                                                                                port=80,
                                                                            )],
                                                                        "resources": {
                                                                            "requests": azure_native.containerinstance.ResourceRequestsArgs(
                                                                                cpu=data.get("container_cpu"),
                                                                                memory_in_gb=data.get("container_memory"),
                                                                            ),
                                                                        }
                                                                    }],
                                                                    ip_address=azure_native.containerinstance.IpAddressArgs(
                                                                        type="Private",
                                                                        ports=[{
                                                                            "port": 80,
                                                                            "protocol": "TCP",
                                                                        }],
                                                                    ),
                                                                    network_profile=azure_native.containerinstance.ContainerGroupNetworkProfileArgs(
                                                                        id=network_profile.id,
                                                                    ),
                                                                    os_type="Linux",
                                                                    resource_group_name=resource_group.name,
                                                                    )
