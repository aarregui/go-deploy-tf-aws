terraform {
  backend "s3" {
    bucket = "terraform-state-go-deploy-tf-aws"
    region = "us-west-1"
  }
}
