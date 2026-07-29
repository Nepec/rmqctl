// rmqctl is a CLI for administering RabbitMQ nodes via the Management HTTP API.
//
// It is intended for operators and developers who do not have access to or
// do not wish to use the official rabbitmqctl command-line tool.
//
// Basic usage:
//
//	rmqctl [flags] <command> [args...]
//
// Commands include queue, exchange, vhost, and policy management. Use
// 'rmqctl help' for the full command list, or 'rmqctl help <command>'
// for details on a specific command.
//
// Configuration is loaded from flags, environment variables, or a config file
// (~/.rmqctl.yaml by default). See the README for details.
package main

import (
	"fmt"
	"os"

	"github.com/nepec/rmqctl/internal/cli"
)

var version = "dev"

func main() {
	info := cli.BuildInfo{Version: version}

	if err := cli.Execute(info); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "The command terminated due to an error: %v\n", err)
		os.Exit(1)
	}
}
