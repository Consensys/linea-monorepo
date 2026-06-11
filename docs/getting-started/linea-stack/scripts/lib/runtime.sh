#!/usr/bin/env sh
# Shared runtime artifact, .env, JSON, and RPC helpers for Lineth quickstart scripts.
# POSIX sh. This file is safe to source under set -eu and does not print output.

lineth_runtime_init() {
  start="${1:-.}"
  if [ -f "$start" ]; then
    start="$(dirname "$start")"
  fi

  dir="$(CDPATH= cd "$start" 2>/dev/null && pwd -P)" || return 1
  while [ "$dir" != "/" ]; do
    if [ -f "$dir/docker-compose.yml" ] && [ -f "$dir/versions.env" ]; then
      LINETH_STACK_DIR="$dir"
      # Per-instance env file. Defaults to the classic .env; a second stack
      # instance points this at its own file (e.g. instances/i2.env). Relative
      # paths resolve from the stack dir.
      LINETH_ENV_FILE="${LINETH_ENV_FILE:-$LINETH_STACK_DIR/.env}"
      case "$LINETH_ENV_FILE" in
        /*) ;;
        *) LINETH_ENV_FILE="$LINETH_STACK_DIR/$LINETH_ENV_FILE" ;;
      esac
      export LINETH_STACK_DIR LINETH_ENV_FILE
      # Per-instance artifact root (compose interpolates the same variable for
      # its bind mounts). Defaults to ./artifacts as before.
      artifacts_dir="$(lineth_env_or_default LINETH_ARTIFACTS_DIR "$LINETH_STACK_DIR/artifacts")"
      case "$artifacts_dir" in
        /*) LINETH_ARTIFACTS_DIR="$artifacts_dir" ;;
        *) LINETH_ARTIFACTS_DIR="$LINETH_STACK_DIR/${artifacts_dir#./}" ;;
      esac
      LINETH_ACCOUNTS_DIR="$LINETH_ARTIFACTS_DIR/accounts"
      LINETH_GENESIS_DIR="$LINETH_ARTIFACTS_DIR/genesis"
      LINETH_CONFIG_DIR="$LINETH_ARTIFACTS_DIR/config"
      LINETH_DEPLOYMENTS_DIR="$LINETH_ARTIFACTS_DIR/deployments"
      LINETH_REPORTS_DIR="$LINETH_ARTIFACTS_DIR/reports"
      export LINETH_ARTIFACTS_DIR LINETH_ACCOUNTS_DIR LINETH_GENESIS_DIR
      export LINETH_CONFIG_DIR LINETH_DEPLOYMENTS_DIR LINETH_REPORTS_DIR
      return 0
    fi
    dir="$(dirname "$dir")"
  done

  printf 'lineth_runtime_init: could not find linea-stack root from %s\n' "$start" >&2
  return 1
}

lineth_valid_env_key() {
  case "$1" in
    ''|[0-9]*|*[!A-Za-z0-9_]*)
      return 1
      ;;
    *)
      return 0
      ;;
  esac
}

lineth_env_value() {
  key="$1"
  lineth_valid_env_key "$key" || return 1
  env_file="${LINETH_ENV_FILE:-${LINETH_STACK_DIR:-.}/.env}"
  [ -f "$env_file" ] || return 1
  sed -nE "s/^${key}=([^#[:space:]].*)$/\1/p" "$env_file" | tail -1
}

lineth_env_or_default() {
  key="$1"
  fallback="$2"
  lineth_valid_env_key "$key" || return 1
  eval "current=\${$key:-}"
  if [ -n "$current" ]; then
    printf '%s' "$current"
    return 0
  fi
  from_env="$(lineth_env_value "$key" || true)"
  if [ -n "$from_env" ]; then
    printf '%s' "$from_env"
  else
    printf '%s' "$fallback"
  fi
}

lineth_host_port() {
  lineth_env_or_default "$1" "$2"
}

lineth_l1_mode() {
  mode="$(lineth_env_or_default L1_MODE sepolia)"
  case "$mode" in
    sepolia|local)
      printf '%s' "$mode"
      ;;
    *)
      printf '%s' "$mode"
      return 1
      ;;
  esac
}

# Local-L1 instance role. "owner" (default) starts and owns the local L1;
# "attach" runs L2-only and anchors to an owner instance's L1 over the shared
# Docker network. Only meaningful when L1_MODE=local.
lineth_l1_local_role() {
  role="$(lineth_env_or_default L1_LOCAL_ROLE owner)"
  case "$role" in
    owner|attach)
      printf '%s' "$role"
      ;;
    *)
      printf '%s' "$role"
      return 1
      ;;
  esac
}

# Docker network an attach-role instance joins to reach the owner's L1.
# Default matches the owner's compose-created network: <owner-project>_l1network.
lineth_l1_attach_network() {
  lineth_env_or_default LINETH_L1_ATTACH_NETWORK "linea-stack_l1network"
}

# Per-instance container name prefix (compose renders container_name with the
# same variable). Empty for the default single-instance path.
lineth_container() {
  printf '%s%s' "$(lineth_env_or_default LINETH_CONTAINER_PREFIX "")" "$1"
}

# Canonical docker compose invocation for this instance: instance env file plus
# the external-L1 attach overlay when this instance runs in attach role.
# Callers append `--profile <name>` and the subcommand. Paths are relative to
# the stack dir except LINETH_ENV_FILE, which runtime-init absolutizes.
lineth_compose_cmd() {
  cmd="docker compose --env-file versions.env --env-file ${LINETH_ENV_FILE:-.env}"
  if [ "$(lineth_l1_mode || true)" = "local" ] && [ "$(lineth_l1_local_role || true)" = "attach" ]; then
    cmd="$cmd -f docker-compose.yml -f docker-compose.l1-attach.yml"
  fi
  printf '%s' "$cmd"
}

lineth_l1_host_rpc_url() {
  mode="$(lineth_l1_mode || true)"
  if [ "$mode" = "local" ]; then
    printf 'http://localhost:%s' "$(lineth_host_port HOST_PORT_L1_RPC 8445)"
  else
    lineth_env_or_default L1_RPC_URL ""
  fi
}

lineth_l1_container_rpc_url() {
  mode="$(lineth_l1_mode || true)"
  if [ "$mode" = "local" ]; then
    printf '%s' "http://l1-el-node:8545"
  else
    lineth_env_or_default L1_RPC_URL ""
  fi
}

lineth_l1_deployer_shell_env() {
  mode="$(lineth_l1_mode || true)"
  if [ "$mode" = "local" ]; then
    local_default_key='0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80'
    local_key="$(lineth_env_or_default L1_LOCAL_DEPLOYER_PRIVATE_KEY "$local_default_key")"
    local_addr="$(lineth_env_or_default L1_LOCAL_DEPLOYER_ADDRESS "")"
    if [ "$local_key" = "$local_default_key" ] && [ -z "$local_addr" ]; then
      local_addr='0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266'
    fi
    if [ -n "$local_addr" ]; then
      printf "L1_MODE='local'\n"
      printf "L1_RPC_URL='http://localhost:%s'\n" "$(lineth_host_port HOST_PORT_L1_RPC 8445)"
      printf "L1_DEPLOYER_ADDRESS='%s'\n" "$local_addr"
      printf "L1_DEPLOYER_SOURCE='local-genesis'\n"
      printf "L1_DEPLOYER_PRIVATE_KEY='%s'\n" "$local_key"
      return 0
    fi
    # L1_LOCAL_DEPLOYER_PRIVATE_KEY overridden without a companion
    # L1_LOCAL_DEPLOYER_ADDRESS: derive the address via ts-node below.
  fi

  root="$(git -C "$LINETH_STACK_DIR" rev-parse --show-toplevel 2>/dev/null || true)"
  if [ -n "$root" ] && [ -x "$root/node_modules/.bin/ts-node" ]; then
    (
      cd "$root/contracts"
      export NODE_PATH="$root/node_modules:$root/contracts/node_modules${NODE_PATH:+:$NODE_PATH}"
      export LINETH_STACK_DIR
      export LINETH_ENV_FILE
      TS_NODE_TRANSPILE_ONLY=1 \
      TS_NODE_COMPILER_OPTIONS='{"module":"CommonJS","moduleResolution":"Node"}' \
        pnpm -s exec ts-node "$LINETH_STACK_DIR/scripts/internal/deployer-wallet.ts" emit-shell-env --context host
    )
    return $?
  fi

  if [ "$mode" = "local" ]; then
    printf 'host ts-node is required to derive the address for an overridden L1_LOCAL_DEPLOYER_PRIVATE_KEY; run pnpm install or set L1_LOCAL_DEPLOYER_ADDRESS alongside it\n' >&2
    return 1
  fi

  legacy_key="$(lineth_env_or_default L1_DEPLOYER_PRIVATE_KEY "")"
  if [ -n "$legacy_key" ]; then
    printf "L1_MODE='sepolia'\n"
    printf "L1_RPC_URL='%s'\n" "$(lineth_l1_host_rpc_url | sed "s/'/'\\\\''/g")"
    printf "L1_DEPLOYER_ADDRESS=''\n"
    printf "L1_DEPLOYER_SOURCE='legacy-raw-key'\n"
    printf "L1_DEPLOYER_PRIVATE_KEY='%s'\n" "$(printf '%s' "$legacy_key" | sed "s/'/'\\\\''/g")"
    return 0
  fi

  printf 'host ts-node is required to resolve the generated Sepolia deployer; run pnpm install or use ./scripts/start.sh first\n' >&2
  return 1
}

lineth_l1_deployer_private_key() {
  set +x 2>/dev/null || true
  eval "$(lineth_l1_deployer_shell_env)"
  printf '%s' "$L1_DEPLOYER_PRIVATE_KEY"
}

lineth_l1_address_link() {
  addr="$1"
  [ -n "$addr" ] || return 0
  if [ "$(lineth_l1_mode || true)" = "local" ]; then
    printf 'local L1 address %s rpc=%s' "$addr" "$(lineth_l1_host_rpc_url)"
  else
    printf 'https://sepolia.etherscan.io/address/%s' "$addr"
  fi
}

lineth_l1_tx_link() {
  tx_hash="$1"
  [ -n "$tx_hash" ] || return 0
  if [ "$(lineth_l1_mode || true)" = "local" ]; then
    printf 'local L1 tx %s rpc=%s' "$tx_hash" "$(lineth_l1_host_rpc_url)"
  else
    printf 'https://sepolia.etherscan.io/tx/%s' "$tx_hash"
  fi
}

lineth_accounts_file() {
  printf '%s/%s\n' "$LINETH_ACCOUNTS_DIR" "$1"
}

lineth_genesis_file() {
  printf '%s/%s\n' "$LINETH_GENESIS_DIR" "$1"
}

lineth_config_file() {
  printf '%s/%s\n' "$LINETH_CONFIG_DIR" "$1"
}

lineth_deployments_file() {
  printf '%s/%s\n' "$LINETH_DEPLOYMENTS_DIR" "$1"
}

lineth_reports_file() {
  printf '%s/%s\n' "$LINETH_REPORTS_DIR" "$1"
}

lineth_artifact_file() {
  case "$1" in
    addresses-precomputed.json|runtime-keys.env|demo-traffic.env)
      lineth_accounts_file "$1"
      ;;
    genesis-besu.json|genesis-maru.json|fork-timestamp.txt)
      lineth_genesis_file "$1"
      ;;
    *)
      lineth_deployments_file "$1"
      ;;
  esac
}

lineth_artifact_exists() {
  [ -f "$(lineth_artifact_file "$1")" ]
}

lineth_artifact_section_addr() {
  file="$1"
  section="$2"
  key="$3"
  lineth_json_section_addr "$(lineth_artifact_file "$file")" "$section" "$key"
}

lineth_require_file() {
  path="$1"
  message="$2"
  [ -s "$path" ] && return 0
  printf '%s\n' "$message" >&2
  return 1
}

lineth_json_section_addr() {
  file="$1"
  section="$2"
  key="$3"
  [ -s "$file" ] || return 0
  sed -nE "/\"$section\"[[:space:]]*:/,/^[[:space:]]*}/ s/.*\"$key\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" "$file" | head -1
}

lineth_json_root_addr() {
  file="$1"
  key="$2"
  [ -s "$file" ] || return 0
  sed -nE "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"(0x[a-fA-F0-9]{40})\".*/\1/p" "$file" | head -1
}

