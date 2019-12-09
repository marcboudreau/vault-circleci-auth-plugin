#!/bin/sh

set -e

echo '{"plugin_directory": "/vault/plugins/"}' > /vault/config/local.json

docker-entrypoint.sh \
        vault \
        server \
        -dev \
        -dev-root-token-id=$VAULT_TOKEN \
        -config=/vault/config \
        ${VAULT_LOG_LEVEL:+"-log-level=$VAULT_LOG_LEVEL"} &
pid=$!

cat /vault/config/local.json

while ! vault status; do
	sleep 1
done

sha_sum=$(sha256sum /vault/plugins/vault-circleci-auth-plugin \
        | cut -d ' ' -f 1)

vault write \
        sys/plugins/catalog/vault-circleci-auth \
        sha_256=$sha_sum command=vault-circleci-auth-plugin

vault auth \
        enable \
        -path=circleci \
        -plugin-name=vault-circleci-auth \
        plugin

wait $pid
