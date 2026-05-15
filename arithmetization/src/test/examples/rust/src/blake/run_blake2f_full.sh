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
  # The zkc program result is reported in the final output line.
  output=$(make -f "$makefile" TEST=blake/blake_with_in_bytes.rs IN_BYTES="$in_bytes" 2>&1)
  last_line=$(printf '%s\n' "$output" | sed '/^[[:space:]]*$/d' | tail -n 1)
  if [[ "$last_line" == "Program exited successfully (exit with code 0)" ]]; then
    echo ok
  else
    echo "failed ($last_line)"
    failures+=("$n: $last_line")
  fi
done <<< "$cases"

if ((${#failures[@]})); then
  echo "failed cases: ${failures[*]}"
  exit 1
fi

echo "all cases passed"
