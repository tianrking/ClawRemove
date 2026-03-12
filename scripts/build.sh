#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
mkdir -p "${DIST_DIR}"

VERSION="${CLAWREMOVE_VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"

targets=(
  "darwin amd64"
  "darwin arm64"
  "linux amd64"
  "linux arm64"
  "windows amd64"
  "windows arm64"
)

# Clear existing checksums
> "${DIST_DIR}/sha256sums.txt"

for target in "${targets[@]}"; do
  read -r goos goarch <<<"${target}"
  suffix=""
  archive_ext=".tar.gz"
  if [[ "${goos}" == "windows" ]]; then
    suffix=".exe"
    archive_ext=".zip"
  fi
  
  binary_name="claw-remove-${goos}-${goarch}${suffix}"
  output="${DIST_DIR}/${binary_name}"
  
  echo "==> Building ${goos}/${goarch} (Version: ${VERSION})"
  GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w -X github.com/tianrking/ClawRemove/internal/app.Version=${VERSION}" -o "${output}" ./cmd/claw-remove
    
  echo "==> Packaging ${goos}/${goarch}"
  archive_name="claw-remove-${goos}-${goarch}${archive_ext}"
  cd "${DIST_DIR}"
  if [[ "${goos}" == "windows" ]]; then
    zip -q "${archive_name}" "${binary_name}"
  else
    tar -czf "${archive_name}" "${binary_name}"
  fi
  
  # Checksum
  shasum -a 256 "${archive_name}" >> "sha256sums.txt"
  cd "${ROOT_DIR}"
done

echo "==> Release artifacts generated in ${DIST_DIR}"
