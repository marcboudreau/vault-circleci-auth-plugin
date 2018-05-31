#!/usr/bin/env bash
set -euo pipefail

if (git describe --abbrev=0 --exact-match &> /dev/null); then
    git describe --abbrev=0 --exact-match | sed 's/v\(.*\)/\1/'
else
    tags=$(git rev-list --tags --max-count=1 2> /dev/null || true)
    if [[ -z $tags ]]; then
        v="0.0.0"
    else
        v=$(git describe --abbrev=0 --tags $tags 2> /dev/null | sed 's/v\(.*\)/\1/')
    fi

    a=( ${v//./ } )
    (( a[2]++ ))

    echo "${a[0]}.${a[1]}.${a[2]}-${CIRCLECI_BUILD_NUM:-dev}"
fi