#!/bin/bash
set -euo pipefail

trap clean_up ERR EXIT

function clean_up() {
    # Clean up Vault docker containers
    docker_vault_container=$(docker ps -f name=vault -q)
    if [[ $docker_vault_container ]]; then
        docker rm -f $docker_vault_container > /dev/null
    fi

    # Clean up CircleCI docker containers
    for i in 1 2 3 4 5; do
        docker_circle_container=$(docker ps -aq -f name=circle$i)
        if [[ $docker_circle_container ]]; then
            docker rm -f $docker_circle_container > /dev/null
        fi
    done

    # Clean up docker network
    docker_network=$(docker network ls -f name=vaulttest -q)
    if [[ $docker_network ]]; then
        docker network rm $docker_network > /dev/null
    fi
}

status_codes=(200 200 200 404 500)
grep_expressions=("circleci build is not currently running"
                  "provided VCS revision does not match the revision reported by circleci"
                  ""
                  'Errors: * 404:'
                  'Errors: * 500:')

# Creating the Docker Network vaulttest
echo -n "Creating docker network: " ; docker network create vaulttest

# Creating the mock CircleCI server containers
for i in 1 2 3 4 5; do
    echo -n "Creating docker container for mock circleci server $i: "
    docker create --rm --name circle$i --network vaulttest \
            marcboudreau/dumb-server:latest \
            -sc ${status_codes[$((i-1))]} -resp /response
    docker cp ./responses/circle$i circle$i:/response
    echo -n "Starting docker container " ; docker start circle$i

done

# Creating the Vault server container
mkdir -p ./plugins
cp -f $(dirname $0)/../vault-circleci-auth-plugin ./plugins/

echo -n "Creating docker container for vault: "
docker create --rm --name vault --network vaulttest \
        -e VAULT_TOKEN=root -e VAULT_ADDR=http://127.0.0.1:8200 \
        -e VAULT_LOCAL_CONFIG='{"plugin_directory": "/vault/plugins/"}' \
        vault:0.9.2 server -dev -dev-root-token-id=root \
        -log-level trace
docker cp plugins vault:/vault/
echo -n "Starting docker container " ; docker start vault

sha_sum=$(docker exec vault sha256sum /vault/plugins/vault-circleci-auth-plugin | cut -d ' ' -f 1)

docker exec vault vault write sys/plugins/catalog/vault-circleci-auth \
        sha_256=$sha_sum command=vault-circleci-auth-plugin

# Testing login endpoint
for i in 1 2 3 4 5; do
    docker exec vault vault auth enable -path=test$i \
            -plugin-name=vault-circleci-auth plugin
    docker exec vault vault write auth/test$i/config circleci_token=fake \
            vcs_type=github owner=johnsmith ttl=5m max_ttl=15m \
            base_url=http://circle$i:7979

    read response <<< $(docker exec vault vault write -format=json \
            auth/test$i/login project=someproject build_num=100 \
            vcs_revision=babababababababababababababababababababa 2>&1)

    if [[ ${grep_expressions[$((i-1))]} ]]; then
        echo "$response" | grep -F "${grep_expressions[$((i-1))]}" > /dev/null \
                && echo "Test $i PASSED"
    else
        [[ $(echo "$response" | jq -r '.auth.client_token' | wc -c) -gt 0 ]] \
                && echo "Test $i PASSED"
    fi
done

# Testing a second attempt at authenticating the same build
read response <<< $(docker exec vault vault write auth/test5/login project=someproject build_num=100 \
        vcs_revision=babababababababababababababababababababa 2>&1)

echo $response | grep -F "an attempt to authenticate as this build has already been made" > /dev/null \
        && echo "Test $i-supplement PASSED"
