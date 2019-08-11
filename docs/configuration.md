# Docker Registry Configuration

The registry and associated services can all be configured with local files, however for ease of deployment, HashiCorp Vault has been used as a central config / secret manager.

Deployment scripts rely on this assumption, however all config files can be created using the provided template files.

## Dependencies

- \*NIX environment
- Bash
- Docker
- Docker Compose
- vault
- jq
- yq

### Optional Dependencies

- LDAP / AD server

## config.yml

The Docker Registry relies on a `config.yml` file to control the registry-specific configuration. Review the documentation here: https://docs.docker.com/registry/configuration/

To use the default configuration, `cp config.templ.yml config.yml`. This uses the local filesystem as the storage driver, and will store the registry data in a Docker named volume `registry_data` (defined in the `docker-compose.yml` file).

The `./scripts/storage` script will pull the storage configuration values from vault and update the `config.yml` file to reflect the driver and configuration values specified in Vault.

If using the Vault configs, a `secrets.json` file will be required, so it is recommended to run `make config.yml` to handle the creation of the secret file as well as the config file.

## .registry.env

The `registrymanager` and associated `postgres` database rely on specific configuration parameters and secrets to operate.

These are defined in a `.registry.env` file. This file will be generated from the `./scripts/secrets load` script with values pulled from Vault.

`make secrets.json` to pull the secrets from Vault. `make vaultsecrets` to push the local `secrets.json` file up to vault.

## LDAP Authentication

To utilize LDAP authentication, set the `LDAP_` variables for your environment.

## SSL / TLS

To utilize SSL / TLS, provide the path to the cert and key files. By default the `./scripts/secrets load` script will pull these from the Vault secret and place these in `$PWD/certs` and mount these into the `registrymanager` in the `/certs` directory.
