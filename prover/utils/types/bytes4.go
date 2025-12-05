package types

import (
	"fmt"
	"io"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Bytes4 represents an arbitrary bytes string of 4 bytes that must be smaller
// than the Koalabear field modulus (2^31 - 2^24 + 1 = 2013265921). It is commonly
// used to represent values that need to fit in a single Koalabear field element.
type Bytes4 [4]byte

// Marshal "e" into JSON format
func (f Bytes4) MarshalJSON() ([]byte, error) {
	marshalled := MarshalHexBytesJSON(f[:])
	return marshalled, nil
}

// Unmarshal a Bytes4 from JSON format. The expected format is an hex string.
// The value must be smaller than the Koalabear modulus.
func (f *Bytes4) UnmarshalJSON(b []byte) error {

	decoded, err := DecodeQuotedHexString(b)
	if err != nil {
		return fmt.Errorf(
			"could not decode bytes4 `%v`, expected an hex string of 4 bytes : %w",
			string(b), err,
		)
	}

	if len(decoded) != 4 {
		return fmt.Errorf(
			"could not unmarshal bytes4 %x : should have 4 bytes but has %v bytes",
			decoded, len(decoded),
		)
	}

	copy((*f)[:], decoded)

	// Validate that the value is smaller than the Koalabear modulus
	var fe field.Element
	if err := fe.SetBytesCanonical((*f)[:]); err != nil {
		return fmt.Errorf(
			"could not unmarshal bytes4 %x : value must be smaller than Koalabear modulus: %w",
			decoded, err,
		)
	}

	return nil
}

// Writes the bytes4 into the given writer.
func (b Bytes4) WriteTo(w io.Writer) (int64, error) {
	_, err := w.Write(b[:])
	if err != nil {
		panic(err) // hard forbid any error
	}
	return 4, nil
}

// Reads a bytes4 from the given reader
func (b *Bytes4) ReadFrom(r io.Reader) (int64, error) {
	n, err := r.Read((*b)[:])
	return int64(n), err
}

/*
Cmp two Bytes4s. The Bytes4 are interpreted as big-endians big integers and
then are compared. Returns:
  - a < b : -1
  - a == b : 0
  - a > b : 1
*/
func Bytes4Cmp(a, b Bytes4) int {
	var bigA, bigB big.Int
	bigA.SetBytes(a[:])
	bigB.SetBytes(b[:])
	return bigA.Cmp(&bigB)
}

// Returns an hexstring representation of the Bytes4
func (d Bytes4) Hex() string {
	return fmt.Sprintf("0x%x", [4]byte(d))
}

// Constructs a dummy Bytes4 from an integer.
// Panics if i >= Koalabear modulus.
func DummyBytes4(i int) (d Bytes4) {
	d[0] = byte(i >> 24)
	d[1] = byte(i >> 16)
	d[2] = byte(i >> 8)
	d[3] = byte(i)
	// Validate that the value is canonical
	var f field.Element
	if err := f.SetBytesCanonical(d[:]); err != nil {
		utils.Panic("DummyBytes4: value %v exceeds Koalabear modulus", i)
	}
	return d
}

// LeftPadToBytes4 pads a bytes slice into a Bytes4 by adding zeroes to
// the left until the slice has 4 bytes.
// Panics if the resulting value >= Koalabear modulus.
func LeftPadToBytes4(b []byte) Bytes4 {
	if len(b) > 4 {
		utils.Panic("Passed a string of %v element but the max is 4", len(b))
	}
	c := append(make([]byte, 4-len(b)), b...)
	return AsBytes4(c)
}

// Create a bytes4 from a slice.
// Panics if the value >= Koalabear modulus.
func AsBytes4(b []byte) (d Bytes4) {
	// Sanity-check the length of the digest
	if len(b) != len(Bytes4{}) {
		utils.Panic("Passed a string of %v bytes but expected %v", len(b), 4)
	}
	copy(d[:], b)
	// Validate that the value is smaller than the Koalabear modulus
	var f field.Element
	if err := f.SetBytesCanonical(d[:]); err != nil {
		utils.Panic("AsBytes4: value 0x%x exceeds Koalabear modulus", b)
	}
	return d
}

// Creates a bytes4 from an hexstring. Panic if it fails. Mostly useful for testing.
// the string s is left padded with zeroes if less than 4 characters are provided
// if more than 4 characters are provided, the function will panic
// function expects an even number of chars
// 0x prefix is optional
// Panics if the value >= Koalabear modulus.
func Bytes4FromHex(s string) Bytes4 {
	b, err := utils.HexDecodeString(s)
	if err != nil {
		utils.Panic("not an hexadecimal %v", s)
	}
	if len(b) > 4 {
		utils.Panic("String passed should have even length <= 4 bytes")
	}

	var res Bytes4
	copy(res[4-len(b):], b)

	var f field.Element
	if err := f.SetBytesCanonical(res[:]); err != nil {
		utils.Panic("Bytes4FromHex: value %v exceeds Koalabear modulus", s)
	}
	return res
}

// ToFieldElement returns a field.Element from the Bytes4.
// Panics if the value >= Koalabear modulus.
func (d Bytes4) ToFieldElement() field.Element {
	var f field.Element
	if err := f.SetBytesCanonical(d[:]); err != nil {
		utils.Panic("ToFieldElement: value 0x%x exceeds Koalabear modulus", d[:])
	}
	return f
}
