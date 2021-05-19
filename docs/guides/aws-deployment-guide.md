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

First, we need to set up the Twingate Terraform provider by providing your network ID and the API key you provisioned earlier. TODO EXPLAIN HOW SECRETS ARE RETRIEVED IF NEEDED

```terraform
provider "twingate" {
  api_token = data.sops_file.secret.data["autoco_api_token"]
  network   = "autoco"
  url       = lookup(local.twingate_domain, var.tenant_namespace)
}
```

## Creating the Remote Network and Connectors in Twingate

Next, we'll create the objects in Twingate that correspond to the AWS network that we're deploying Twingate into: A Remote Network to represent the AWS VPC, and a Connector to be deployed in that VPC. We'll use these objects when we're deploying the Connector image and creating Resources to access through Twingate.

```terraform
resource "twingate_remote_network" "my_aws_network" {
  name = "AWS Network"
}

resource "twingate_connector" "aws_connector" {
  remote_network_id = twingate_remote_network.my_aws_network.id
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
  owners = [lookup(local.ami_owners, var.tenant_namespace)]
}
```

Next, we need to create a shell script to [run a command](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html) to configure the Connector's tokens when the AMI is launched. We recommend you use the following template, which we will name `aws-connector-runner.sh.tpl`:

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

Now, let's configure the template with your Twingate URL (which will always be the same for a given organization) and the Connector tokens created in the previous step.

```terraform
data "template_file" "aws_init" {
  template = file("${path.module}/bin/aws-connector-runner.sh")

  vars = {
    url           = "https://autoco.twingate.com"
    access_token  = twingate_connector_tokens.aws_connector_tokens.access_token
    refresh_token = twingate_connector_tokens.aws_connector_tokens.refresh_token
  }
}
```

Now we're ready to deploy the Connector AMI image to the VPC. For the purpose of this example, we'll assume you already have a VPC, subnet, and security group created. We'll deploy the Connector on a private subnet, because it doesn't need and shouldn't have a public IP address.

```terraform
module "aws_connector" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "2.16.0"

  name                   = "AWS Connector"
  instance_count         = 1
  user_data              = data.template_file.cloud_init.rendered
  ami                    = data.aws_ami.connector.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [module.my_security_group.this_security_group_id]
  subnet_id              = module.my_vpc.private_subnets[0]

  depends_on = [module.vpc_beamreach] TODO ROMAN DO WE ACTUALLY NEED THIS?
}
```

## Creating Resources

TODO NEED EXAMPLES OF CREATING TG RESOURCES