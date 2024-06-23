terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = "us-west-1"
  default_tags {
    tags = {
      app-id  = "${var.app_name}-backend"
      env     = var.environment
    }
  }
}
