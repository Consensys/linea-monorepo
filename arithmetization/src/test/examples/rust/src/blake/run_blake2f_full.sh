#!/usr/bin/env bash
set -u

cd "$(dirname "$0")/.."

makefile="../../Makefile"
vectors="${1:-blake/blake2f.full}"
[[ -f "$vectors" ]] || vectors="blake/blake2f.all"
[[ -f "$vectors" ]] || { echo "missing vector file: $vectors"; exit 1; }

# Convert .full/.all rows to IN_BYTES values expected by the Rust test.
cases=$(python3 blake/blake2f_to_in_bytes.py "$vectors") || exit 1

failures=()
while IFS=: read -r n in_bytes; do
  printf 'case %s... ' "$n"
  # Run one zkVM test invocation per vector.
  if make -f "$makefile" TEST=blake/blake_with_in_bytes.rs IN_BYTES="$in_bytes" >/dev/null; then
    echo ok
  else
    rc=$?
    echo "failed ($rc)"
    failures+=("$n")
  fi
done <<< "$cases"

if ((${#failures[@]})); then
  echo "failed cases: ${failures[*]}"
  exit 1
fi

echo "all cases passed"
