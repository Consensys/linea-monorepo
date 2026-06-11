#!/usr/bin/env sh
# Stop the quickstart and remove both Docker volumes and host-backed artifacts.
set -eu

SCRIPT_DIR="$(CDPATH= cd "$(dirname "$0")" && pwd -P)"
STACK_DIR="$(CDPATH= cd "$SCRIPT_DIR/.." && pwd -P)"
LINETH_LOG_CONTEXT="reset"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/logging.sh"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/runtime.sh"

RESET_LOCAL_L1=false
FORGET_DEPLOYER=false

usage() {
  cat <<'EOF'
Usage: ./scripts/reset.sh [--local-l1] [--forget-deployer]

  --local-l1          also remove the quickstart local L1 data volume
  --forget-deployer  remove the generated Sepolia deployer keystore
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --local-l1)
      RESET_LOCAL_L1=true
      ;;
    --forget-deployer)
      FORGET_DEPLOYER=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      lineth_die "unknown argument: $1"
      ;;
  esac
  shift
done

cd "$STACK_DIR"
lineth_runtime_init "$STACK_DIR"
COMPOSE="$(lineth_compose_cmd) --profile stack-partial-prover"
COMPOSE_PROJECT="$(lineth_env_or_default COMPOSE_PROJECT_NAME linea-stack)"
L1_LOCAL_ROLE="$(lineth_l1_local_role || true)"
if [ "$(lineth_l1_mode || true)" = "local" ] && [ "$L1_LOCAL_ROLE" != "attach" ]; then
  RESET_LOCAL_L1=true
fi

lineth_banner "reset · clean quickstart state"

lineth_section "docker compose"
if [ "$RESET_LOCAL_L1" = "true" ]; then
  COMPOSE="$COMPOSE --profile local-l1"
fi
# Attach-role instances reference the owner's external L1 network; if the
# owner is already gone, fall back to a plain compose file so `down` still works.
if [ "$L1_LOCAL_ROLE" = "attach" ] \
  && ! docker network inspect "$(lineth_l1_attach_network)" >/dev/null 2>&1; then
  COMPOSE="docker compose --env-file versions.env --env-file $LINETH_ENV_FILE --profile stack-partial-prover"
fi
# shellcheck disable=SC2086
$COMPOSE down -v --remove-orphans
# Pre-refactor legacy volumes only ever existed for the single default instance.
docker volume rm \
  linea-stack-shared-config \
  linea-stack-l2-genesis \
  linea-stack-rendered-config \
  linea-stack-postman-runtime-config >/dev/null 2>&1 || true
if [ "$RESET_LOCAL_L1" = "true" ]; then
  docker volume rm "${COMPOSE_PROJECT}-local-l1-data" >/dev/null 2>&1 || true
fi

lineth_section "host artifacts"
DEPLOYER_KEYSTORE_DIR="$LINETH_ARTIFACTS_DIR/accounts/deployer-keystore"
PRESERVED_DEPLOYER_DIR=""
if [ "$FORGET_DEPLOYER" != "true" ] && [ -d "$DEPLOYER_KEYSTORE_DIR" ]; then
  PRESERVED_DEPLOYER_DIR="$(mktemp -d)"
  cp -a "$DEPLOYER_KEYSTORE_DIR"/. "$PRESERVED_DEPLOYER_DIR"/
fi

rm -rf \
  "$LINETH_ARTIFACTS_DIR/accounts" \
  "$LINETH_ARTIFACTS_DIR/genesis" \
  "$LINETH_ARTIFACTS_DIR/config" \
  "$LINETH_ARTIFACTS_DIR/deployments" \
  "$LINETH_ARTIFACTS_DIR/reports"

if [ -n "$PRESERVED_DEPLOYER_DIR" ]; then
  mkdir -p "$DEPLOYER_KEYSTORE_DIR"
  cp -a "$PRESERVED_DEPLOYER_DIR"/. "$DEPLOYER_KEYSTORE_DIR"/
  rm -rf "$PRESERVED_DEPLOYER_DIR"
  lineth_ok "removed generated host artifacts; preserved Sepolia deployer keystore"
elif [ "$FORGET_DEPLOYER" = "true" ]; then
  lineth_ok "removed generated host artifacts, including Sepolia deployer keystore"
else
  lineth_ok "removed generated host artifacts"
fi
