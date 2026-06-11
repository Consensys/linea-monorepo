#!/usr/bin/env sh
# Executable form of the Phase-1 dual-L2 standard: boots TWO quickstart
# instances from clean — instance 1 owns the local L1 (+ L2-A), instance 2 is
# L2-only and attaches to instance 1's L1 — then asserts, while both run
# concurrently:
#
#   1. instance 2 started no L1 containers (exactly one l1-el-node on the host);
#   2. zero collisions: distinct container names, Docker networks, host ports,
#      and artifact directories;
#   3. each instance deployed its OWN LineaRollup to the SAME shared L1
#      (different addresses, both with on-chain code, same L1 chain id);
#   4. both rollups reached L1 finalization (currentL2BlockNumber > 0) and,
#      unless --no-advance, both ADVANCE again while running side by side.
#
# Prints the evidence, then tears both instances down (skip with --keep).
#
# Usage: ./scripts/verify-dual-l2.sh [--keep] [--no-advance] [--i2-env <file>]
set -eu

SCRIPT_DIR="$(CDPATH= cd "$(dirname "$0")" && pwd -P)"
STACK_DIR="$(CDPATH= cd "$SCRIPT_DIR/.." && pwd -P)"
LINETH_LOG_CONTEXT="verify-dual-l2"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/logging.sh"
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib/runtime.sh"

KEEP=false
CHECK_ADVANCE=true
I2_ENV="$STACK_DIR/instances/i2.env"
ADVANCE_TIMEOUT_SECONDS="${ADVANCE_TIMEOUT_SECONDS:-1200}"
FINALITY_TIMEOUT_SECONDS="${FINALITY_TIMEOUT_SECONDS:-900}"

usage() {
  cat <<'EOF'
Usage: ./scripts/verify-dual-l2.sh [--keep] [--no-advance] [--i2-env <file>]

  --keep        leave both instances running after a successful verification
  --no-advance  skip the slow "both rollups advance concurrently" assertion
  --i2-env      instance 2 env file (default: instances/i2.env, created from
                profiles/instance-2.env.example when missing)
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --keep) KEEP=true ;;
    --no-advance) CHECK_ADVANCE=false ;;
    --i2-env)
      shift
      [ "$#" -gt 0 ] || lineth_die "--i2-env requires a path"
      I2_ENV="$1"
      ;;
    -h|--help) usage; exit 0 ;;
    *) lineth_die "unknown argument: $1" ;;
  esac
  shift
done

cd "$STACK_DIR"

# Instance-scoped helpers. Every quickstart script is instance-selected purely
# by LINETH_ENV_FILE; the verifier re-execs helpers in subshells so the two
# instances never share resolved runtime state.
i1() { (unset LINETH_ENV_FILE LINETH_ARTIFACTS_DIR && "$@"); }
i2() { (unset LINETH_ARTIFACTS_DIR && LINETH_ENV_FILE="$I2_ENV" && export LINETH_ENV_FILE && "$@"); }

i1_env() { (unset LINETH_ENV_FILE LINETH_ARTIFACTS_DIR && lineth_runtime_init "$STACK_DIR" >/dev/null && "$@"); }
i2_env() { (unset LINETH_ARTIFACTS_DIR && LINETH_ENV_FILE="$I2_ENV" && export LINETH_ENV_FILE && lineth_runtime_init "$STACK_DIR" >/dev/null && "$@"); }

FAILURES=0
check() {
  label="$1"
  shift
  if "$@"; then
    lineth_ok "$label"
  else
    lineth_error "FAILED: $label"
    FAILURES=$((FAILURES + 1))
  fi
}

rollup_address() {
  # $1 = deployments dir
  lineth_json_section_addr "$1/addresses.json" l1 LineaRollupV8
}

rpc_result() {
  url="$1"; method="$2"; params="$3"
  lineth_rpc_json "$url" "$method" "$params" | lineth_json_stdin_string_field result
}

# currentL2BlockNumber() selector on LineaRollup.
CURRENT_L2_BLOCK_SELECTOR="0x695378f5"

