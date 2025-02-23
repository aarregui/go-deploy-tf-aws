.PHONY: build

BINARY_NAME=main.out
COVERAGE=coverage
COV_PROFILE=${COVERAGE}.out
COV_HTML=${COVERAGE}.html
 
all: deps build test

local-deps:
	go install github.com/cortesi/modd/cmd/modd@latest

build:
	go build -o ${BINARY_NAME} main.go

test:
	go test -coverprofile ${COV_PROFILE} ./...

coverage:
	go tool cover -html=${COV_PROFILE} -o ${COV_HTML}

install:
	go install

start:
	docker-compose up -d

start-local: build
	./${BINARY_NAME} serve

watch:
	modd

reset:
	docker-compose down -v
	docker-compose up -d
	go run main.go migrate up

logs:
	docker-compose logs -f go-deploy-tf-aws
