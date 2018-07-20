#!/bin/bash
set -euo pipefail

trap clean_up ERR EXIT

function clean_up() {
    # Clean up Vault and Circle docker container
    docker rm -f vault circle > /dev/null 2>&1 || true

    # Clean up docker network
    docker network rm vaulttest > /dev/null 2>&1 || true
}

base_dir=$(dirname $0)

status_codes=(200 200 200 404 500)
grep_expressions=("circleci build is not currently running"
                  "provided VCS revision does not match the revision reported by circleci"
                  ""
                  '* 404: {"message":"Not Found","documentation_url":"https://developer.github.com/v3/repos/#get"}'
                  '* 500: An internal error occurred')

# Creating the Docker Network vaulttest
echo -n "Creating docker network: " ; docker network create vaulttest

# Creating the mock CircleCI server containers
for i in 1 2 3 4 5; do
    echo -n "Creating docker container for mock circleci server $i: "
    docker create --rm --name circle --network vaulttest \
            marcboudreau/dumb-server:latest \
            -sc ${status_codes[$((i-1))]} -resp /response
    docker cp $base_dir/responses/circle$i circle:/response
    echo -n "Starting docker container " ; docker start circle

    echo -n "Creating docker container for vault: "
    docker run \
            --rm \
            -d \
            --name vault \
            --network vaulttest \
            -e VAULT_TOKEN=root \
            -e VAULT_ADDR=http://127.0.0.1:8200 \
            -e VAULT_LOG_LEVEL=trace \
            vault-circleci-auth-plugin:test

    while ! docker exec vault vault auth list | grep 'circleci/' > /dev/null ; do
        echo "Still waiting for Vault server to finish initializing..."
    done

    attempt_cache_expiry=
    if (( $i == 3 )); then
        attempt_cache_expiry="attempt_cache_expiry=5s"
    fi

    docker exec vault vault write auth/circleci/config circleci_token=fake \
            vcs_type=github owner=johnsmith ttl=5m max_ttl=15m \
            base_url=http://circle:7979 $attempt_cache_expiry

    response=$(docker exec vault vault write -format=json \
            auth/circleci/login project=someproject build_num=100 \
            vcs_revision=babababababababababababababababababababa 2>&1 || true)

    #sleep 1

    #docker logs vault  2>&1 > vault.$i.log

    if [[ ${grep_expressions[$((i-1))]} ]]; then
        echo "$response" | grep -F "${grep_expressions[$((i-1))]}" > /dev/null
        echo "Test $i PASSED"
    else
        [[ $(echo "$response" | jq -r '.auth.client_token' | wc -c) -gt 0 ]]
        echo "Test $i PASSED"
    fi

    if (( $i == 3 )); then
        # Testing a second attempt at authenticating the same build
        response=$(docker exec vault \
                vault write \
                auth/circleci/login \
                project=someproject \
                build_num=100 \
                vcs_revision=babababababababababababababababababababa 2>&1 || true)

        echo "$response" | grep -F "an attempt to authenticate as this build has already been made" > /dev/null
        echo "Test 3b PASSED"

        # Testing that the cache actually gets cleared after some
        #   period of time.  The timeout has been set to 5s, but the
        #   cache is only cleaned every 60s, so we wait 90s to make
        #   sure that the cache has been cleaned.
        echo "Waiting 90 seconds so that expired Attempts cache entries have been cleaned..."
        sleep 90s

        response=$(docker exec vault vault write -format=json \
        auth/circleci/login project=someproject build_num=100 \
        vcs_revision=babababababababababababababababababababa 2>&1 || true)

        [[ $(echo "$response" | jq -r '.auth.client_token' | wc -c) -gt 0 ]]
        echo "Test 3c PASSED"
    fi

    echo -n "Removing docker containers: "
    docker rm -f circle vault | tr '\n' ' '
done