lineth_json_root_value() {
  file="$1"
  key="$2"
  [ -s "$file" ] || return 0
  sed -nE "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"?([^\",}]*)\"?.*/\1/p" "$file" | head -1
}

lineth_json_meta_value() {
  file="$1"
  key="$2"
  [ -s "$file" ] || return 0
  sed -nE "/\"_meta\"[[:space:]]*:/,/^[[:space:]]*}/ s/.*\"$key\"[[:space:]]*:[[:space:]]*\"?([^\",}]+)\"?.*/\1/p" "$file" | head -1
}

lineth_json_stdin_string_field() {
  key="$1"
  sed -nE "s/.*\"${key}\"[[:space:]]*:[[:space:]]*\"([^\"]*)\".*/\1/p" | head -1
}

lineth_json_stdin_number_field() {
  key="$1"
  sed -nE "s/.*\"$key\"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p" | head -1
}

lineth_json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

lineth_json_value() {
  value="$1"
  if [ -n "$value" ]; then
    printf '"%s"' "$(lineth_json_escape "$value")"
  else
    printf 'null'
  fi
}

lineth_json_bool() {
  case "$1" in
    true) printf 'true' ;;
    *) printf 'false' ;;
  esac
}

lineth_is_address() {
  printf '%s\n' "$1" | grep -qE '^0x[a-fA-F0-9]{40}$'
}

