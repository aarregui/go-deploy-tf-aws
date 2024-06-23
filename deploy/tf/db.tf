resource "aws_security_group" "rds" {
  name = "${local.identifier}-sg-rds"
  vpc_id = aws_vpc.main.id

  ingress {
    protocol        = "tcp"
    from_port       = 5432
    to_port         = 5432
    cidr_blocks     = ["0.0.0.0/0"]
    security_groups = [aws_security_group.ecs.id]
  }

  egress {
    from_port   = 0
    to_port     = 65535
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "db" {
  name       = local.identifier
  subnet_ids = aws_subnet.private.*.id
}

resource "random_password" "rds" {
  length            = 40
  special           = true
  min_special       = 5
  override_special  = "!#$%^&*()-_=+[]{}<>:?"
  keepers           = {
    pass_version  = 1
  }
}

resource "aws_secretsmanager_secret" "rds" {
  name = "${local.identifier}-rds-master-password"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "rds" {
  secret_id     = aws_secretsmanager_secret.rds.id
  secret_string = random_password.rds.result
}

resource "aws_db_instance" "default" {
  identifier                = local.identifier
  allocated_storage         = 5
  engine                    = "postgres"
  engine_version            = "14.10"
  instance_class            = "db.t4g.micro"
  db_name                   = "godeploytfaws"
  username                  = "dbuser"
  password                  = random_password.rds.result
  final_snapshot_identifier = "${local.identifier}-final-snapshot"
  apply_immediately         = true
  vpc_security_group_ids    = [aws_security_group.rds.id]
  db_subnet_group_name      = aws_db_subnet_group.db.name
}
