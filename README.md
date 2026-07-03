# rabbitctl

A CLI to remotely administer RabbitMQ nodes via the Management HTTP API.

## Overview

`rabbitctl` is a lightweight alternative to `rabbitmqctl`, the CLI bundled up with RabbitMQ, which allows for programmatic operations in any environment where the `rabbitmqctl` binary is unavailable or inconvenient.

It is not a complete RabbitMQ client.

The tool is built in Go using spf13/cobra for CLI structure and michaelklishin/rabbit-hole for the Management API client.

## Features

...

## Installation

### Install the latest release

```bash
go install github.com/nepec/rabbitctl@latest
```

Or download a pre-built binary from the Releases page.

### Build from source

```bash
git clone https://github.com/nepec/rabbitctl.git
cd rabbitctl
go build -o rabbitctl .
```

## Quick start

```bash
# Connection via flags
rabbitctl list queues --hostname rabbitmq.example.com -u admin -p admin

# Connection via env variables
export RMQ_HOSTNAME=rabbitmq.example.com
export RMQ_USERNAME=admin
export RMQ_PASSWORD=admin
rabbitctl list vhosts

# List queues with filters
rabbitctl list queues --vhosts '/' --contains "substring" --type quorum

# Apply the merge
rabbitctl merge-policy queues \
  --file definitions.json \
  --vhosts /
```

## Configuration

| Flag                | Short | Default     | Description                       |
|---------------------|-------|-------------|-----------------------------------|
| `--hostname`        | `-H`  | `localhost` | RabbitMQ Management API host      |
| `--port`            | `-P`  | `15672`     | RabbitMQ Management API port      |
| `--username`        | `-u`  | `guest`     | Username for Basic Auth           |
| `--password`        | `-p`  | `guest`     | Password for Basic Auth           |
| `--log-level`       |       | `info`      | Log level (debug, info, warn, error) |
| `--config`          |       |             | Path to a YAML config file        |

### Environment variables

All root flags can be set via environment variables with the prefix `RMQ_` and hyphens replaced by underscores:

| Env Variable    | Corresponds to  |
|-----------------|-----------------|
| `RMQ_HOSTNAME`  | `--hostname`    |
| `RMQ_PORT`      | `--port`        |
| `RMQ_USERNAME`  | `--username`    |
| `RMQ_PASSWORD`  | `--password`    |
| `RMQ_LOG_LEVEL` | `--log-level`   |

### Config file

Place credentials in `~/.rabbitctl.yaml` (or pass `--config /path/to/file`):

```yaml
hostname: rabbitmq.example.com
port: 15672
username: admin
password: secret
log-level: info
```

Precedence (highest to lowest): `--config` flag → `RMQ_*` env vars → config file → flag defaults.

## Roadmap

- [ ] integration tests with testcontainers
- [ ] Native TLS support
- [ ] Rest of CRUD commands for queues
- [ ] JSON output

## Development

### Prerequisites

- Go 1.25 or later

### Build

```bash
go build -o rabbitctl ./cmd/rabbitctl/main.go
```

### Run unit tests

```bash
go test ./pkg/...
```

### Run integration tests

Start a RabbitMQ Management container, then run the integration suite:

```bash
docker run -d --rm \
  --name rmq-test \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:management

go test ./integration/...
```

Update golden.snapshot files after intentional output changes:

```bash
GOLDEN_UPDATE=true go test ./integration/...
```
