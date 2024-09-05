package test_utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/test"
	"github.com/consensys/zkevm-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/constraints"
	"math"
	"math/big"
	"os"
	"strings"
	"testing"
)

func printAsHexHint(_ *big.Int, ins, outs []*big.Int) error {
	byts := make([]byte, len(ins))
	for i := range byts {
		if b := ins[i].Uint64(); !ins[i].IsUint64() || b > 255 {
			return errors.New("not a byte")
		} else {
			byts[i] = byte(b)
		}
	}
	for i := range outs {
		outs[i].SetUint64(uint64(i))
	}
	fmt.Print("<PrintVarsAsHex> 0x")
	fmt.Println(StrBlocks(hex.EncodeToString(byts), 64))

	return nil
}

func PrintVarsAsHex(api frontend.API, v []frontend.Variable) {
	if _, err := api.Compiler().NewHint(printAsHexHint, 1, v...); err != nil {
		panic(err)
	}
}

func PrintBlocksAsHex(api frontend.API, blocks ...[32]frontend.Variable) {
	ins := make([]frontend.Variable, len(blocks)*32)
	k := 0
	for i := range blocks {
		for j := range blocks[i] {
			ins[k] = blocks[i][j]
			k++
		}
	}
	if _, err := api.Compiler().NewHint(printAsHexHint, 1, ins...); err != nil {
		panic(err)
	}
}

func PrintBytesAsNum(b []byte) {
	var i big.Int
	i.SetBytes(b)
	fmt.Printf("%s 0x%s\n", i.Text(10), i.Text(16))
}

// StrBlocks injects a space in a string once every n characters
func StrBlocks(s string, n int) string {
	if len(s) < n {
		return s
	}
	var sb strings.Builder
	expectedLen := len(s) + (len(s)+n-1)/n - 1
	sb.Grow(expectedLen)

	sb.Write([]byte(s[:n]))
	for i := n; i < len(s); i += n {
		sb.WriteByte(' ')
		sb.Write([]byte(s[i:min(i+n, len(s))]))
	}

	if sb.Len() != expectedLen { // TODO delete
		panic(fmt.Errorf("expected length %d got %d", expectedLen, sb.Len()))
	}

	return sb.String()
}

// ProgressReporter prints percentages on console out; useful for monitoring long tasks
type ProgressReporter struct {
	n, pct int
}

func NewProgressReporter(n int) *ProgressReporter {
	return &ProgressReporter{
		n:   n,
		pct: -1,
	}
}

func (r *ProgressReporter) Update(i int) {
	if pct := i * 100 / r.n; pct != r.pct {
		r.pct = pct
		fmt.Printf("%d%%\n", pct)
	}
}

var snarkFunctionStore = make(map[uint64]func(frontend.API) []frontend.Variable) // todo make thread safe

type snarkFunctionTestCircuit struct {
	Outs   []frontend.Variable
	funcId uint64 // this workaround is necessary because deepEquals fails on objects with function fields
}

func (c *snarkFunctionTestCircuit) Define(api frontend.API) error {
	outs := snarkFunctionStore[c.funcId](api)
	delete(snarkFunctionStore, c.funcId)

	// todo replace with SliceEquals
	if len(outs) != len(c.Outs) {
		return errors.New("SnarkFunctionTest: unexpected number of output")
	}
	for i := range outs {
		api.AssertIsEqual(outs[i], c.Outs[i])
	}
	return nil
}

func SnarkFunctionTest(f func(frontend.API) []frontend.Variable, outs ...frontend.Variable) func(*testing.T) {

	return func(t *testing.T) {

		c := snarkFunctionTestCircuit{
			Outs: make([]frontend.Variable, len(outs)),
		}
		var b [8]byte
		_, err := rand.Read(b[:])
		assert.NoError(t, err)
		c.funcId = binary.BigEndian.Uint64(b[:])
		snarkFunctionStore[c.funcId] = f

		a := snarkFunctionTestCircuit{
			Outs: outs,
		}
		assert.NoError(t, test.IsSolved(&c, &a, ecc.BLS12_377.ScalarField()))
	}
}

func Range[T constraints.Integer](length int, startingPoints ...T) []T {
	if len(startingPoints) == 0 {
		startingPoints = []T{0}
	}
	res := make([]T, length*len(startingPoints))
	for i := range startingPoints {
		FillRange(res[i*length:(i+1)*length], startingPoints[i])
	}
	return res
}

func FillRange[T constraints.Integer](dst []T, start T) {
	for l := range dst {
		dst[l] = T(l) + start
	}
}

func BlocksToHex(b ...[][32]byte) []string {
	res := make([]string, 0)
	for i := range b {
		for j := range b[i] {
			res = append(res, utils.HexEncodeToString(b[i][j][:]))
		}
	}
	return res
}

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
