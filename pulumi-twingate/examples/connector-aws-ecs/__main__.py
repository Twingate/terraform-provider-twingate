import pulumi
import pulumi_aws as aws
import pulumi_awsx as awsx
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
    vpc_id=vpc.id
)

# Create a Twingate remote network
remote_network = tg.TwingateRemoteNetwork(data.get("tg_remote_network"), name=data.get("tg_remote_network"))

connectors = data.get("connectors")

cluster = aws.ecs.Cluster(data.get("cluster_name"), name=data.get("cluster_name"))

# Create a Fargate Service For Each Connector
for i in range(1, connectors + 1):
    connector = tg.TwingateConnector(f"twingate_connector_{i}", name="", remote_network_id=remote_network.id)
    connector_token = tg.TwingateConnectorTokens(f"connector_token_{i}", connector_id=connector.id)
    service = awsx.ecs.FargateService(f"Twingate-Connector-{i}",
                                      name=pulumi.Output.all(connector.name).apply(
                                          lambda v: f"tg-{v[0]}"),
                                      cluster=cluster.arn,
                                      network_configuration=aws.ecs.ServiceNetworkConfigurationArgs(
                                          subnets=[private_subnet.id],
                                          security_groups=[sg.id]
                                      ),
                                      desired_count=1,
                                      task_definition_args=awsx.ecs.FargateServiceTaskDefinitionArgs(
                                          container=awsx.ecs.TaskDefinitionContainerDefinitionArgs(
                                              image="twingate/connector:1",
                                              cpu=data.get("container_cpu"),
                                              memory=data.get("container_memory"),
                                              environment=[
                                                  awsx.ecs.TaskDefinitionKeyValuePairArgs(name="TENANT_URL",
                                                                                          value=f"https://{tg_account}.twingate.com"),
                                                  awsx.ecs.TaskDefinitionKeyValuePairArgs(name="ACCESS_TOKEN",
                                                                                          value=connector_token.access_token),
                                                  awsx.ecs.TaskDefinitionKeyValuePairArgs(name="REFRESH_TOKEN",
                                                                                          value=connector_token.refresh_token),
                                                  awsx.ecs.TaskDefinitionKeyValuePairArgs(name="TWINGATE_LABEL_DEPLOYEDBY",
                                                                                          value="tg-pulumi-aws-ecs")]
                                          ),
                                      ))