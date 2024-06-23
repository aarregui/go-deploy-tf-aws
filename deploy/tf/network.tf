resource "aws_vpc" "main" {
  cidr_block           = var.cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = local.identifier
  }
}

resource "aws_subnet" "private" {
  count             = length(var.private_subnets)
  vpc_id            = aws_vpc.main.id
  cidr_block        = element(var.private_subnets, count.index)
  availability_zone = element(var.availability_zones, count.index)

  tags = {
    Name = "${local.identifier}-private-${count.index}"
  }
}
 
resource "aws_subnet" "public" {
  count                   = length(var.public_subnets)
  vpc_id                  = aws_vpc.main.id
  cidr_block              = element(var.public_subnets, count.index)
  availability_zone       = element(var.availability_zones, count.index)
  map_public_ip_on_launch = true

  tags = {
    Name = "${local.identifier}-public-${count.index}"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id  = aws_vpc.main.id

  tags = {
    Name = local.identifier
  }
}

resource "aws_route" "internet_access" {
  route_table_id         = aws_vpc.main.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.main.id
}

resource "aws_eip" "main" {
  count      = length(var.public_subnets)
  vpc        = true
  depends_on = [aws_internet_gateway.main]

  tags = {
    Name = "${local.identifier}-${count.index}"
  }
}

resource "aws_nat_gateway" "main" {
  count         = length(var.public_subnets)
  subnet_id     = element(aws_subnet.public.*.id, count.index)
  allocation_id = element(aws_eip.main.*.id, count.index)

  tags = {
    Name = "${local.identifier}-${count.index}"
  }
}

resource "aws_route_table" "private" {
  count  = length(var.public_subnets)
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = element(aws_nat_gateway.main.*.id, count.index)
  }

  tags = {
    Name = "${local.identifier}-${count.index}"
  }
}

resource "aws_route_table_association" "private" {
  count          = length(var.public_subnets)
  subnet_id      = element(aws_subnet.private.*.id, count.index)
  route_table_id = element(aws_route_table.private.*.id, count.index)
}
