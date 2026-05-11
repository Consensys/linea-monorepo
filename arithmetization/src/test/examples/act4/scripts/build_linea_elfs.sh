#!/bin/bash
# Build the ACT4 self-checking ELFs for the Linea (RV64IM+Zicclsm) config.
#
# Uses the pre-built docker image from the riscv-arch-test repo
# (default tag `riscv-act4:latest`, built via `docker build -t riscv-act4 .`
# in the riscv-arch-test checkout).
#
# Inputs:
#   ACT4_CONFIG_DIR     directory containing `linea-rv64im-zicclsm/...`
#                       (i.e. `<...>/zkevm-test-monitor/act4-configs/linea`)
#   ACT4_WORK_DIR       output directory (writable; ELFs land in `linea-rv64im-zicclsm/elfs`)
#   ACT4_IMAGE          docker image tag    (default: riscv-act4:latest)
#   ACT4_EXTENSIONS     comma-list of extensions to build (default: I,M)
#   ACT4_JOBS           parallel build jobs (default: 4)
#
# Output: ELFs in $ACT4_WORK_DIR/linea-rv64im-zicclsm/elfs/

set -u

SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" && pwd)
ACT4_DIR=$(cd -- "$SCRIPT_DIR/.." && pwd)

# Default ACT4_WORK_DIR lives inside this script's tree, so reproducing
# the run doesn't depend on the user's filesystem layout. Default
# ACT4_CONFIG_DIR walks up to a peer `zkevm-test-monitor/act4-configs/`
# checkout next to `linea-monorepo` (the layout described in README.md).
DEFAULT_WORK_DIR="$ACT4_DIR/bin/work"
DEFAULT_CONFIG_DIR="$ACT4_DIR/../../../../../../zkevm-test-monitor/act4-configs/linea"

ACT4_WORK_DIR="${ACT4_WORK_DIR:-$DEFAULT_WORK_DIR}"
ACT4_CONFIG_DIR="${ACT4_CONFIG_DIR:-$DEFAULT_CONFIG_DIR}"
ACT4_IMAGE="${ACT4_IMAGE:-riscv-act4:latest}"
ACT4_EXTENSIONS="${ACT4_EXTENSIONS:-I,M}"
ACT4_JOBS="${ACT4_JOBS:-4}"

# Resolve to absolute paths (docker requires absolute mount sources).
mkdir -p "$ACT4_WORK_DIR"
ACT4_WORK_DIR=$(cd -- "$ACT4_WORK_DIR" && pwd)

if [ ! -d "$ACT4_CONFIG_DIR/linea-rv64im-zicclsm" ]; then
    echo "error: '$ACT4_CONFIG_DIR/linea-rv64im-zicclsm' does not exist." >&2
    echo "Set ACT4_CONFIG_DIR to the directory that contains linea-rv64im-zicclsm/." >&2
    exit 2
fi
ACT4_CONFIG_DIR=$(cd -- "$ACT4_CONFIG_DIR" && pwd)

if ! docker image inspect "$ACT4_IMAGE" >/dev/null 2>&1; then
    echo "error: docker image '$ACT4_IMAGE' not found." >&2
    echo "Build it from the riscv-arch-test repo:" >&2
    echo "  cd /path/to/riscv-arch-test && docker build -t $ACT4_IMAGE ." >&2
    exit 2
fi

set -x
docker run --rm \
    -v "$ACT4_CONFIG_DIR:/act4/config/cores/linea:ro" \
    -v "$ACT4_WORK_DIR:/act4/work" \
    "$ACT4_IMAGE" \
    bash -c "rm -rf /act4/work/linea-rv64im-zicclsm && \
             CONFIG_FILES=config/cores/linea/linea-rv64im-zicclsm/test_config.yaml \
             EXTENSIONS=$ACT4_EXTENSIONS \
             FAST=True \
             make --jobs $ACT4_JOBS --keep-going"
