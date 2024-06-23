resource "aws_security_group" "ecs" {
  name   = "${local.identifier}-sg-ecs"
  vpc_id = aws_vpc.main.id
  
  ingress {
    from_port       = var.app_port
    to_port         = var.app_port
    protocol        = "tcp"
    security_groups = [aws_security_group.lb.id]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_ecs_cluster" "go_deploy_tf_aws_cluster" {
  name = "${local.identifier}-cluster"
}

resource "aws_cloudwatch_log_group" "go_deploy_tf_aws" {
  name = "${local.identifier}-log-group"
}

resource "aws_ecs_task_definition" "go_deploy_tf_aws_task" {
  family                   = "${local.identifier}-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  memory                   = 512
  cpu                      = 256
  execution_role_arn       = aws_iam_role.go_deploy_tf_aws.arn
  task_role_arn            = aws_iam_role.go_deploy_tf_aws.arn
  container_definitions    = <<DEFINITION
  [{
    "name": "${local.identifier}-task",
    "image": "${data.aws_ecr_repository.go_deploy_tf_aws_repo.repository_url}:${var.tag}",
    "essential": true,
    "memory": 512,
    "cpu": 256,
    "networkMode": "awsvpc",
    "portMappings": [
      {
        "protocol": "tcp",
        "containerPort": ${var.app_port}, 
        "hostPort": ${var.app_port}
      }
    ],
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "${aws_cloudwatch_log_group.go_deploy_tf_aws.name}",
        "awslogs-region": "${var.region}",
        "awslogs-stream-prefix": "ecs"
      }
    },
    "environment": [
      {
        "name": "VERSION",
        "value": "${var.tag}"
      },
      {
        "name": "APP_ENV",
        "value": "${var.environment}"
      },
      {
        "name": "APP_PORT",
        "value": "${var.app_port}"
      },
      {
        "name": "DB_HOST",
        "value": "${aws_db_instance.default.address}"
      },
      {
        "name": "DB_PORT",
        "value": "${aws_db_instance.default.port}"
      },
      {
        "name": "DB_DATABASE",
        "value": "${aws_db_instance.default.db_name}"
      },
      {
        "name": "DB_USERNAME",
        "value": "${aws_db_instance.default.username}"
      },
      {
        "name": "AWS_RDS_MASTER_PASSWORD_SECRET_ID",
        "value": "${aws_secretsmanager_secret.rds.id}"
      },
      {
        "name": "APP_LOG_LEVEL",
        "value": "${var.app_log_level}"
      },
      {
        "name": "DB_LOG_LEVEL",
        "value": "${var.db_log_level}"
      }
    ]
  }]
  DEFINITION
}

resource "aws_iam_role" "go_deploy_tf_aws" {
  name               = "${local.identifier}-ecs-task-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "go_deploy_tf_aws_ecs_policy" {
  role       = aws_iam_role.go_deploy_tf_aws.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "go_deploy_tf_aws_secretsmanager_policy" {
  role       = aws_iam_role.go_deploy_tf_aws.name
  policy_arn = "arn:aws:iam::aws:policy/SecretsManagerReadWrite"
}

resource "aws_iam_role_policy_attachment" "go_deploy_tf_aws_ecs_exec_policy" {
  role       = aws_iam_role.go_deploy_tf_aws.name
  policy_arn = aws_iam_policy.ecs_exec_task.arn
}

resource "aws_iam_policy" "ecs_exec_task" {
  name = "${local.identifier}-ecs-exec-task"
  policy = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
       {
       "Effect": "Allow",
       "Action": [
            "ssmmessages:CreateControlChannel",
            "ssmmessages:CreateDataChannel",
            "ssmmessages:OpenControlChannel",
            "ssmmessages:OpenDataChannel"
       ],
      "Resource": "*"
      }
   ]
}
EOF
}

resource "aws_ecs_service" "go_deploy_tf_aws_service" {
  name                    = "${local.identifier}-ecs-service"
  cluster                 = aws_ecs_cluster.go_deploy_tf_aws_cluster.id
  task_definition         = aws_ecs_task_definition.go_deploy_tf_aws_task.arn
  launch_type             = "FARGATE"
  desired_count           = 1
  enable_execute_command  = true

  load_balancer {
    target_group_arn = aws_lb_target_group.main.arn
    container_name   = aws_ecs_task_definition.go_deploy_tf_aws_task.family
    container_port   = var.app_port
  }

   network_configuration {
    subnets         = aws_subnet.private.*.id
    security_groups = [aws_security_group.ecs.id]
  }
}

resource "aws_lb" "main" {
  name               =  "${local.identifier}-lb"
  security_groups    = [aws_security_group.lb.id]
  subnets            = aws_subnet.public.*.id
}

resource "aws_security_group" "lb" {
  name   =  "${local.identifier}-sg-lb"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}
 
resource "aws_lb_target_group" "main" {
  name        =  "${local.identifier}-lb-tg"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"
 
  health_check {
    healthy_threshold   = "3"
    interval            = "30"
    protocol            = "HTTP"
    matcher             = "200"
    timeout             = "3"
    path                = "/"
    unhealthy_threshold = "2"
  }
}

resource "aws_alb_listener" "http" {
  load_balancer_arn = aws_lb.main.id
  port              = 80
  protocol          = "HTTP"
 
  default_action {
   type = "redirect"
 
   redirect {
     port        = 443
     protocol    = "HTTPS"
     status_code = "HTTP_301"
   }
  }

  tags = {
    Name = "${local.identifier}-http"
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.id
  port              = 443
  protocol          = "HTTPS"
  certificate_arn   = aws_acm_certificate.main.arn
 
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.id
  }


  tags = {
    Name = "${local.identifier}-https"
  }

  depends_on = [
    aws_acm_certificate_validation.cert
  ]
}

resource "aws_acm_certificate" "main" {
  domain_name       = aws_route53_record.main.fqdn
  validation_method = "DNS"
  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name = local.identifier
  }
}

resource "aws_route53_record" "cert_validation" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.main.domain_validation_options)[0].resource_record_name
  records         = [ tolist(aws_acm_certificate.main.domain_validation_options)[0].resource_record_value ]
  type            = tolist(aws_acm_certificate.main.domain_validation_options)[0].resource_record_type
  zone_id  = data.aws_route53_zone.main.id
  ttl      = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = [ aws_route53_record.cert_validation.fqdn ]
}

resource "aws_route53_record" "main" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = "${local.api_host}"
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}
