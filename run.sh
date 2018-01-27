#!/bin/bash

set -euo pipefail

${DEBUG:+set -x}

GOOS=linux GOARCH=amd64 go build -o vault-circleci-auth-plugin

docker rm -f vault || true

docker run -d -e VAULT_ADDR=http://127.0.0.1:8200 -e VAULT_LOCAL_CONFIG='{"plugin_directory": "/vault/plugins"}' -v $PWD/vault-circleci-auth-plugin:/vault/plugins/vault-circleci-auth-plugin --name=vault vault:0.9.2 server -dev -dev-root-token-id=root

docker exec vault vault login root

SHASUM=$(shasum -a 256 "./vault-circleci-auth-plugin" | cut -d " " -f1)
docker exec vault vault write sys/plugins/catalog/vault-circleci-auth sha_256="$SHASUM" command="vault-circleci-auth-plugin"

docker exec vault vault auth enable -path=circleci -plugin-name=vault-circleci-auth plugin

docker exec -it vault ash
