---
subcategory: "aws"
page_title: "Deploy AWS EC2 connector - Twingate Provider"
description: |-
An example of how to deploy a connector using a precompiled AMI
---

## Deploy a connector in AWS with a precompiled AMI

Given that we have defined a remote network and connector:

```terraform
resource "twingate_remote_network" "aws_remote_network" {
  name = "aws-remote-network"
}

resource "twingate_connector" "aws_connector" {
  remote_network_id = twingate_remote_network.aws_remote_network.id
}

resource "twingate_connector_tokens" "aws_connector_tokens" {
  connector_id = twingate_connector.aws_connector.id
}
```

Here is an proposed example of how you could deploy a connector using aws modules:

```terraform
# Getting the latest connector version
data "aws_ami" "latest" {
  most_recent = true
  filter {
    name = "name"
    values = [
      "twingate/images/hvm-ssd/twingate-amd64-*",
    ]
  }
  owners = ["617935088040"]
}

# define or use an existing VPC
module "demo_vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "2.64.0"

  name = "demo_vpc"
  cidr = "10.0.0.0/16"

  azs                            = ["us-east-1a"]
  private_subnets                = ["10.0.1.0/24"]
  public_subnets                 = ["10.0.2.0/24"]
  enable_classiclink_dns_support = true
  enable_dns_hostnames           = true
  enable_nat_gateway             = true

}

# define or use an existing Security group , the connector requires egress traffic enabled
module "demo_sg" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "3.17.0"
  vpc_id  = module.demo_vpc.vpc_id
  name    = "demo_security_group"
  egress_cidr_blocks = ["0.0.0.0/0"]
  egress_rules = ["all-tcp", "all-udp", "all-icmp"]
}

#spin off a ec2 instance from Twingate AMI
module "ec2_tenant_connector" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "2.19.0"

  name                   = "demo_connector"
  user_data = <<-EOT
    #!/bin/bash
    set -e
    mkdir -p /etc/twingate/
    {
      echo TWINGATE_URL="https://[NETWORK_NAME_HERE].twignate.com"
      echo TWINGATE_ACCESS_TOKEN="${twingate_connector_tokens.connector_tokens.access_token}"
      echo TWINGATE_REFRESH_TOKEN="${twingate_connector_tokens.connector_tokens.refresh_token}"
    } > /etc/twingate/connector.conf
    sudo systemctl enable --now twingate-connector
  EOT
  ami                    = data.aws_ami.latest.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [module.demo_sg.this_security_group_id]
  subnet_id              = module.demo_vpc.private_subnets[0]
}
```



