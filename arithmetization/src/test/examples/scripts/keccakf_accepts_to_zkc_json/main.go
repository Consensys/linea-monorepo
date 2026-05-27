// Strips the `result` field from a keccakf_batched `.accepts` JSON file and
// emits the remaining ROM inputs (n_vectors, block_counts, blocks) to stdout
// as a single JSON object suitable for `zkc exec`.
//
// The .accepts format used by the Go test harness includes `result` as the
// expected output. Passing it through `zkc exec` causes the CLI to error out
// ("input field `result` is unknown"), so we drop it here.
//
// Usage:
//
//	keccakf_accepts_to_zkc_json <input.accepts>
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: keccakf_accepts_to_zkc_json <input.accepts>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: read:", err)
		os.Exit(1)
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		fmt.Fprintln(os.Stderr, "error: parse:", err)
		os.Exit(1)
	}

	delete(obj, "result")

	out, err := json.Marshal(obj)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: marshal:", err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(out); err != nil {
		fmt.Fprintln(os.Stderr, "error: write:", err)
		os.Exit(1)
	}
}
