package gkrverifier

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// zkcField is the field the ZkC programs of this package are written for.
const zkcField = "KOALABEAR_16"

// zkcInput is a ZkC input array: the words of an array of u32, in address order.
type zkcInput []uint32

func (in zkcInput) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteString("0x")
	for i, w := range in {
		if i != 0 {
			sb.WriteByte('_')
		}
		sb.WriteString(strings.ToLower(hex8(w)))
	}
	return json.Marshal(sb.String())
}

func hex8(w uint32) string {
	const digits = "0123456789abcdef"
	var b [8]byte
	for i := 7; i >= 0; i-- {
		b[i] = digits[w&0xf]
		w >>= 4
	}
	return string(b[:])
}

// runZkc executes the ZkC program main, with the given library files, on
// inputs. It reports whether the program ran to completion; a
// program that hits `fail` yields false, while a compilation error
// or any other failure to run aborts the test.
func runZkc(t *testing.T, inputs map[string]zkcInput, main string, libs ...string) bool {
	t.Helper()

	bz, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	inputFile := filepath.Join(t.TempDir(), "input.json")
	if err = os.WriteFile(inputFile, bz, 0o600); err != nil {
		t.Fatal(err)
	}

	args := append([]string{"tool", "zkc", "execute", "--field", zkcField, inputFile, main}, libs...)
	out, err := exec.Command("go", args...).CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "machine panic") {
			t.Fatalf("running zkc: %v\n%s", err, out)
		}
		return false
	}
	return true
}
