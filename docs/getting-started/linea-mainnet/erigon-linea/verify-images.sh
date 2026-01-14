#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 2 ]; then
  echo "Usage: $0 <image_a> <image_b>"
  echo "Example: $0 morislineats/erigon:v3.3.0-linea-patched consensys/linea-erigon:v3.3.0-linea-patched"
  exit 2
fi

img_a="$1"
img_b="$2"
platform="${PLATFORM:-linux/amd64}"

echo "Pulling images..."
docker pull --platform="$platform" "$img_a" >/dev/null
docker pull --platform="$platform" "$img_b" >/dev/null

erigon_meta() {
  local img="$1"
  echo "== $img =="
  docker run --platform="$platform" --rm --entrypoint /usr/local/bin/erigon "$img" version 2>/dev/null || \
    docker run --platform="$platform" --rm --entrypoint /usr/local/bin/erigon "$img" --version 2>/dev/null || true
  docker run --platform="$platform" --rm --entrypoint /bin/sh "$img" -lc 'set -e; command -v sha256sum >/dev/null 2>&1 || { echo "sha256sum missing"; exit 1; }; sha256sum /usr/local/bin/erigon | awk "{print \$1}"'
}

sum_a="$(docker run --platform="$platform" --rm --entrypoint /bin/sh "$img_a" -lc 'sha256sum /usr/local/bin/erigon | awk "{print \$1}"')"
sum_b="$(docker run --platform="$platform" --rm --entrypoint /bin/sh "$img_b" -lc 'sha256sum /usr/local/bin/erigon | awk "{print \$1}"')"

erigon_meta "$img_a"
erigon_meta "$img_b"

echo
echo "Binary sha256:"
echo "  $img_a  $sum_a"
echo "  $img_b  $sum_b"

if [ "$sum_a" = "$sum_b" ]; then
  echo "OK: binaries are identical."
  exit 0
else
  echo "MISMATCH: binaries differ. Do not assume the images are the same."
  exit 1
fi

