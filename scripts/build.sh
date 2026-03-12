#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
mkdir -p "${DIST_DIR}"

targets=(
  "darwin amd64"
  "darwin arm64"
  "linux amd64"
  "linux arm64"
  "windows amd64"
  "windows arm64"
)

for target in "${targets[@]}"; do
  read -r goos goarch <<<"${target}"
  suffix=""
  if [[ "${goos}" == "windows" ]]; then
    suffix=".exe"
  fi
  output="${DIST_DIR}/claw-remove-${goos}-${goarch}${suffix}"
  echo "==> ${goos}/${goarch}"
  GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" -o "${output}" ./cmd/claw-remove
done
