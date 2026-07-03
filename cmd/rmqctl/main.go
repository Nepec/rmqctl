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
	"github.com/nepec/rmqctl/internal/cli"
	_ "github.com/nepec/rmqctl/internal/cli/apply"
	_ "github.com/nepec/rmqctl/internal/cli/list"
	_ "github.com/nepec/rmqctl/internal/cli/mergepolicy"
)

func main() {
	cli.Execute()
}
