#!/bin/bash
set -euo pipefail

echo "Launching Vault with Prometheus..."
docker-compose up -d --force-recreate
sleep 1

echo "Configuring circleci authentication backend..."
curl -s -H "X-Vault-Token: root" \
    -d '{"circleci_token":"fake","base_url":"http://mock-circleci:7979/api/v1.1","vcs_type":"github","owner":"fred","ttl":"15m","max_ttl":"60m","attempt_cache_expiry":"5h"}' \
    http://localhost:8200/v1/auth/circleci/config

echo "Waiting 2 minutes to get a few scrapes in before starting stress test..."
sleep 120

echo "Starting test $(date)"
i=0
n=${NUM_ITERS:-1000}
while (( $i < $n )); do
    curl -s -d '{"project":"someproject","build_num":"'$i'","vcs_revision":"babababababababababababababababababababa"}' \
        http://localhost:8200/v1/auth/circleci/login > /dev/null
    i=$((i+1))
    echo -en "\r$((100*i/n )) %"
done
echo
echo "Test finished $(date)"