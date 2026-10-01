data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = { Name = "demo-vpc" }
}

# All instances are on a public subnet to make debugging easier.
# All access is still restricted by the security group below as well as toggleable firewall rules.
# Production implementations should be moved to a private subnet.
resource "aws_subnet" "main" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.0.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = { Name = "demo-subnet" }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = { Name = "demo-igw" }
}

resource "aws_route_table" "main" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = { Name = "demo-rt" }
}

resource "aws_route_table_association" "main" {
  subnet_id      = aws_subnet.main.id
  route_table_id = aws_route_table.main.id
}

resource "aws_security_group" "internal" {
  name   = "demo-internal"
  vpc_id = aws_vpc.main.id

  ingress {
    protocol  = "tcp"
    from_port = 0
    to_port   = 65535
    self      = true
  }

  dynamic "ingress" {
    for_each = var.debug_ssh ? [1] : []

    content {
      protocol        = "tcp"
      from_port       = 22
      to_port         = 22
      prefix_list_ids = [data.aws_ec2_managed_prefix_list.eic.id]
    }
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "demo-internal-sg" }
}

# AWS-managed list of the IP ranges the console's EC2 Instance Connect uses.
data "aws_ec2_managed_prefix_list" "eic" {
  name = "com.amazonaws.${var.aws_region}.ec2-instance-connect"
}
