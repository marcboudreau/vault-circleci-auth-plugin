#!/usr/bin/env bash
set -e

XC_OSARCH=${XC_OSARCH:-"linux/386 linux/amd64 darwin/386 darwin/amd64 windows/386 windows/amd64 freebsd/386 freebsd/amd64 freebsd/arm openbsd/386 openbsd/amd64 openbsd/arm netbsd/386 netbsd/amd64 solaris/amd64"}

for XC in $XC_OSARCH; do
    array=( ${XC/\// } )
    export GOOS=${array[0]}
    export GOARCH=${array[1]}
    go build -o ./pkg/${GOOS}_${GOARCH}/vault-circleci-auth-plugin .
done

for platform in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
    OSARCH=$(basename $platform)

    pushd $platform >/dev/null 2>&1
    zip ../${OSARCH}.zip ./*
    popd >/dev/null 2>&1
done