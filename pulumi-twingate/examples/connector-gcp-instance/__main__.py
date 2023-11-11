import pulumi
from pulumi_gcp import compute
import pulumi_twingate as tg
import os

config = pulumi.Config()
data = config.require_object("data")
twingate_config = pulumi.Config("twingate")

# Set to True to enable ssh to the connector VM instance
enable_ssh = False

try:
    if twingate_config.get("network"):
        tg_account = twingate_config.get("network")
    else:
        tg_account = os.getenv('TWINGATE_NETWORK')
except:
    tg_account = os.getenv('TWINGATE_NETWORK')

# Create an VPC
vpc = compute.Network(
    data.get("vpc_name"),
    name=data.get("vpc_name"),
    auto_create_subnetworks=False,
)

# Create a Subnet
subnet = compute.Subnetwork(data.get("subnet_name"),
                            name=data.get("subnet_name"),
                            ip_cidr_range=data.get("subnet_cidr"),
                            network=vpc.id,
                            )

# Create a Cloud Router
router = compute.Router(data.get("router_name"),
                        name=data.get("router_name"),
                        network=vpc.id
                        )
# Create a Cloud NAT
nat = compute.RouterNat(data.get("nat_name"),
                        name=data.get("nat_name"),
                        router=router.name,
                        nat_ip_allocate_option="AUTO_ONLY",
                        source_subnetwork_ip_ranges_to_nat="ALL_SUBNETWORKS_ALL_IP_RANGES",
                        log_config=compute.RouterNatLogConfigArgs(
                            enable=True,
                            filter="ERRORS_ONLY",
                        ))

if enable_ssh:
    compute_firewall = compute.Firewall(
        "firewall",
        name="enable-ssh",
        network=vpc.self_link,
        allows=[compute.FirewallAllowArgs(
            protocol="tcp",
            ports=["22"],
        )],
        source_ranges=["0.0.0.0/0"]
    )

# Create a Twingate Remote Network
remote_network = tg.TwingateRemoteNetwork(data.get("tg_remote_network"), name=data.get("tg_remote_network"))

connectors = data.get("connectors")

# Create a VM Instance For Each Connector
for i in range(1, connectors + 1):
    connector = tg.TwingateConnector(f"connector_{i}", name="", remote_network_id=remote_network.id)
    connector_token = tg.TwingateConnectorTokens(f"token_{i}", connector_id=connector.id)
    start_script = pulumi.Output.all(connector_token.access_token, connector_token.refresh_token).apply(lambda v: f'''\
    curl "https://binaries.twingate.com/connector/setup.sh" | \
    sudo TWINGATE_ACCESS_TOKEN="{v[0]}" \
    TWINGATE_REFRESH_TOKEN="{v[1]}" \
    TWINGATE_URL="https://{tg_account}.twingate.com" bash; \
    sudo bash -c 'echo TWINGATE_LABEL_DEPLOYEDBY="tg-pulumi-gcp-vm" >> /etc/twingate/connector.conf'; \
    sudo bash -c 'echo "
Unattended-Upgrade::Origins-Pattern {{
    \\"site=packages.twingate.com\\";
    }};" >> /etc/apt/apt.conf.d/50unattended-upgrades'; \
    sudo bash -c 'echo "
Unattended-Upgrade::Automatic-Reboot-Time \\"02:00\\";" >> /etc/apt/apt.conf.d/50unattended-upgrades'; \
    sudo systemctl daemon-reload; \
    sudo systemctl enable twingate-connector.service; \
    sudo service twingate-connector restart''')

    vm = compute.Instance(f"twingate-connector-{i}",
                          name=pulumi.Output.all(connector.name).apply(
                              lambda v: f"tg-{v[0]}"),
                          machine_type=data.get("vm_type"),
                          boot_disk=compute.InstanceBootDiskArgs(
                              initialize_params=compute.InstanceBootDiskInitializeParamsArgs(
                                  image="ubuntu-os-cloud/ubuntu-2204-lts",
                              ),
                          ),
                          network_interfaces=[compute.InstanceNetworkInterfaceArgs(
                              network=vpc.id,
                              subnetwork=subnet.id,
                          )],
                          metadata_startup_script=start_script,
                          )
