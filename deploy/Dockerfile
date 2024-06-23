FROM golang:1.22.0

ADD . ./go-deploy-tf-aws

RUN cd ./go-deploy-tf-aws && make install

COPY ./internal/migrations /go/internal/migrations

RUN chmod 755 ./go-deploy-tf-aws/deploy/entrypoint.sh

ENTRYPOINT ["./go-deploy-tf-aws/deploy/entrypoint.sh"]
