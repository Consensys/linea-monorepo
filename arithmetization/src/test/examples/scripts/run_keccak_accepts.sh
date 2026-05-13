#!/usr/bin/env bash
#
# run_keccak_accepts.sh — Run the keccak/keccak.rs RISC-V program once per
# test vector in a JSON-Lines .accepts file (as produced by go-corset's
# zkc bench fixtures).
#
# Strategy (variant B): compile the ELF once, then for each vector
# regenerate the ZKC JSON with the right IN_BYTES and invoke `zkc exec`.
# This avoids 1000 redundant cargo/Make invocations while preserving the
# exact same runtime semantics as `make exec`.
#
# IN_BYTES layout for each vector (see rust/src/keccak/keccak.rs):
#   - 680 bytes : left-padded big-endian 5440-bit message
#   -   8 bytes : msg_len_bits as little-endian u64
#   -  32 bytes : expected Keccak-256 digest
# Total = 720 bytes (1440 hex chars), prefixed with "0x".
#
# Usage:
#   ./run_keccak_accepts.sh [<accepts-file>]
#
# Default accepts-file:
#   $HOME/go/src/github/Consensys/go-corset/testdata/zkc/bench/keccakf_with_padding.accepts
#
# Outputs (per invocation, under scripts/output-<timestamp>/):
#   compile.log         — output of the one-shot `make compile`
#   run_<NNNN>.json     — ZKC input JSON for vector NNNN (preserved)
#   run_<NNNN>.log      — diagnostics for vector NNNN (only on failure;
#                         `zkc exec` output is otherwise discarded because
#                         it is ~16 MB per run)
#   failures.txt        — one line per failed vector ("<NNNN> exit=<rc>")
#   summary.txt         — totals at the end of the run
#
# To replay any failing vector with full output:
#   zkc exec <out>/run_<NNNN>.json <repo>/src/main/riscv/main.zkc

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EX_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ACCEPTS_DEFAULT="$HOME/go/src/github/Consensys/go-corset/testdata/zkc/bench/keccakf_with_padding.accepts"
ACCEPTS="${1:-$ACCEPTS_DEFAULT}"

if [[ ! -f "$ACCEPTS" ]]; then
  echo "ERROR: accepts file not found: $ACCEPTS" >&2
  exit 1
fi

for tool in jq python3 zkc go make; do
  command -v "$tool" >/dev/null || { echo "ERROR: missing tool: $tool" >&2; exit 1; }
done

STAMP="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="$SCRIPT_DIR/output-$STAMP"
mkdir -p "$OUT_DIR"
echo "Output dir: $OUT_DIR"

# Resolve paths the same way the Makefile does, so we stay in sync with it.
ELF="$EX_DIR/rust/target/riscv64im-unknown-none-elf/release/keccak"
ZKC_MAIN="$EX_DIR/../../main/riscv/main.zkc"
MAIN_GO="$EX_DIR/main.go"
MAIN_BIN="$OUT_DIR/elf2zkcjson"

if [[ ! -f "$ZKC_MAIN" ]]; then
  echo "ERROR: ZKC main not found at $ZKC_MAIN" >&2
  exit 1
fi

# Pre-build the Go helper that turns (ELF + IN_BYTES) into the ZKC JSON,
# so we don't pay `go run` startup on every iteration.
echo ">>> Pre-building the ELF→JSON helper..."
go build -o "$MAIN_BIN" "$MAIN_GO"

# 1) Build the ELF once. IN_BYTES does not influence the binary, so any
#    valid placeholder works. We go through `make compile` so the linker
#    script + ELF + (initial, throwaway) JSON are produced in the
#    standard way.
echo ">>> Compiling once (this may take a moment on a cold build)..."
( cd "$EX_DIR" && make compile TEST=keccak/keccak.rs IN_BYTES=0x00 ) \
  >"$OUT_DIR/compile.log" 2>&1

if [[ ! -f "$ELF" ]]; then
  echo "ERROR: ELF not built at $ELF (see $OUT_DIR/compile.log)" >&2
  exit 1
fi

# 2) Iterate over every vector serially.
FAILURES="$OUT_DIR/failures.txt"
SUMMARY="$OUT_DIR/summary.txt"
: >"$FAILURES"

TOTAL="$(wc -l <"$ACCEPTS")"
echo ">>> Running $TOTAL vectors serially..."

PASS=0
FAIL=0
i=0

# `|| [[ -n "$line" ]]` handles a possibly missing trailing newline.
while IFS= read -r line || [[ -n "$line" ]]; do
  i=$((i + 1))
  ID="$(printf '%04d' "$i")"
  JSON="$OUT_DIR/run_$ID.json"
  LOG="$OUT_DIR/run_$ID.log"

  # Build IN_BYTES = "0x" + message + msg_len(LE) + result.
  # We do this inside python3 to keep the byte-reversal robust.
  IN_BYTES="$(printf '%s' "$line" | python3 -c '
import json, sys
o = json.loads(sys.stdin.read())
msg = o["message"][2:]
le  = bytes.fromhex(o["message_length"][2:])[::-1].hex()
res = o["result"][2:]
assert len(msg) == 1360, f"unexpected message hex length: {len(msg)}"
assert len(le)  == 16,   f"unexpected length-field hex length: {len(le)}"
assert len(res) == 64,   f"unexpected result hex length: {len(res)}"
sys.stdout.write("0x" + msg + le + res)
')" || { echo "ERROR: failed to build IN_BYTES for line $i" >&2; exit 1; }

  # Regenerate the ZKC JSON for this vector. We capture stderr to LOG
  # (cheap) so any failure is diagnosable.
  set +e
  "$MAIN_BIN" "$ELF" "$IN_BYTES" 0x00000000 0x08800000 0x00000000 \
    >"$JSON" 2>"$LOG"
  rc=$?
  set -e
  if [[ "$rc" -ne 0 ]]; then
    FAIL=$((FAIL + 1))
    echo "$ID exit=$rc (json-gen failed)" >>"$FAILURES"
    printf '[%s/%s] FAIL (json-gen) exit=%d (see %s)\n' "$ID" "$TOTAL" "$rc" "$LOG"
    continue
  fi

  # Run `zkc exec` on the freshly generated JSON. Its output is ~16 MB,
  # so discard it; we only care about the exit code. On failure leave a
  # tiny breadcrumb in LOG; full output is reproducible via:
  #   zkc exec "$JSON" "$ZKC_MAIN"
  set +e
  zkc exec "$JSON" "$ZKC_MAIN" >/dev/null 2>&1
  rc=$?
  set -e
  if [[ "$rc" -eq 0 ]]; then
    PASS=$((PASS + 1))
    # JSON-gen succeeded with empty stderr; discard the empty LOG to
    # keep the output directory tidy.
    [[ -s "$LOG" ]] || rm -f "$LOG"
    printf '[%s/%s] PASS\n' "$ID" "$TOTAL"
  else
    FAIL=$((FAIL + 1))
    {
      echo "zkc exec failed (rc=$rc)"
      echo "replay with: zkc exec \"$JSON\" \"$ZKC_MAIN\""
    } >>"$LOG"
    echo "$ID exit=$rc" >>"$FAILURES"
    printf '[%s/%s] FAIL exit=%d (see %s)\n' "$ID" "$TOTAL" "$rc" "$LOG"
  fi
done <"$ACCEPTS"

{
  echo "total=$TOTAL"
  echo "pass=$PASS"
  echo "fail=$FAIL"
  echo "out=$OUT_DIR"
} | tee "$SUMMARY"