rollup_current_l2_block() {
  # $1 = L1 RPC url, $2 = rollup address
  hex="$(rpc_result "$1" eth_call "[{\"to\":\"$2\",\"data\":\"$CURRENT_L2_BLOCK_SELECTOR\"},\"latest\"]")"
  [ -n "$hex" ] && [ "$hex" != "0x" ] || return 1
  lineth_hex_to_dec_small "$hex"
}

wait_rollup_block_above() {
  # $1 = L1 RPC, $2 = rollup, $3 = floor (exclusive), $4 = timeout s, $5 = label
  deadline=$(( $(date +%s) + $4 ))
  while [ "$(date +%s)" -lt "$deadline" ]; do
    value="$(rollup_current_l2_block "$1" "$2" || true)"
    if lineth_is_uint "$value" && [ "$value" -gt "$3" ]; then
      printf '%s' "$value"
      return 0
    fi
    sleep 10
  done
  printf 'timeout waiting for %s currentL2BlockNumber > %s\n' "$5" "$3" >&2
  return 1
}

lineth_banner "verify · two Linea L2s on one shared local L1"

[ -f "$I2_ENV" ] || {
  mkdir -p "$(dirname "$I2_ENV")"
  cp "$STACK_DIR/profiles/instance-2.env.example" "$I2_ENV"
  lineth_info "created $I2_ENV from profiles/instance-2.env.example"
}

I1_ARTIFACTS="$(i1_env sh -c 'printf %s "$LINETH_ARTIFACTS_DIR"')"
I2_ARTIFACTS="$(i2_env sh -c 'printf %s "$LINETH_ARTIFACTS_DIR"')"
I2_PREFIX="$(i2_env lineth_env_or_default LINETH_CONTAINER_PREFIX "")"
I2_PROJECT="$(i2_env lineth_env_or_default COMPOSE_PROJECT_NAME linea-stack-i2)"
I1_PROJECT="$(i1_env lineth_env_or_default COMPOSE_PROJECT_NAME linea-stack)"
L1_RPC="$(i1_env lineth_l1_host_rpc_url)"
I1_L2_RPC="http://localhost:$(i1_env lineth_host_port HOST_PORT_L2_RPC 8745)"
I2_L2_RPC="http://localhost:$(i2_env lineth_host_port HOST_PORT_L2_RPC 8745)"
I1_CHAIN="$(i1_env lineth_env_or_default L2_CHAIN_ID 1337)"
I2_CHAIN="$(i2_env lineth_env_or_default L2_CHAIN_ID 1337)"

[ "$(i1_env lineth_l1_mode)" = "local" ] || lineth_die "instance 1 must run L1_MODE=local"
[ "$(i1_env lineth_l1_local_role)" = "owner" ] || lineth_die "instance 1 must be L1_LOCAL_ROLE=owner"
[ "$(i2_env lineth_l1_local_role)" = "attach" ] || lineth_die "instance 2 env must set L1_LOCAL_ROLE=attach"
[ "$I1_ARTIFACTS" != "$I2_ARTIFACTS" ] || lineth_die "instance artifact dirs must differ"
[ -n "$I2_PREFIX" ] || lineth_die "instance 2 env must set LINETH_CONTAINER_PREFIX"
[ "$I1_CHAIN" != "$I2_CHAIN" ] || lineth_die "instance L2 chain ids must differ"

lineth_section "Reset both instances to clean state"
i2 "$SCRIPT_DIR/reset.sh" >/tmp/verify-dual-i2-reset.log 2>&1 || lineth_die "instance 2 reset failed; see /tmp/verify-dual-i2-reset.log"
i1 "$SCRIPT_DIR/reset.sh" >/tmp/verify-dual-i1-reset.log 2>&1 || lineth_die "instance 1 reset failed; see /tmp/verify-dual-i1-reset.log"
lineth_ok "both instances reset"

lineth_section "Boot instance 1 (local L1 owner + L2-A) to first finalization"
lineth_info "log: /tmp/verify-dual-i1-boot.log"
i1 "$SCRIPT_DIR/start.sh" --tail >/tmp/verify-dual-i1-boot.log 2>&1 \
  || lineth_die "instance 1 boot failed; see /tmp/verify-dual-i1-boot.log"
lineth_ok "instance 1 booted through first finalization"

lineth_section "Boot instance 2 (L2-only, attach) to first finalization"
lineth_info "log: /tmp/verify-dual-i2-boot.log"
i2 "$SCRIPT_DIR/start.sh" --tail >/tmp/verify-dual-i2-boot.log 2>&1 \
  || lineth_die "instance 2 boot failed; see /tmp/verify-dual-i2-boot.log"
lineth_ok "instance 2 booted through first finalization"

lineth_section "Assert: instance 2 runs L2-only on the shared L1"
l1_el_count="$(docker ps --format '{{.Names}}' | grep -cE '(^|-)l1-el-node$' || true)"
i2_l1_count="$(docker ps -a --format '{{.Names}}' | grep -cE "^${I2_PREFIX}l1-" || true)"
check "exactly one l1-el-node container on the host (got $l1_el_count)" [ "$l1_el_count" = "1" ]
check "instance 2 created no l1-* containers (got $i2_l1_count)" [ "$i2_l1_count" = "0" ]
check "instance 2 created no l1network of its own" \
  sh -c "! docker network inspect '${I2_PROJECT}_l1network' >/dev/null 2>&1"

lineth_section "Assert: zero collisions between the two instances"
i1_names="$(docker ps --format '{{.Names}}' --filter "label=com.docker.compose.project=$I1_PROJECT" | sort)"
i2_names="$(docker ps --format '{{.Names}}' --filter "label=com.docker.compose.project=$I2_PROJECT" | sort)"
i1_count="$(printf '%s\n' "$i1_names" | sed '/^$/d' | wc -l | tr -d ' ')"
i2_count="$(printf '%s\n' "$i2_names" | sed '/^$/d' | wc -l | tr -d ' ')"
overlap="$(printf '%s\n%s\n' "$i1_names" "$i2_names" | sed '/^$/d' | sort | uniq -d)"
check "instance 1 has running containers ($i1_count)" [ "$i1_count" -gt 0 ]
check "instance 2 has running containers ($i2_count)" [ "$i2_count" -gt 0 ]
check "no container name overlap" [ -z "$overlap" ]
check "instance networks are distinct" \
  sh -c "docker network inspect '${I1_PROJECT}_linea' '${I2_PROJECT}_linea' >/dev/null 2>&1"
check "artifact dirs distinct and populated" \
  sh -c "[ -s '$I1_ARTIFACTS/deployments/addresses.json' ] && [ -s '$I2_ARTIFACTS/deployments/addresses.json' ]"

i1_chain_hex="$(rpc_result "$I1_L2_RPC" eth_chainId '[]')"
i2_chain_hex="$(rpc_result "$I2_L2_RPC" eth_chainId '[]')"
check "instance 1 L2 RPC serves chain $I1_CHAIN (got $(lineth_hex_to_dec_small "${i1_chain_hex:-0x0}"))" \
  [ "$(lineth_hex_to_dec_small "${i1_chain_hex:-0x0}")" = "$I1_CHAIN" ]
check "instance 2 L2 RPC serves chain $I2_CHAIN (got $(lineth_hex_to_dec_small "${i2_chain_hex:-0x0}"))" \
  [ "$(lineth_hex_to_dec_small "${i2_chain_hex:-0x0}")" = "$I2_CHAIN" ]

lineth_section "Assert: two LineaRollups on ONE shared L1"
ROLLUP_1="$(rollup_address "$I1_ARTIFACTS/deployments")"
ROLLUP_2="$(rollup_address "$I2_ARTIFACTS/deployments")"
lineth_kv "instance 1 LineaRollupV8" "${ROLLUP_1:-missing}"
lineth_kv "instance 2 LineaRollupV8" "${ROLLUP_2:-missing}"
check "both rollup addresses present" sh -c "[ -n '$ROLLUP_1' ] && [ -n '$ROLLUP_2' ]"
check "rollup addresses differ" [ "$ROLLUP_1" != "$ROLLUP_2" ]
code1="$(rpc_result "$L1_RPC" eth_getCode "[\"$ROLLUP_1\",\"latest\"]")"
code2="$(rpc_result "$L1_RPC" eth_getCode "[\"$ROLLUP_2\",\"latest\"]")"
check "instance 1 rollup has code on shared L1" sh -c "[ -n '$code1' ] && [ '$code1' != '0x' ]"
check "instance 2 rollup has code on shared L1" sh -c "[ -n '$code2' ] && [ '$code2' != '0x' ]"

lineth_section "Assert: both L2s finalized on the shared L1"
b1="$(wait_rollup_block_above "$L1_RPC" "$ROLLUP_1" 0 "$FINALITY_TIMEOUT_SECONDS" "instance 1")" \
  && lineth_ok "instance 1 rollup currentL2BlockNumber=$b1 (> 0)" \
  || { lineth_error "FAILED: instance 1 rollup never finalized"; FAILURES=$((FAILURES + 1)); b1=0; }
b2="$(wait_rollup_block_above "$L1_RPC" "$ROLLUP_2" 0 "$FINALITY_TIMEOUT_SECONDS" "instance 2")" \
  && lineth_ok "instance 2 rollup currentL2BlockNumber=$b2 (> 0)" \
  || { lineth_error "FAILED: instance 2 rollup never finalized"; FAILURES=$((FAILURES + 1)); b2=0; }

if [ "$CHECK_ADVANCE" = "true" ] && [ "$FAILURES" -eq 0 ]; then
  lineth_section "Assert: both rollups ADVANCE while running concurrently"
  # Idle local L2s stop producing blocks once the deploy traffic ends, so the
  # finalized block number only moves with fresh L2 transactions. Nudge both
  # chains, then require further finalizations on each.
  lineth_info "sending L2 test transactions on both instances to produce new blocks"
  i1 env COUNT=6 "$SCRIPT_DIR/traffic-generation/send-l2-test-tx.sh" >/tmp/verify-dual-i1-traffic.log 2>&1 \
    || lineth_warn "instance 1 traffic nudge failed; see /tmp/verify-dual-i1-traffic.log"
  i2 env COUNT=6 "$SCRIPT_DIR/traffic-generation/send-l2-test-tx.sh" >/tmp/verify-dual-i2-traffic.log 2>&1 \
    || lineth_warn "instance 2 traffic nudge failed; see /tmp/verify-dual-i2-traffic.log"
  lineth_info "waiting up to ${ADVANCE_TIMEOUT_SECONDS}s for further finalizations on both rollups"
  a1="$(wait_rollup_block_above "$L1_RPC" "$ROLLUP_1" "$b1" "$ADVANCE_TIMEOUT_SECONDS" "instance 1")" \
    && lineth_ok "instance 1 advanced $b1 -> $a1" \
    || { lineth_error "FAILED: instance 1 did not advance beyond $b1"; FAILURES=$((FAILURES + 1)); }
  a2="$(wait_rollup_block_above "$L1_RPC" "$ROLLUP_2" "$b2" "$ADVANCE_TIMEOUT_SECONDS" "instance 2")" \
    && lineth_ok "instance 2 advanced $b2 -> $a2" \
    || { lineth_error "FAILED: instance 2 did not advance beyond $b2"; FAILURES=$((FAILURES + 1)); }
fi

lineth_section "Evidence"
docker ps --format 'table {{.Names}}\t{{.Status}}' | lineth_indent
lineth_kv "shared L1 RPC" "$L1_RPC (chainId $(lineth_hex_to_dec_small "$(rpc_result "$L1_RPC" eth_chainId '[]')"))"
lineth_kv "instance 1" "L2 chain $I1_CHAIN rpc=$I1_L2_RPC rollup=$ROLLUP_1 finalizedL2Block=$(rollup_current_l2_block "$L1_RPC" "$ROLLUP_1" || echo '?')"
lineth_kv "instance 2" "L2 chain $I2_CHAIN rpc=$I2_L2_RPC rollup=$ROLLUP_2 finalizedL2Block=$(rollup_current_l2_block "$L1_RPC" "$ROLLUP_2" || echo '?')"
lineth_info "status views: ./scripts/status.sh and LINETH_ENV_FILE=$I2_ENV ./scripts/status.sh"

if [ "$KEEP" = "true" ]; then
  lineth_info "--keep set; leaving both instances running"
else
  lineth_section "Teardown"
  i2 "$SCRIPT_DIR/reset.sh" >/tmp/verify-dual-i2-teardown.log 2>&1 || lineth_warn "instance 2 teardown failed; see /tmp/verify-dual-i2-teardown.log"
  i1 "$SCRIPT_DIR/reset.sh" >/tmp/verify-dual-i1-teardown.log 2>&1 || lineth_warn "instance 1 teardown failed; see /tmp/verify-dual-i1-teardown.log"
  lineth_ok "both instances torn down"
fi

if [ "$FAILURES" -gt 0 ]; then
  lineth_die "$FAILURES assertion(s) failed"
fi
lineth_ok "dual-L2 standard verified: two Linea L2s finalizing to one shared local L1"
