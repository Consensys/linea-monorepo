package test_utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"os"
	"strconv"
	"strings"
)

type FakeTestingT struct{}

func (FakeTestingT) Errorf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format+"\n", args...))
}

func (FakeTestingT) FailNow() {
	os.Exit(-1)
}

func RandIntN(n int) int {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	if n > math.MaxInt {
		panic("RandIntN: n too large")
	}
	return int(binary.BigEndian.Uint64(b[:]) % uint64(n)) // #nosec G115 -- Above check precludes an overflow
}

func RandIntSliceN(length, n int) []int {
	res := make([]int, length)
	for i := range res {
		res[i] = RandIntN(n)
	}
	return res
}

// AssertBytesEqual between byte slices a,b
// a readable error message would show in case of inequality
// TODO error options: block size, check forwards or backwards etc
func AssertBytesEqual(t assert.TestingT, expected, actual []byte) {
	l := min(len(expected), len(actual))

	failure := 0
	for failure < l {
		if expected[failure] != actual[failure] {
			break
		}
		failure++
	}

	if len(expected) == len(actual) {
		return
	}

	// there is a mismatch
	var sb strings.Builder

	const (
		radius    = 40
		blockSize = 32
	)

	printCentered := func(b []byte) {

		for i := max(failure-radius, 0); i <= failure+radius; i++ {
			if i%blockSize == 0 && i != failure-radius {
				sb.WriteString("  ")
			}
			if i >= 0 && i < len(b) {
				sb.WriteString(hex.EncodeToString([]byte{b[i]})) // inefficient, but this whole error printing sub-procedure will not be run more than once
			} else {
				sb.WriteString("  ")
			}
		}
	}

	sb.WriteString("expected: ")
	printCentered(expected)
	sb.WriteString("\n")

	sb.WriteString("actual:   ")
	printCentered(actual)
	sb.WriteString("\n")

	sb.WriteString("          ")
	for i := max(failure-radius, 0); i <= failure+radius; {
		if i%blockSize == 0 && i != failure-radius {
			s := strconv.Itoa(i)
			sb.WriteString("  ")
			sb.WriteString(s)
			i += len(s) / 2
			if len(s)%2 != 0 {
				sb.WriteString(" ")
				i++
			}
		} else {
			if i == failure {
				sb.WriteString("^^")
			} else {
				sb.WriteString("  ")
			}
			i++
		}
	}

	t.Errorf("mismatch starting at byte %d\n%s\n", failure, sb.String())
}
