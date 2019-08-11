# Deployment Requirements

## Building Application

### Dependencies

- \*NIX environment
- Bash
- Docker
- Docker Compose
- Golang
- vault
- jq
- yq

### registrymanager

#### Native

`cd registrymanager && go build -o registrymanager main.go` will create a `registrymanager` binary at `registrymanager/registrymanager`.

#### Docker

`make registrymanager` will build the application into an image tagged `registrymanager:latest`.

#### Dependencies

The `registrymanager` application makes use of a `postgres` database to store user information.

This connection can be configured through environment variables (see `.env-sample` or `secrets-example.json`).

By default the `docker-compose.yml` configuration will deploy a `postgres:10` container in the application stack.
