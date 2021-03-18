#!/usr/bin/env bash

set -e

REPO_DIR="$(
    cd "$(dirname "$BASH_SOURCE[0]")/../" || return
    pwd
)"
VENDOR_DIR="${REPO_DIR}/vendor"
OUT_DIR="${REPO_DIR}/_output"
OUT_BIN="${OUT_DIR}/bin"
CMD_DIR="${REPO_DIR}/cmd"

GIT_VERSION=$(git describe --always --tags)
VERSION=${CI_UPSTREAM_VERSION:-${GIT_VERSION}}
GIT_COMMIT=$(git rev-list -1 HEAD)
COMMIT=${CI_UPSTREAM_COMMIT:-${GIT_COMMIT}}
BUILD_DATE=$(date --utc -Iseconds)

TARGET_GOOS="${TARGET_GOOS:-linux}"
TARGET_GOARCH="${TARGET_GOARCH:-amd64}"
