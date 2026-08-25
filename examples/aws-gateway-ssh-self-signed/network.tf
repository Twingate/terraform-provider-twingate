data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = { Name = "demo-vpc" }
}

# Private subnet for all instances. They receive no public IP and reach the
# internet through the NAT gateway; operators connect via an EC2 Instance Connect Endpoint (EIC).
resource "aws_subnet" "main" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = false

  tags = { Name = "demo-private-subnet" }
}

# Public subnet hosts only the NAT gateway.
resource "aws_subnet" "public" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = { Name = "demo-public-subnet" }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = { Name = "demo-igw" }
}

resource "aws_eip" "nat" {
  domain = "vpc"

  tags = { Name = "demo-nat-eip" }
}

# NAT gateway provides outbound internet for the private instances so they can
# pull packages, reach Twingate, and register with SSM.
resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public.id

  tags = { Name = "demo-nat" }

  depends_on = [aws_internet_gateway.main]
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = { Name = "demo-public-rt" }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table" "main" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }

  tags = { Name = "demo-private-rt" }
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

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "demo-internal-sg" }
}

resource "aws_security_group" "eic" {
  name   = "demo-eic-endpoint"
  vpc_id = aws_vpc.main.id

  egress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = [aws_subnet.main.cidr_block]
  }

  tags = { Name = "demo-eic-endpoint-sg" }
}

resource "aws_key_pair" "debug-key" {
  key_name_prefix = "debug-key-"
  public_key      = var.ssh_public_key
}

resource "aws_security_group_rule" "eic_to_internal_ssh" {
  type                     = "ingress"
  protocol                 = "tcp"
  from_port                = 22
  to_port                  = 22
  security_group_id        = aws_security_group.internal.id
  source_security_group_id = aws_security_group.eic.id
}

