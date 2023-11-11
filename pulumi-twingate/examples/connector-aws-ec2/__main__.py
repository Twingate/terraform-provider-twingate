import pulumi
import pulumi_aws as aws
import pulumi_twingate as tg
import os

config = pulumi.Config()
data = config.require_object("data")
twingate_config = pulumi.Config("twingate")

# Set to True to enable SSH to the connector EC2 instance
ssh_enabled = False


try:
    tg_account = twingate_config.get("network")
    if tg_account is None:
        tg_account = os.getenv('TWINGATE_NETWORK')
except:
    tg_account = os.getenv('TWINGATE_NETWORK')

# Create a VPC
vpc = aws.ec2.Vpc(
    data.get("vpc_name"),
    cidr_block=data.get("vpc_cidr"),
    enable_dns_hostnames=True,
    tags={
        "Name": data.get("vpc_name"),
    }
)

# Create a Private Subnet
private_subnet = aws.ec2.Subnet(data.get("prv_subnet_name"),
                                vpc_id=vpc.id,
                                cidr_block=data.get("prv_cidr"),
                                map_public_ip_on_launch=False,
                                tags={
                                    "Name": data.get("prv_subnet_name"),
                                })

# Create a Public Subnet
public_subnet = aws.ec2.Subnet(data.get("pub_subnet_name"),
                               vpc_id=vpc.id,
                               cidr_block=data.get("pub_cidr"),
                               map_public_ip_on_launch=True,
                               tags={
                                   "Name": data.get("pub_subnet_name"),
                               })

# Create an Elastic IP
eip = aws.ec2.Eip(data.get("eip_name"),
                  vpc=True)

# Create an Internet Gateway
igw = aws.ec2.InternetGateway(data.get("igw_name"),
                              vpc_id=vpc.id,
                              tags={
                                  "Name": data.get("igw_name"),
                              })

# Create a NatGateway
nat_gateway = aws.ec2.NatGateway(data.get("natgw_name"),
                                 allocation_id=eip.allocation_id,
                                 subnet_id=public_subnet.id,
                                 tags={
                                     "Name": data.get("natgw_name"),
                                 },
                                 opts=pulumi.ResourceOptions(depends_on=[igw]))

# Create a Public Route Table
pub_route_table = aws.ec2.RouteTable(data.get("pubrttable_name"),
                                     vpc_id=vpc.id,
                                     routes=[
                                         aws.ec2.RouteTableRouteArgs(
                                             cidr_block="0.0.0.0/0",
                                             gateway_id=igw.id,
                                         )
                                     ],
                                     tags={
                                         "Name": data.get("pubrttable_name"),
                                     })

# Create a private Route Table
prv_route_table = aws.ec2.RouteTable(data.get("prvrttable_name"),
                                     vpc_id=vpc.id,
                                     routes=[
                                         aws.ec2.RouteTableRouteArgs(
                                             cidr_block="0.0.0.0/0",
                                             gateway_id=nat_gateway.id,
                                         )
                                     ],
                                     tags={
                                         "Name": data.get("prvrttable_name"),
                                     })

# Create a Public Route Association
pub_route_association = aws.ec2.RouteTableAssociation(
    data.get("pubrtasst_name"),
    route_table_id=pub_route_table.id,
    subnet_id=public_subnet.id
)

# Create a Private Route Association
prv_route_association = aws.ec2.RouteTableAssociation(
    data.get("prvrtasst_name"),
    route_table_id=prv_route_table.id,
    subnet_id=private_subnet.id
)

# Enable ssh if ssh_enable is true
sg_extra_args = {}
if ssh_enabled:
    sg_extra_args["ingress"] = [
        {
            "protocol": "tcp",
            "from_port": 22,
            "to_port": 22,
            "cidr_blocks": ["0.0.0.0/0"],

        }
    ]

# Create a Security Group
sg = aws.ec2.SecurityGroup(
    data.get("sec_grp_name"),
    egress=[
        {
            "protocol": "-1",
            "from_port": 0,
            "to_port": 0,
            "cidr_blocks": ["0.0.0.0/0"],
        }
    ],
    vpc_id=vpc.id,
    **sg_extra_args
)

# Get the Key Pair, Can Also Create New one
keypair = aws.ec2.get_key_pair(key_name=data.get("key_name"), include_public_key=True)

# Getting Twingate Connector AMI
ami = aws.ec2.get_ami(most_recent=True,
                      owners=["617935088040"],
                      filters=[{"name": "name", "values": ["twingate/images/hvm-ssd/twingate-amd64-*"]}])

# Create a Twingate remote network
remote_network = tg.TwingateRemoteNetwork(data.get("tg_remote_network"), name=data.get("tg_remote_network"))

connectors = data.get("connectors")

# Create a EC2 Instance For Each Connector
for i in range(1, connectors + 1):
    connector = tg.TwingateConnector(f"twingate_connector_{i}", name="", remote_network_id=remote_network.id)
    connector_token = tg.TwingateConnectorTokens(f"connector_token_{i}", connector_id=connector.id)
    user_data = pulumi.Output.all(connector_token.access_token, connector_token.refresh_token).apply(lambda v: f'''#!/bin/bash
    sudo mkdir -p /etc/twingate/
    HOSTNAME_LOOKUP=$(curl http://169.254.169.254/latest/meta-data/local-hostname)
    EGRESS_IP=$(curl https://checkip.amazonaws.com)
    {{
    echo TWINGATE_URL="https://{tg_account}.twingate.com"
    echo TWINGATE_ACCESS_TOKEN="{v[0]}"
    echo TWINGATE_REFRESH_TOKEN="{v[1]}"
    echo TWINGATE_LOG_ANALYTICS=v1
    echo TWINGATE_LABEL_HOSTNAME=$HOSTNAME_LOOKUP
    echo TWINGATE_LABEL_EGRESSIP=$EGRESS_IP
    echo TWINGATE_LABEL_DEPLOYEDBY=tg-pulumi-aws-ec2
    }} > /etc/twingate/connector.conf
    sudo systemctl enable --now twingate-connector
    ''')

    ec2_instance = aws.ec2.Instance(
        f"Twingate-Connector-{i}",
        tags={
            "Name": pulumi.Output.all(connector.name).apply(
                lambda v: f"tg-{v[0]}"),
        },
        instance_type=data.get("ec2_type"),
        vpc_security_group_ids=[sg.id],
        ami=ami.id,
        key_name=keypair.key_name,
        user_data=user_data,
        subnet_id=private_subnet.id,
        associate_public_ip_address=False,
    )
