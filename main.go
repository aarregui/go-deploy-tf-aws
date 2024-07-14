package main

import "github.com/aarregui/go-deploy-tf-aws/internal/cli"

func main() {
	cli := cli.New()
	startCLI(cli)
}

func startCLI(cli cli.CLIClient) {
	cli.Execute()
}
