---
subcategory: "aws"
page_title: "Deploy AWS EC2 connector - Twingate Provider"
description: |-
An example of how to deploy a connector using a precompiled AMI
---

## Deploy a connector in AWS with a precompiled AMI

Given that we have defined a remote network and connector:

```terraform
variable "tenant_name" {
}

resource "twingate_remote_network" "network" {
  name = "${var.tenant_name}-network"
}

resource "twingate_connector" "connector" {
  remote_network_id = twingate_remote_network.network.id
}

resource "twingate_connector_tokens" "connector_tokens" {
  connector_id = twingate_connector.connector.id
}
```

Here is an proposed example of how you could deploy a connector using aws modules:

```terraform
# Getting the latest connector version
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

# AMI startup template
data "template_file" "cloud_init" {
  template = file("${path.module}/aws-connector-runner.sh.tpl")
  vars = {
    url           = "https://${var.tenant_name}.twignate.com"
    access_token  = twingate_connector_tokens.connector_tokens.access_token
    refresh_token = twingate_connector_tokens.connector_tokens.refresh_token
  }
}

module "vpc_beamreach" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "2.64.0"

  name = "vpc"
  cidr = "10.0.0.0/16"

  azs                            = ["us-east-1a"]
  private_subnets                = ["10.0.1.0/24"]
  public_subnets                 = ["10.0.2.0/24"]
  enable_classiclink_dns_support = true
  enable_dns_hostnames           = true
  enable_nat_gateway             = true


  tags = {
    Environment = var.tenant_name
  }
}

module "beamreach_sg_connector" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "3.17.0"
  vpc_id  = module.vpc_beamreach.vpc_id
  name    = format("%s-connector", var.tenant_name)

  ingress_cidr_blocks = ["10.0.0.0/16"]
  ingress_rules       = ["ssh-tcp", "http-80-tcp", "all-icmp"]

  egress_cidr_blocks = ["0.0.0.0/0"]

  egress_rules = ["all-tcp", "all-udp", "all-icmp"]
  depends_on   = [module.vpc_beamreach]
}

module "ec2_beamreach_connector" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "2.19.0"

  name                   = "connector"
  instance_count         = 1
  user_data              = data.template_file.cloud_init.rendered
  ami                    = data.aws_ami.connector.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [module.beamreach_sg_connector.this_security_group_id]
  subnet_id              = module.vpc_beamreach.private_subnets[0]

  tags = {
    Environment    = var.tenant_name
    connector_name = twingate_connector.connector.name
  }
  depends_on = [module.vpc_beamreach]
}
```

And the sh file used in the example

```sh
#!/bin/bash
set -e
mkdir -p /etc/twingate/

{
  echo TWINGATE_URL="${url}"
  echo TWINGATE_ACCESS_TOKEN="${access_token}"
  echo TWINGATE_REFRESH_TOKEN="${refresh_token}"
} > /etc/twingate/connector.conf

sudo systemctl enable --now twingate-connector
```

