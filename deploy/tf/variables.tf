variable "tag" {}
variable "environment" {}

variable "region" {
  default = "us-west-1"
}

variable "app_port" {
  default = 8000
}

variable "app_log_level" {
  default = "1"
}

variable "db_log_level" {
  default = "1"
}

variable "app_name" {
  default = "go-deploy-tf-aws"
}

variable "domain_name" {
  default = "go-deploy-aws.com"
}

variable "availability_zones" {
  description = "a comma-separated list of availability zones, defaults to all AZ of the region, if set to something other than the defaults, both private_subnets and public_subnets have to be defined as well"
  # default     = ["us-west-1a", "us-west-1b", "us-west-1c"]
  default     = ["us-west-1a", "us-west-1b"]
}

variable "cidr" {
  description = "The CIDR block for the VPC."
  default     = "10.0.0.0/16"
}

variable "private_subnets" {
  description = "a list of CIDRs for private subnets in your VPC, must be set if the cidr variable is defined, needs to have as many elements as there are availability zones"
  # default     = ["10.0.0.0/20", "10.0.32.0/20", "10.0.64.0/20"]
  default     = ["10.0.0.0/20", "10.0.32.0/20"]
}

variable "public_subnets" {
  description = "a list of CIDRs for public subnets in your VPC, must be set if the cidr variable is defined, needs to have as many elements as there are availability zones"
  # default     = ["10.0.16.0/20", "10.0.48.0/20", "10.0.80.0/20"]
  default     = ["10.0.16.0/20",  "10.0.48.0/20"]
}
