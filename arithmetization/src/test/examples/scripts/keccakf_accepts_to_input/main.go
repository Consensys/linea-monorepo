// Converts a native-zkc keccakf `.accepts` fixture into the input JSON that
// `zkc exec` expects.
//
// The `.accepts` fixture is a single JSON object with four hex-string fields:
//   - "n_vectors"     : ROM input (number of independent hashes)
//   - "block_counts"  : ROM input (per-vector block count)
//   - "blocks"        : ROM input (flat concatenation of all input blocks)
//   - "result"        : WOM output (expected per-vector 256-bit digest)
//
// `zkc exec` accepts only ROM inputs; passing the WOM `result` field along
// with the inputs makes it fail. This helper strips `result` and emits a
// compact JSON object with the three input fields in canonical order.
//
// Usage:
//
//	keccakf_accepts_to_input <in.accepts> <out.json>
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: keccakf_accepts_to_input <in.accepts> <out.json>")
		os.Exit(1)
	}
	inPath, outPath := os.Args[1], os.Args[2]

	data, err := os.ReadFile(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read %s: %v\n", inPath, err)
		os.Exit(1)
	}

	// json.RawMessage avoids re-parsing the giant hex strings: we just shuffle
	// the (already-validated) JSON tokens into the output buffer untouched.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "error: parse %s: %v\n", inPath, err)
		os.Exit(1)
	}

	inputKeys := []string{"n_vectors", "block_counts", "blocks"}
	for _, k := range inputKeys {
		if _, ok := raw[k]; !ok {
			fmt.Fprintf(os.Stderr, "error: missing required ROM input key %q in %s\n", k, inPath)
			os.Exit(1)
		}
	}

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range inputKeys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kj, _ := json.Marshal(k)
		buf.Write(kj)
		buf.WriteByte(':')
		buf.Write(raw[k])
	}
	buf.WriteByte('}')

	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: write %s: %v\n", outPath, err)
		os.Exit(1)
	}

	hadResult := ""
	if _, ok := raw["result"]; ok {
		hadResult = " (dropped WOM output field `result`)"
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes) with keys %v%s\n",
		outPath, buf.Len(), inputKeys, hadResult)
}
