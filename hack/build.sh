#!/usr/bin/env bash

set -e

# shellcheck source=common.sh
source "$(dirname "$0")/common.sh"

mkdir -p "${OUT_BIN}"
LDFLAGS="-s -w  \
-X github.com/cynepco3ahue/pod-restarter/pkg/version.Version=${VERSION} \
-X github.com/cynepco3ahue/pod-restarter/pkg/version.GitCommit=${COMMIT} \
-X github.com/cynepco3ahue/pod-restarter/pkg/version.BuildDate=${BUILD_DATE}"
GOOS=${TARGET_GOOS} GOARCH=${TARGET_GOARCH} go build -i -ldflags="${LDFLAGS}" -mod=vendor -o "${OUT_BIN}/pod-restarter" "${CMD_DIR}/pod-restarter"
