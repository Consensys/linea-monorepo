#!/usr/bin/env bash
cd "$(dirname "$0")/.."

makefile="../../Makefile"
vectors="blake/blake2f.all"
selector="${1:-}"
[[ -f "$vectors" ]] || { echo "missing vector file: $vectors"; exit 1; }

# Convert .all rows to IN_BYTES values expected by the Rust test
cases=$(python3 blake/blake2f_all_to_in_bytes.py "$vectors") || exit 1
if [[ -n "$selector" ]]; then
  if [[ "$selector" =~ ^([0-9]+)-([0-9]+)$ ]]; then
    cases=$(printf '%s\n' "$cases" | awk -F: -v start="${BASH_REMATCH[1]}" -v end="${BASH_REMATCH[2]}" '$1 >= start && $1 <= end')
  elif [[ "$selector" =~ ^[0-9]+$ ]]; then
    cases=$(printf '%s\n' "$cases" | awk -F: -v n="$selector" '$1 == n')
  else
    echo "selector must be N or START-END"
    exit 1
  fi
  [[ -n "$cases" ]] || { echo "no cases matched: $selector"; exit 1; }
fi

failures=()
while IFS=: read -r n in_bytes; do
  printf 'case %s... ' "$n"
  # The zkc program result is reported in the second last output line
  output=$(make -f "$makefile" TEST=blake/blake_with_in_bytes.rs IN_BYTES="$in_bytes" 2>&1)
  result_line=$(printf '%s\n' "$output" | sed '/^[[:space:]]*$/d' | tail -n 2 | head -n 1)
  if [[ "$result_line" == "Program exited successfully (exit with code 0)." ]]; then
    echo ok
  else
    echo "failed ($result_line)"
    failures+=("$n: $result_line")
  fi
done <<< "$cases"

if ((${#failures[@]})); then
  echo "failed cases: ${failures[*]}"
  exit 1
fi

echo "all cases passed"
