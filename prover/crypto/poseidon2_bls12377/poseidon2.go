package poseidon2_bls12377

import (
	"errors"
	"fmt"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/encoding"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const BlockSize = 1

// TODO @thomas fixme arbitrary max size of the buffer
const maxSizeBuf = 1024

// MDHasher Merkle Damgard Hasher using Poseidon2 as compression function
type MDHasher struct {
	maxValue types.Bytes32 // the maximal value obtainable with that hasher

	// Sponge construction Sstate
	Sstate fr.Element

	// data to hash
	Buffer []fr.Element

	verbose bool
}

var (
	ErrInvalidSizebuffer = errors.New("the size of the input should match the size of the hash buffer")
	compressor           poseidon2.Permutation
	once                 sync.Once
)

func init() {
	once.Do(func() {
		// default parameters
		compressor = *poseidon2.NewPermutation(2, 6, 26)
	})
}

// Constructor for Poseidon2MDHasher
func NewMDHasher(verbose ...bool) *MDHasher {
	// var maxVal fr.Element
	// maxVal.SetOne().Neg(&maxVal)
	return &MDHasher{
		Sstate:  fr.Element{},
		Buffer:  make([]fr.Element, 0),
		verbose: len(verbose) > 0 && verbose[0],
	}
}

// Reset clears the buffer, and reset state to iv
func (d *MDHasher) Reset() {
	d.Buffer = d.Buffer[:0]
	d.Sstate = fr.Element{}
}

func (d *MDHasher) SetStateFrElement(s fr.Element) {
	d.Reset()
	d.Sstate.Set(&s)
}

// State returns the state of the hasher. The function must not mutate the
// MDHasher. If it needs to flush the buffer to access the state, it will clone
// the struct and flush there.
func (d *MDHasher) State() fr.Element {

	// If the buffer is clean, we can short-path the execution and directly
	// return the state.
	if len(d.Buffer) == 0 {
		return d.Sstate
	}

	// If the buffer is not clean, we cannot clean it locally as it would modify
	// the state of the hasher locally. Instead, we clone the buffer and flush
	// the buffer on the clone.
	clone := NewMDHasher()
	clone.Buffer = make([]fr.Element, len(d.Buffer))
	copy(clone.Buffer, d.Buffer)
	clone.Sstate = d.Sstate
	_ = clone.SumElement()
	return clone.Sstate
}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) WriteKoalabearElements(elmts ...field.Element) {
	_elmts := encoding.EncodeKoalabearsToFrElement(elmts)
	var ele fr.Element
	ele.SetString("44800556421982794402844311666878798555334702622203491398871381848572998469")
	if _elmts[0] == ele {
		fmt.Printf("[WriteKoalabearElements] elements %v, \n", _elmts[0].String())

		if len(d.Buffer) < 2 {
			fmt.Printf("[WriteKoalabearElements] blsbuffer = %v\n", fr.Vector(d.Buffer[:]).String())
		} else {
			fmt.Printf("[WriteKoalabearElements] blsbuffer = %v\n", fr.Vector(d.Buffer[:2]).String())
		}

	}

	d.WriteElements(_elmts...)
	if _elmts[0] == ele {

		if len(d.Buffer) < 2 {
			fmt.Printf("[WriteKoalabearElements] blsbuffer after= %v\n", fr.Vector(d.Buffer[:]).String())
		} else {
			fmt.Printf("[WriteKoalabearElements] blsbuffer after= %v\n", fr.Vector(d.Buffer[:2]).String())
		}
	}

}

// WriteElements adds a slice of field elements to the running hash.
func (d *MDHasher) WriteElements(elmts ...fr.Element) {
	var ele, ele2 fr.Element
	ele.SetString("44800556421982794402844311666878798555334702622203491398871381848572998469")
	ele2.SetString("296415441165787297985648523858374568612462956832226556151653813342063403288")
	quo := (len(d.Buffer) + len(elmts)) / maxSizeBuf
	rem := (len(d.Buffer) + len(elmts)) % maxSizeBuf
	off := len(d.Buffer)
	if elmts[0] == ele || elmts[0] == ele2 {
		fmt.Printf("quo %v, \n", quo)
		fmt.Printf("rem %v, \n", rem)
		fmt.Printf("off %v, \n", off)
		fmt.Printf("len(d.Buffer) %v, \n", len(d.Buffer))
		fmt.Printf("len(elmts) %v, \n", len(elmts))

	}

	for i := 0; i < quo; i++ {
		d.Buffer = append(d.Buffer, elmts[:maxSizeBuf-off]...)
		fmt.Printf("d.Buffer %v, \n", fr.Vector(d.Buffer[:2]).String())
		fmt.Print("Before SumElement\n")

		sum := d.SumElement()
		fmt.Printf("After SumElement %v, \n", sum.String())
		d.Buffer = d.Buffer[:0] // flush the buffer once maxSizeBuf is reached
		off = len(d.Buffer)
		elmts = elmts[maxSizeBuf-off:] // Update data to the remaining part
	}
	d.Buffer = append(d.Buffer, elmts[:rem-off]...)
}

func compress(left, right fr.Element) fr.Element {
	var x [2]fr.Element
	x[0].Set(&left)
	x[1].Set(&right)
	res := x[1] // save right to feed forward later
	compressor.Permutation(x[:])
	res.Add(&res, &x[1])
	return res
}

func (d *MDHasher) SumElement() fr.Element {
	if d.verbose {
		fmt.Printf("[native fs flush] oldState %v, buffer = %v\n", d.Sstate.String(), fr.Vector(d.Buffer[:1]).String())
	}
	for i := 0; i < len(d.Buffer); i++ {
		d.Sstate = compress(d.Sstate, d.Buffer[i])
	}
	d.Buffer = d.Buffer[:0]
	return d.Sstate
}
