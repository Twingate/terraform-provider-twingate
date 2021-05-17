
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

module "vpc_tenant" {
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

module "tenant_sg_connector" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "3.17.0"
  vpc_id  = module.vpc_tenant.vpc_id
  name    = format("%s-connector", var.tenant_name)

  ingress_cidr_blocks = ["10.0.0.0/16"]
  ingress_rules       = ["ssh-tcp", "http-80-tcp", "all-icmp"]

  egress_cidr_blocks = ["0.0.0.0/0"]

  egress_rules = ["all-tcp", "all-udp", "all-icmp"]
  depends_on   = [module.vpc_tenant]
}

module "ec2_tenant_connector" {
  source  = "terraform-aws-modules/ec2-instance/aws"
  version = "2.19.0"

  name                   = "connector"
  instance_count         = 1
  user_data              = data.template_file.cloud_init.rendered
  ami                    = data.aws_ami.connector.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [module.tenant_sg_connector.this_security_group_id]
  subnet_id              = module.vpc_tenant.private_subnets[0]

  tags = {
    Environment    = var.tenant_name
    connector_name = twingate_connector.connector.name
  }
  depends_on = [module.vpc_tenant]
}