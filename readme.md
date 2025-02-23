# go-deploy-tf-aws
An example of how to use CI/CD to deploy a Go webserver to AWS using Terraform and Github actions.

Some of the AWS resources used are:
* ECR (new images are pushed with Github Actions)
* ECS / EC2 / Load Balancer
* VPC
* RDS (locked inside a private subnet)
* ACM (to enable HTTPS)

## AWS Setup
The following AWS resources have to be created manually:
* S3 bucket for Terraform state (`terraform-state-go-deploy-tf-aws`)
* ECR for the Docker images, to be used by the ECS tasks (`go-deploy-tf-aws`)
* A domain for the app, ie: [Route 53 hosted zone](https://github.com/aarregui/go-deploy-tf-aws/blob/master/deploy/tf/variables.tf#L24-L26)

To create the S3 bucket use:
```
aws s3api create-bucket --bucket terraform-state-go-deploy-tf-aws --create-bucket-configuration LocationConstraint=us-west-1
```

To create the ECS use:
```
aws ecr create-repository --repository-name go-deploy-tf-aws
```

## CI/CD

Semver is used for tagging, branches must be created with the correct format to generate the correct version tags and releases.

* **dev** - pushing to a `[feature|fix]/*` branch
* **prod** - merge a `[feature|fix]/*` branch into `master` or push directly to it

## Conect to the DB instance within the VPC
1. Find the Task ID
    ```bash
    aws ecs list-tasks --cluster prod-go-deploy-tf-aws-cluster
    ```
1. Find the DB ID
    ```bash
    aws rds describe-db-instances --db-instance-identifier prod-go-deploy-tf-aws
    ```
1. Get the DB password:
    ```bash
    aws secretsmanager get-secret-value --secret-id prod-go-deploy-tf-aws-rds-master-password
    ```
1. Get inside the Task, install psql and connect to the DB instance
    ```bash
    aws ecs execute-command --cluster prod-go-deploy-tf-aws-cluster \
        --task {task_id} \
        --container prod-go-deploy-tf-aws \
        --interactive \
        --command "/bin/sh"

    apt update && \
    apt install postgresql -y

    psql -h {db_host} -p 5432 -U dbuser -d go-deploy-tf-aws
    ```

## Local Development
1. Run `cp .env.example .env`
1. Run `make local-deps`
    * [air](https://github.com/air-verse/air) is used for hot reloading in local environments.

To start the server in your host machine:
1. Run `docker compose up -d go-deploy-tf-aws-db`
1. Run `make watch`

To start the server with docker:
1. Run `make start && make logs`
