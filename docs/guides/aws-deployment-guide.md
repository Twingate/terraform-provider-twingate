---
subcategory: "aws"
page_title: "AWS EC2 Deployment Guide"
description: |-
This document walks you through a basic deployment using Twingate's Terraform provider on AWS
---

# Deployment Guide

This deployment guide walks you through a basic AWS deployment of Twingate. For more information about Twingate, please reference the Twingate [documentation](https://docs.twingate.com/docs). It assumes basic knowledge of Twingate's service, the AWS Terraform provider, and a pre-existing AWS deployment in Terraform.

## Before you begin

* Sign up for an account on the [Twingate website](https://www.twingate.com). You will need the Twingate Enterprise tier to use Terraform with Twingate.
* Create a Twingate [API key](https://docs.twingate.com/docs/api-overview). The key will need to have full permissions to Read, Write, & Provision, in order to deploy Connectors through Terraform.

## Setting up the Provider

First, we need to set up the Twingate Terraform provider by providing your network ID and the API key you provisioned earlier.

```terraform
provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}
```

## Creating the Remote Network and Connectors in Twingate

Next, we'll create the objects in Twingate that correspond to the AWS network that we're deploying Twingate into: A Remote Network to represent the AWS VPC, and a Connector to be deployed in that VPC. We'll use these objects when we're deploying the Connector image and creating Resources to access through Twingate.

```terraform
resource "twingate_remote_network" "aws_network" {
  name = "AWS Network"
}

resource "twingate_connector" "aws_connector" {
  remote_network_id = twingate_remote_network.aws_network.id
}

resource "twingate_connector_tokens" "aws_connector_tokens" {
  connector_id = twingate_connector.aws_connector.id
}
```

## Deploying the Connector

Now that we have the data types created in Twingate, we need to deploy a Connector into the AWS VPC to handle Twingate traffic. We'll use the pre-existing AWS AMI image for the Twingate Connector. First, we need to look up the latest AMI ID.

```terraform
data "aws_ami" "connector" {
  most_recent = true
  filter {
    name = "name"
    values = [
      "twingate/images/hvm-ssd/twingate-amd64-*",
    ]
  }
  owners = ["617935088040"]
}
```

Now, let's go ahead and deploy the AMI. For this example, we're creating a new VPC and security group, but you can use an existing one too. We'll deploy the Connector on a private subnet, because it doesn't need and shouldn't have a public IP address. Note the shell script that we use to configure the Connector tokens when the AMI launches.

```terraform
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

# define or use an existing Security group, the Connector requires egress traffic enabled but does not require ingress
module "demo_sg" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "3.17.0"
  vpc_id  = module.demo_vpc.vpc_id
  name    = "demo_security_group"
  egress_cidr_blocks = ["0.0.0.0/0"]
  egress_rules = ["all-tcp", "all-udp", "all-icmp"]
}

# spin off a ec2 instance from Twingate AMI and configure tokens in user_data
module "ec2_tenant_connector" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "2.19.0"

  name                   = "demo_connector"
  user_data = <<-EOT
    #!/bin/bash
    set -e
    mkdir -p /etc/twingate/
    {
      echo TWINGATE_URL="https://autoco.twignate.com"
      echo TWINGATE_ACCESS_TOKEN="${twingate_connector_tokens.aws_connector_tokens.access_token}"
      echo TWINGATE_REFRESH_TOKEN="${twingate_connector_tokens.aws_connector_tokens.refresh_token}"
    } > /etc/twingate/connector.conf
    sudo systemctl enable --now twingate-connector
  EOT
  ami                    = data.aws_ami.latest.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [module.demo_sg.this_security_group_id]
  subnet_id              = module.demo_vpc.private_subnets[0]
}
```

## Creating Resources

Now that you've deployed the Connector, we can create Resources on the same Remote Network that can be accessed through Twingate. You'll need to define the Group ID explicitly, which you can pull from the Twingate Admin Console or the API. It's the 12 character ID at the end of the URL when you're viewing the Group in the Admin Console. For this example, we'll assume you already have an `aws_instance` defined.

```terraform
resource "twingate_resource" "tg_instance" {
  name = "My AWS Instance"
  address = aws_instance.my_instance.private_dns
  remote_network_id = twingate_remote_network.my_aws_network.id
  group_ids = ["R3JvdXG6OGky"]
}
```