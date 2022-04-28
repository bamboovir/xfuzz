#!/usr/bin/env bash

set -o pipefail
set -o errexit
set -o xtrace

if ! [ -x "$(command -v git)" ]; then
    printf "%s\n" 'Error: git is not installed.' >&2
    exit 1
fi

if ! [ -x "$(command -v go)" ]; then
    printf "%s\n" 'Error: go is not installed.' >&2
    exit 1
fi
PROJECT_ROOT=$(git rev-parse --show-toplevel)

build() {
    local bin="$1"; shift
    go build -tags "" -mod=vendor -o "${PROJECT_ROOT}/bin/${bin}" "${PROJECT_ROOT}/case/${bin}"
}

build "panic"
build "segmentation_fault"
