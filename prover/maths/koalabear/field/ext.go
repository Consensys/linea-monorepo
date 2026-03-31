package field

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"math/rand/v2"
	"reflect"
	"runtime"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

const ExtensionDegree int = 4

// Embedding
type Ext = extensions.E4

// NewExtFromString only sets the first coordinate of the field extension
func NewExtFromString(s string) (res Ext) {
	res.B0.A0 = NewFromString(s)
	return res
}

// var RootPowers = []int{1, 3}, v^2=u and u^2=3.
var RootPowers = []int{1, 3}

// BatchInvertExt compute the inverses of all elements in the provided slice
// using the Montgommery trick. Zeroes are ignored.
func BatchInvertExt(a []Ext) []Ext {
	return extensions.BatchInvertE4(a)
}

// BatchInvertInto computes the inverses of all elements in a and writes the result into res.
// use with caution, avoid copies and allocation in parallel contexts.
// TODO @gbotrel move in gnark-crypto
func BatchInvertExtInto(a, res []Ext) {
	if len(a) != len(res) {
		// TODO @gbotrel add check that a != res
		panic("input and output slices must have the same length")
	}
	if len(a) == 0 {
		return
	}

	zeroes := make([]bool, len(a))
	var accumulator Ext
	accumulator.SetOne()

	for i := 0; i < len(a); i++ {
		if a[i].IsZero() {
			zeroes[i] = true
			continue
		}
		res[i].Set(&accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	accumulator.Inverse(&accumulator)

	for i := len(a) - 1; i >= 0; i-- {
		if zeroes[i] {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}
}

// PseudoRandExt returns a random field extension element.
func PseudoRandExt(rng *rand.Rand) Ext {
	result := new(Ext).SetZero()
	result.B0.A0 = PseudoRand(rng)
	result.B0.A1 = PseudoRand(rng)
	result.B1.A0 = PseudoRand(rng)
	result.B1.A1 = PseudoRand(rng)
	return *result
}

// IsBase checks if the field extensionElement is a base element.  An Element is
// considered a base element if all coordinates are 0 except for the first one.
func IsBase(z *Ext) bool {
	return z.B0.A1[0] == 0 && z.B1.A0[0] == 0 && z.B1.A1[0] == 0
}

// GetBase attempts to unlift a field extension element into a base field
// element and returns a boolean indicating success.
func GetBase(z *Ext) (Element, bool) {
	if IsBase(z) {
		return z.B0.A0, true
	}
	return Element{}, false
}

// AddByBase implements mixed addition of a base field element.
func AddByBase(z *Ext, first *Ext, second *Element) *Ext {
	z.Set(first)
	z.B0.A0.Add(&z.B0.A0, second)
	return z
}

// MulByBase implements element-wise multiplication by a base field scalar.
func MulByBase(z *Ext, first *Ext, second *Element) *Ext {
	z.B0.A0.Mul(&first.B0.A0, second)
	z.B0.A1.Mul(&first.B0.A1, second)
	z.B1.A0.Mul(&first.B1.A0, second)
	z.B1.A1.Mul(&first.B1.A1, second)
	return z
}

// DivByBase implements a division by a base field element.
func DivByBase(z *Ext, first *Ext, second *Element) *Ext {
	z.B0.A0.Div(&first.B0.A0, second)
	z.B0.A1.Div(&first.B0.A1, second)
	z.B1.A0.Div(&first.B1.A0, second)
	z.B1.A1.Div(&first.B1.A1, second)
	return z
}

// SetInterface is a generic purpose function to set a field element from an
// interface{}
func SetInterface(z *Ext, i1 interface{}) (*Ext, error) {
	if i1 == nil {
		return nil, errors.New("can't set fr.Element with <nil>")
	}

	switch c1 := i1.(type) {
	case Ext:
		return z.Set(&c1), nil
	case *Ext:
		if c1 == nil {
			return nil, errors.New("can't set fext.Element with <nil>")
		}
		return z.Set(c1), nil
	case Element:
		return SetExtFromBase(z, &c1), nil
	case *Element:
		if c1 == nil {
			return nil, errors.New("can't set fext.Element with <nil>")
		}
		return SetExtFromBase(z, c1), nil
	case uint8:
		return SetExtFromUInt(z, uint64(c1)), nil
	case uint16:
		return SetExtFromUInt(z, uint64(c1)), nil
	case uint32:
		return SetExtFromUInt(z, uint64(c1)), nil
	case uint:
		return SetExtFromUInt(z, uint64(c1)), nil
	case uint64:
		return SetExtFromUInt(z, c1), nil
	case int8:
		return SetExtFromInt(z, int64(c1)), nil
	case int16:
		return SetExtFromInt(z, int64(c1)), nil
	case int32:
		return SetExtFromInt(z, int64(c1)), nil
	case int64:
		return SetExtFromInt(z, c1), nil
	case int:
		return SetExtFromInt(z, int64(c1)), nil
	case string:
		z.B0.A0.SetString(c1)
		z.B0.A1.SetZero()
		z.B1.SetZero()
		return z, nil
	case *big.Int:
		if c1 == nil {
			return nil, errors.New("can't set fr.Element with <nil>")
		}
		z.B0.A0.SetBigInt(c1)
		return z, nil
	case big.Int:
		z.B0.A0.SetBigInt(&c1)
		return z, nil
	case []byte:
		z := BytesToExt(c1)
		return &z, nil
	default:
		return nil, errors.New("can't set fext.Element from type " + reflect.TypeOf(i1).String())
	}
}

func ExtToText(z *Ext, base int) string {
	if base < 2 || base > 36 {
		panic("invalid base")
	}
	if z == nil {
		return "<nil>"
	}

	res := fmt.Sprintf("%s + %s*u + (%s + %s*u)*v", z.B0.A0.Text(base), z.B0.A1.Text(base), z.B1.A0.Text(base), z.B1.A1.Text(base))
	return res
}

func ParBatchInvertExt(a []Ext, numCPU int) []Ext {

	if numCPU == 0 {
		numCPU = runtime.GOMAXPROCS(0)
	}

	res := make([]Ext, len(a))
	parallel.Execute(len(a), func(start, stop int) {
		BatchInvertExtInto(a[start:stop], res[start:stop])
	}, numCPU)

	return res
}

// MulRInv multiplies the field element by R^-1, where R is the Montgommery constant
func MulRInvExt(x Ext) Ext {
	var res Ext
	res.MulByElement(&x, &MontConstantInv)
	return res
}

// One returns 1
func OneExt() Ext {
	var res Ext
	res.B0.A0.SetOne()
	return res
}

// SetZero sets an E4 elmt to zero
func ZeroExt() Ext {
	var res Ext
	return res
}

// ExtToUint64s returns the value of z as a tuple of 4 uint64.
func ExtToUint64s(z *Ext) (uint64, uint64, uint64, uint64) {
	return uint64(z.B0.A0.Bits()[0]),
		uint64(z.B0.A1.Bits()[0]),
		uint64(z.B1.A0.Bits()[0]),
		uint64(z.B1.A1.Bits()[0])
}

// SetExtFromUInt sets z to v and returns z. After conversion, z is in the base
// field and can safely be converted into an [Element].
func SetExtFromUInt(z *Ext, v uint64) *Ext {
	//  sets z LSB to v (non-Montgomery form) and convert z to Montgomery form
	z.B0.A0.SetUint64(v)
	return z // z.toMont()
}

// SetExtFromInt sets z to v and returns z. After conversion, z is in the base
// field and can safely be converted into an [Element].
func SetExtFromInt(z *Ext, v int64) *Ext {
	z.B0.A0.SetInt64(v)
	return z // z.toMont()
}

// SetExtFromBase sets z to v and returns z. After conversion, z is in the base
// field and can safely be converted into an [Element].
func SetExtFromBase(z *Ext, x *Element) *Ext {
	z.B0.A0[0] = x[0]
	z.B0.A1[0] = 0
	z.B1.A0[0] = 0
	z.B1.A1[0] = 0
	return z
}

// Uint64ToExt constructs a field extension from a uint64. The returned element
// is convertible into a base field element.
func Uint64ToExt(b uint64) Ext {
	var res Ext
	res.B0.A0.SetUint64(b)
	return res
}

// UintsToExt constructs a field extension from 4 uint64
func UintsToExt(v1, v2, v3, v4 uint64) Ext {
	var z Ext
	z.B0.A0.SetUint64(v1)
	z.B0.A1.SetUint64(v2)
	z.B1.A0.SetUint64(v3)
	z.B1.A1.SetUint64(v4)
	return z
}

// IntsToExt constructs a field extension from 4 int64
func IntsToExt(v1, v2, v3, v4 int64) Ext {
	var z Ext
	z.B0.A0.SetInt64(v1)
	z.B0.A1.SetInt64(v2)
	z.B1.A0.SetInt64(v3)
	z.B1.A1.SetInt64(v4)
	return z
}

func Lift(v Element) Ext {
	var res Ext
	res.B0.A0.Set(&v)
	return res
}

// PseudoRand generates a field using a pseudo-random number generator
func RandomElementExt() Ext {
	var res Ext
	res.SetRandom()
	return res
}

func ExpByIntExt(z *Ext, x Ext, k int) *Ext {
	if k == 0 {
		return z.SetOne()
	}

	if k < 0 {
		x.Inverse(&x)
		k = -k
	}

	z.Set(&x)

	for i := bits.Len(uint(k)) - 2; i >= 0; i-- {
		z.Square(z)
		if (k>>i)&1 == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}

// Bytes returns the value of z as a big-endian byte array
// TODO: check if this way is correct
// the output is 16 bytes, not Bytes32
func ExtToBytes(z *Ext) (res [Bytes * 4]byte) {
	var result [Bytes * 4]byte

	valBytes := z.B0.A0.Bytes()
	copy(result[0:Bytes], valBytes[:])

	valBytes = z.B0.A1.Bytes()
	copy(result[Bytes:2*Bytes], valBytes[:])

	valBytes = z.B1.A0.Bytes()
	copy(result[2*Bytes:3*Bytes], valBytes[:])

	valBytes = z.B1.A1.Bytes()
	copy(result[3*Bytes:4*Bytes], valBytes[:])

	return result
}

func BytesToExt(data []byte) Ext {
	var res Ext
	res.B0.A0 = koalabear.Element{binary.BigEndian.Uint32(data[0:4])}
	res.B0.A1 = koalabear.Element{binary.BigEndian.Uint32(data[4:8])}
	res.B1.A0 = koalabear.Element{binary.BigEndian.Uint32(data[8:12])}
	res.B1.A1 = koalabear.Element{binary.BigEndian.Uint32(data[12:16])}
	return res
}
