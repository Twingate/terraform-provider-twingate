import pulumi
import pulumi_azure_native as azure
import pulumi_twingate as tg
import os
import base64

config = pulumi.Config()
data = config.require_object("data")
twingate_config = pulumi.Config("twingate")

# Set to True to enable SSH to the connector VM instance
ssh_enabled = False

try:
    tg_account = twingate_config.get("network")
    if tg_account is None:
        tg_account = os.getenv('TWINGATE_NETWORK')
except:
    tg_account = os.getenv('TWINGATE_NETWORK')

# Create a resource group
resource_group = azure.resources.ResourceGroup(data.get("resource_group_name"),
                                               resource_group_name=data.get("resource_group_name"))

# Create a virtual network
virtual_network = azure.network.VirtualNetwork(
    data.get("vnet_name"),
    virtual_network_name=data.get("vnet_name"),
    resource_group_name=resource_group.name,
    address_space=azure.network.AddressSpaceArgs(
        address_prefixes=[
            data.get("vnet_cidr"),
        ],
    )
)

# Create a public IP
public_ip_address = azure.network.PublicIPAddress(data.get("pub_ip_name"),
                                                  public_ip_address_name=data.get("pub_ip_name"),
                                                  resource_group_name=resource_group.name,
                                                  public_ip_address_version="IPv4",
                                                  public_ip_allocation_method="Static",
                                                  sku=azure.network.PublicIPAddressSkuArgs(
                                                      name="Standard",
                                                      tier="Regional",
                                                  ))

# Create a NAT gateway
nat_gateway = azure.network.NatGateway(data.get("nat_name"),
                                       nat_gateway_name=data.get("nat_name"),
                                       resource_group_name=resource_group.name,
                                       public_ip_addresses=[azure.network.SubResourceArgs(
                                           id=public_ip_address.id,
                                       )],
                                       sku=azure.network.NatGatewaySkuArgs(
                                           name="Standard"))

# Create a subnet
subnet = azure.network.Subnet(data.get("subnet_name"),
                              resource_group_name=resource_group.name,
                              virtual_network_name=virtual_network.name,
                              address_prefixes=[data.get("subnet_cidr")],
                              nat_gateway=azure.network.SubResourceArgs(id=nat_gateway.id))



# Enable ssh if ssh_enable is true
security_group_extra_args = {}
if ssh_enabled:
    security_group_extra_args["security_rules"] = [
        azure.network.SecurityRuleArgs(
            name="enable_ssh",
            priority=1000,
            direction=azure.network.AccessRuleDirection.INBOUND,
            access="Allow",
            protocol="Tcp",
            source_port_range="*",
            source_address_prefix="*",
            destination_address_prefix="*",
            destination_port_ranges=[
                "22",
            ],
        ),
    ]

# Create a security group
security_group = azure.network.NetworkSecurityGroup(
    data.get("security_group_name"),
    network_security_group_name=data.get("security_group_name"),
    resource_group_name=resource_group.name,
    **security_group_extra_args
)


remote_network = tg.TwingateRemoteNetwork(data.get("tg_remote_network"), name=data.get("tg_remote_network"))


# Setup init script
def get_script(v):
    script = f'''#!/bin/sh
curl "https://binaries.twingate.com/connector/setup.sh" | \
sudo TWINGATE_ACCESS_TOKEN="{v[0]}" \
TWINGATE_REFRESH_TOKEN="{v[1]}" \
TWINGATE_URL="https://{tg_account}.twingate.com" bash; \
sudo bash -c 'echo TWINGATE_LABEL_DEPLOYEDBY="tg-pulumi-azure-vm" >> /etc/twingate/connector.conf'; \
sudo bash -c 'echo "
Unattended-Upgrade::Origins-Pattern {{
    \\"site=packages.twingate.com\\";
    }};" >> /etc/apt/apt.conf.d/50unattended-upgrades'; \
    sudo bash -c 'echo "
Unattended-Upgrade::Automatic-Reboot-Time \\"02:00\\";" >> /etc/apt/apt.conf.d/50unattended-upgrades'; \
    sudo systemctl daemon-reload; \
    sudo systemctl enable twingate-connector.service; \
    sudo service twingate-connector restart'''
    return base64.b64encode(bytes(script, "utf-8")).decode("utf-8")


# Read local ssh key. To create ssh key, see
# https://learn.microsoft.com/en-us/azure/virtual-machines/linux/mac-create-ssh-keys#create-an-ssh-key-pair
with open(data.get("ssh_key_local_path")) as f:
    ssh_public_key = f.read().replace("\n", "")

pulumi.export("test", ssh_public_key)

connectors = data.get("connectors")

# Create a Network Interface and an VM For Each Connector
for i in range(1, connectors + 1):
    # Create a network interface with the virtual network, IP address, and security group.
    network_interface = azure.network.NetworkInterface(
        f"{data.get('network_interface_name')}-{i}",
        network_interface_name=f"{data.get('network_interface_name')}-{i}",
        resource_group_name=resource_group.name,
        network_security_group=azure.network.NetworkSecurityGroupArgs(
            id=security_group.id,
        ),
        ip_configurations=[
            azure.network.NetworkInterfaceIPConfigurationArgs(
                name=f'{data.get("network_interface_name")}-ipconfiguration',
                private_ip_allocation_method=azure.network.IpAllocationMethod.DYNAMIC,
                subnet=azure.network.SubnetArgs(
                    id=subnet.id,
                ),
            ),
        ],
    )

    connector = tg.TwingateConnector(f"twingate_connector_{i}", name="", remote_network_id=remote_network.id)
    connector_token = tg.TwingateConnectorTokens(f"connector_token_{i}", connector_id=connector.id)

    init_script = pulumi.Output.all(connector_token.access_token, connector_token.refresh_token).apply(
        lambda v: get_script(v))

    virtual_machine = azure.compute.VirtualMachine(f"twingate-connector-{i}",
                                                   vm_name=pulumi.Output.all(connector.name).apply(
                                                       lambda v: f"tg-{v[0]}"),
                                                   resource_group_name=resource_group.name,
                                                   hardware_profile=azure.compute.HardwareProfileArgs(
                                                       vm_size=data.get("vm_size"),
                                                   ),
                                                   network_profile=azure.compute.NetworkProfileArgs(
                                                       network_interfaces=[azure.compute.NetworkInterfaceReferenceArgs(
                                                           id=network_interface.id,
                                                           primary=True,
                                                       )],
                                                   ),
                                                   os_profile=azure.compute.OSProfileArgs(
                                                       computer_name=connector.name,
                                                       admin_username="ubuntu",
                                                       custom_data=init_script,
                                                       linux_configuration=azure.compute.LinuxConfigurationArgs(
                                                           disable_password_authentication=True,
                                                           ssh=azure.compute.SshConfigurationArgs(
                                                               public_keys=[
                                                                   azure.compute.SshPublicKeyArgs(
                                                                       key_data=ssh_public_key,
                                                                       path="/home/ubuntu/.ssh/authorized_keys"
                                                                   ),
                                                               ],
                                                           )
                                                       )
                                                   ),
                                                   storage_profile=azure.compute.StorageProfileArgs(
                                                       image_reference=azure.compute.ImageReferenceArgs(
                                                           offer="0001-com-ubuntu-server-jammy",
                                                           publisher="Canonical",
                                                           sku="22_04-lts-gen2",
                                                           version="latest",
                                                       ),
                                                   )
                                                   )
