#!/bin/bash
go-deploy-tf-aws migrate up

if [ $? -eq 0 ]; then
    go-deploy-tf-aws serve
else
    exit 1
fi
