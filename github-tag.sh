#!/usr/bin/env bash
set -euo pipefail

failure=0
VERSION=$1

if [[ -z $VERSION ]]; then
    echo "ERROR: Missing version argument."
    exit 1
fi

# Get the current branch protection
protection=$(curl -s -H "Accept: application/vnd.github.v3+json" \
    https://api.github.com/repos/$USER/$EXECUTABLE/branches/master/protection)

# Delete the branch protection
curl -s -H "Accept: application/vnd.github.v3+json" \
    -X DELETE https://api.github.com/repos/$USER/$EXECUTABLE/branches/master/protection

git tag v$VERSION && git push --tags && failure=1 || true

curl -s -H "Accept: application/vnd.github.v3+json" \
    -d "$protection" -X PUT \
    https://api.github.com/repos/$USER/$EXECUTABLE/branches/master/protection

exit $failure