lineth_is_hash() {
  printf '%s\n' "$1" | grep -qE '^0x[a-fA-F0-9]{64}$'
}

lineth_is_uint() {
  case "$1" in
    ''|*[!0-9]*) return 1 ;;
    *) return 0 ;;
  esac
}

lineth_require_address() {
  label="$1"
  value="$2"
  lineth_is_address "$value" && return 0
  printf '%s missing or invalid\n' "$label" >&2
  return 1
}

lineth_require_hash() {
  label="$1"
  value="$2"
  lineth_is_hash "$value" && return 0
  printf '%s missing or invalid\n' "$label" >&2
  return 1
}

lineth_require_uint() {
  label="$1"
  value="$2"
  lineth_is_uint "$value" && return 0
  printf '%s must be a non-negative integer\n' "$label" >&2
  return 1
}

lineth_hex_to_dec_small() {
  hex="$1"
  hex="${hex#0x}"
  [ -n "$hex" ] || { echo "0"; return; }
  printf '%d\n' "$((16#$hex))" 2>/dev/null || printf '0x%s\n' "$hex"
}

lineth_rpc_json() {
  url="$1"
  method="$2"
  params="$3"
  [ -n "$url" ] || return 0
  curl -sS -X POST -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"$method\",\"params\":$params}" \
    "$url" 2>/dev/null || true
}
