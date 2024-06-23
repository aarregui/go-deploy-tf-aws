data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

data "aws_ecr_repository" "go_deploy_tf_aws_repo" {
  name = var.app_name
}

data "aws_route53_zone" "main" {
  name = var.domain_name
}
