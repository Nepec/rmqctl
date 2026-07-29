# rmqctl

A CLI to remotely administer RabbitMQ nodes via the Management HTTP API.

> [!WARNING]
> `rmqctl` is an early-stage project under active development. Commands,
> flags, and manifest formats may change without notice. Not yet
> recommended for production use.

## Overview

`rmqctl` is a lightweight alternative to `rabbitmqctl` for environments
where the bundled binary isn't available or convenient. It supports two
styles of operation:

- **Inspect/mutate** existing resources imperatively (`list`, `merge`)
- **Declaratively provision** resources from a YAML manifest (`apply`)

It is not a complete RabbitMQ client.

## Installation

```bash
go install github.com/nepec/rmqctl@latest
```

Or build from source:

```bash
git clone https://github.com/nepec/rmqctl.git
cd rmqctl
make build
```

## Usage

```bash
# List resources
rmqctl list queues --hostname rabbitmq.example.com -u admin -p admin
rmqctl list vhosts

# Merge (integrate) existing policy definitions into a vhost
rmqctl merge queues --file definitions.json --vhosts /

# Apply a declarative manifest
rmqctl apply --file manifest.yaml --vhosts / --dry-run
```

See `testdata/` for manifest examples.

## Configuration

| Flag           | Short | Default     | Description                   |
|----------------|-------|-------------|--------------------------------|
| `--hostname`   | `-H`  | `localhost` | Management API hostname       |
| `--port`       | `-P`  | `15672`     | Management API port           |
| `--username`   | `-u`  | `guest`     | Basic Auth username           |
| `--password`   | `-p`  | `guest`     | Basic Auth password           |
| `--log-level`  |       | `info`      | debug, info, warn, error      |
| `--config`     |       |             | Path to a YAML config file    |

All flags can also be set via `RMQ_*` environment variables (e.g.
`RMQ_HOSTNAME`), or in `~/.rmqctl.yaml`.

Precedence: flags → env vars → config file → defaults.

## Development

```bash
make test         # unit tests
make integration   # integration tests (requires Docker)
make fmt           # format with gofumpt
make golangci-lint # lint
```
