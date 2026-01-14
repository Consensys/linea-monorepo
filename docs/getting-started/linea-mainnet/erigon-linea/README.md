# Linea-patched Erigon image

This folder contains the build recipe for the **Linea-patched Erigon** image currently referenced by the Linea getting-started compose files.

## What’s different vs upstream `erigontech/erigon:v3.3.0`

The Dockerfile clones upstream `erigontech/erigon` tag `v3.3.0` and applies 3 tiny patches + builds without Silkworm:

- Disable the block downloader v2 (stability on slower machines / EC2)
- Return `nil` instead of error on empty code for **EIP-7002** (Linea has empty withdrawal contract)
- Return `nil` instead of error on empty code for **EIP-7251** (Linea has empty consolidation contract)
- Build with `BUILD_TAGS=nosilkworm`

See `Dockerfile` for exact patch lines.

## Build locally (load into Docker)

```bash
cd docs/getting-started/linea-mainnet/erigon-linea
make build IMAGE=consensys/linea-erigon TAG=v3.3.0-linea-patched
```

## Push to ConsenSys Docker Hub

You need permissions to push to the ConsenSys Docker Hub org:

```bash
docker login
cd docs/getting-started/linea-mainnet/erigon-linea
make push IMAGE=consensys/linea-erigon TAG=v3.3.0-linea-patched
```

## Verify it matches the previous image (fast, no node testing)

If you want to be sure the **actual `erigon` binary** is identical between images:

```bash
cd docs/getting-started/linea-mainnet/erigon-linea
chmod +x verify-images.sh
./verify-images.sh morislineats/erigon:v3.3.0-linea-patched consensys/linea-erigon:v3.3.0-linea-patched
```

This compares the `sha256` of `/usr/local/bin/erigon` in both images.
If you’re on an ARM machine (e.g. Apple Silicon / Graviton), the script forces `PLATFORM=linux/amd64` by default.

## Suggested tagging

- `consensys/linea-erigon:v3.3.0-linea-patched`
- Optionally also add a short git SHA tag for traceability, e.g. `:v3.3.0-linea-patched-a0c55b44`

